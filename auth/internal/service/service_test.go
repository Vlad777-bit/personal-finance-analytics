package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/domain"
	repositorymocks "github.com/Vlad777-bit/personal-finance-analytics/auth/internal/repository/mocks"
	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/service"
)

type fakePasswordHasher struct {
	hash    func(context.Context, string) (string, error)
	matches func(context.Context, string, string) (bool, error)
}

type fakeTokenIssuer struct {
	issue         func(context.Context, domain.User) (service.AccessToken, error)
	issueRefresh  func(context.Context, domain.User) (service.RefreshToken, error)
	refresh       func(context.Context, string) (service.AccessToken, error)
	refreshTokens func(context.Context, string) (service.TokenPair, error)
}

func (f *fakeTokenIssuer) IssueRefresh(ctx context.Context, user domain.User) (service.RefreshToken, error) {
	if f.issueRefresh == nil {
		return service.RefreshToken{Value: "refresh-token"}, nil
	}
	return f.issueRefresh(ctx, user)
}

func (f *fakeTokenIssuer) Refresh(ctx context.Context, token string) (service.AccessToken, error) {
	if f.refresh == nil {
		return service.AccessToken{Value: "access-token"}, nil
	}
	return f.refresh(ctx, token)
}

func (f *fakeTokenIssuer) RefreshTokens(ctx context.Context, token string) (service.TokenPair, error) {
	if f.refreshTokens == nil {
		return service.TokenPair{UserID: "user-1", AccessToken: service.AccessToken{Value: "access-token"}, RefreshToken: service.RefreshToken{Value: "refresh-token"}}, nil
	}
	return f.refreshTokens(ctx, token)
}

type fakeRefreshSessionRepository struct {
	create  func(context.Context, domain.RefreshSession) error
	consume func(context.Context, string) error
}

func (f *fakeRefreshSessionRepository) Create(ctx context.Context, session domain.RefreshSession) error {
	if f.create == nil {
		return nil
	}
	return f.create(ctx, session)
}

func (f *fakeRefreshSessionRepository) Consume(ctx context.Context, tokenHash string) error {
	if f.consume == nil {
		return nil
	}
	return f.consume(ctx, tokenHash)
}

func (f *fakeRefreshSessionRepository) DeleteExpired(context.Context) error { return nil }

func (f *fakeTokenIssuer) Issue(
	ctx context.Context,
	user domain.User,
) (service.AccessToken, error) {
	if f.issue == nil {
		return service.AccessToken{Value: "access-token"}, nil
	}

	return f.issue(ctx, user)
}

func (f *fakePasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	if f.hash == nil {
		return "password-hash", nil
	}

	return f.hash(ctx, password)
}

func (f *fakePasswordHasher) Matches(
	ctx context.Context,
	passwordHash string,
	password string,
) (bool, error) {
	if f.matches == nil {
		return true, nil
	}

	return f.matches(ctx, passwordHash, password)
}

func TestService_Register(t *testing.T) {
	t.Parallel()

	errDependency := errors.New("dependency error")
	wantUser := domain.User{ID: "user-1", Email: "user@example.com"}
	tests := []struct {
		name    string
		input   service.RegisterInput
		hasher  *fakePasswordHasher
		prepare func(*repositorymocks.UserRepository)
		want    domain.User
		wantErr error
	}{
		{
			name:  "success",
			input: service.RegisterInput{Email: " User@Example.COM ", Password: "password"},
			hasher: &fakePasswordHasher{hash: func(_ context.Context, password string) (string, error) {
				require.Equal(t, "password", password)

				return "password-hash", nil
			}},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().Create(
					mock.Anything,
					mock.MatchedBy(func(user domain.User) bool {
						return user.Email == "user@example.com" && user.PasswordHash == "password-hash"
					}),
				).Return(wantUser, nil)
			},
			want: wantUser,
		},
		{
			name:    "invalid email",
			input:   service.RegisterInput{Email: "invalid", Password: "password"},
			hasher:  &fakePasswordHasher{},
			prepare: func(*repositorymocks.UserRepository) {},
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:    "short password",
			input:   service.RegisterInput{Email: "user@example.com", Password: "short"},
			hasher:  &fakePasswordHasher{},
			prepare: func(*repositorymocks.UserRepository) {},
			wantErr: domain.ErrPasswordTooShort,
		},
		{
			name:  "hasher error",
			input: service.RegisterInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{hash: func(context.Context, string) (string, error) {
				return "", errDependency
			}},
			prepare: func(*repositorymocks.UserRepository) {},
			wantErr: errDependency,
		},
		{
			name:   "duplicate user",
			input:  service.RegisterInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().Create(mock.Anything, mock.AnythingOfType("domain.User")).
					Return(domain.User{}, domain.ErrUserAlreadyExists)
			},
			wantErr: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepository := repositorymocks.NewUserRepository(t)
			tt.prepare(userRepository)
			authService := service.New(userRepository, tt.hasher, &fakeTokenIssuer{}, &fakeRefreshSessionRepository{})

			got, err := authService.Register(t.Context(), tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()

	errDependency := errors.New("dependency error")
	storedUser := domain.User{
		ID: "user-1", Email: "user@example.com", PasswordHash: "password-hash",
	}
	tests := []struct {
		name    string
		input   service.LoginInput
		hasher  *fakePasswordHasher
		prepare func(*repositorymocks.UserRepository)
		issuer  *fakeTokenIssuer
		want    service.LoginResult
		wantErr error
	}{
		{
			name:   "success",
			input:  service.LoginInput{Email: " User@Example.COM ", Password: "password"},
			hasher: &fakePasswordHasher{},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().GetByEmail(mock.Anything, "user@example.com").Return(storedUser, nil)
			},
			issuer: &fakeTokenIssuer{issue: func(_ context.Context, user domain.User) (service.AccessToken, error) {
				require.Equal(t, storedUser, user)

				return service.AccessToken{Value: "access-token"}, nil
			}},
			want: service.LoginResult{
				UserID:       storedUser.ID,
				Email:        storedUser.Email,
				AccessToken:  service.AccessToken{Value: "access-token"},
				RefreshToken: service.RefreshToken{Value: "refresh-token"},
			},
		},
		{
			name:    "invalid email",
			input:   service.LoginInput{Email: "invalid", Password: "password"},
			hasher:  &fakePasswordHasher{},
			issuer:  &fakeTokenIssuer{},
			prepare: func(*repositorymocks.UserRepository) {},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:   "user not found",
			input:  service.LoginInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{},
			issuer: &fakeTokenIssuer{},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().GetByEmail(mock.Anything, "user@example.com").
					Return(domain.User{}, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "wrong password",
			input: service.LoginInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{matches: func(context.Context, string, string) (bool, error) {
				return false, nil
			}},
			issuer: &fakeTokenIssuer{},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().GetByEmail(mock.Anything, "user@example.com").Return(storedUser, nil)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:   "repository error",
			input:  service.LoginInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{},
			issuer: &fakeTokenIssuer{},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().GetByEmail(mock.Anything, "user@example.com").
					Return(domain.User{}, errDependency)
			},
			wantErr: errDependency,
		},
		{
			name:  "hasher error",
			input: service.LoginInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{matches: func(context.Context, string, string) (bool, error) {
				return false, errDependency
			}},
			issuer: &fakeTokenIssuer{},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().GetByEmail(mock.Anything, "user@example.com").Return(storedUser, nil)
			},
			wantErr: errDependency,
		},
		{
			name:   "token issuer error",
			input:  service.LoginInput{Email: "user@example.com", Password: "password"},
			hasher: &fakePasswordHasher{},
			issuer: &fakeTokenIssuer{issue: func(context.Context, domain.User) (service.AccessToken, error) {
				return service.AccessToken{}, errDependency
			}},
			prepare: func(repository *repositorymocks.UserRepository) {
				repository.EXPECT().GetByEmail(mock.Anything, "user@example.com").Return(storedUser, nil)
			},
			wantErr: errDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepository := repositorymocks.NewUserRepository(t)
			tt.prepare(userRepository)
			authService := service.New(userRepository, tt.hasher, tt.issuer, &fakeRefreshSessionRepository{})

			got, err := authService.Login(t.Context(), tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestService_Refresh(t *testing.T) {
	t.Parallel()
	issuer := &fakeTokenIssuer{refreshTokens: func(_ context.Context, token string) (service.TokenPair, error) {
		require.Equal(t, "refresh-token", token)
		return service.TokenPair{UserID: "user-1", AccessToken: service.AccessToken{Value: "new-access"}, RefreshToken: service.RefreshToken{Value: "new-refresh"}}, nil
	}}
	authService := service.New(repositorymocks.NewUserRepository(t), &fakePasswordHasher{}, issuer, &fakeRefreshSessionRepository{})
	got, err := authService.Refresh(t.Context(), service.RefreshInput{RefreshToken: "refresh-token"})
	require.NoError(t, err)
	require.Equal(t, service.TokenPair{UserID: "user-1", AccessToken: service.AccessToken{Value: "new-access"}, RefreshToken: service.RefreshToken{Value: "new-refresh"}}, got)
}

func TestService_Logout(t *testing.T) {
	t.Parallel()
	var consumed string
	sessions := &fakeRefreshSessionRepository{consume: func(_ context.Context, hash string) error {
		consumed = hash
		return nil
	}}
	authService := service.New(repositorymocks.NewUserRepository(t), &fakePasswordHasher{}, &fakeTokenIssuer{}, sessions)
	require.NoError(t, authService.Logout(t.Context(), service.RefreshInput{RefreshToken: "refresh-token"}))
	require.NotEmpty(t, consumed)
}
