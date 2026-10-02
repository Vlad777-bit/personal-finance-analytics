//go:build integration

package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	reflectionv1 "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/grpc/status"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
	authv1 "github.com/Vlad777-bit/personal-finance-analytics/shared/gen/go/auth/v1"
)

func TestAuthGRPC(t *testing.T) {
	const email = "auth-app-integration@example.com"

	databaseURL := os.Getenv("AUTH_DATABASE_URL")
	require.NotEmpty(t, databaseURL, "AUTH_DATABASE_URL must be set")

	applicationContext, cancelApplication := context.WithCancel(context.Background())
	application, err := New(
		applicationContext,
		databaseURL,
		4,
		"0123456789abcdef0123456789abcdef",
		"auth-integration-test",
		15*time.Minute,
		"127.0.0.1:0",
	)
	require.NoError(t, err)

	runErrors := make(chan error, 1)
	go func() {
		runErrors <- application.Run(applicationContext)
	}()

	t.Cleanup(func() {
		cancelApplication()
		require.NoError(t, <-runErrors)

		shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, application.Shutdown(shutdownContext))
	})

	cleanupUser(t, application.database, email)
	t.Cleanup(func() { cleanupUser(t, application.database, email) })

	connection, err := grpc.NewClient(
		application.listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, connection.Close()) })

	client := authv1.NewAuthServiceClient(connection)
	rpcContext, cancelRPC := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelRPC()

	registerResponse, err := client.Register(rpcContext, &authv1.RegisterRequest{
		Email: email, Password: "secure-password",
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResponse.GetUser().GetId())
	require.Equal(t, email, registerResponse.GetUser().GetEmail())
	require.NotNil(t, registerResponse.GetUser().GetCreatedAt())

	_, err = client.Register(rpcContext, &authv1.RegisterRequest{
		Email: email, Password: "secure-password",
	})
	require.Equal(t, codes.AlreadyExists, status.Code(err))

	loginResponse, err := client.Login(rpcContext, &authv1.LoginRequest{
		Email: email, Password: "secure-password",
	})
	require.NoError(t, err)
	require.Equal(t, registerResponse.GetUser().GetId(), loginResponse.GetUserId())
	require.Equal(t, email, loginResponse.GetEmail())
	require.NotEmpty(t, loginResponse.GetAccessToken())
	require.WithinDuration(
		t,
		time.Now().Add(15*time.Minute),
		loginResponse.GetExpiresAt().AsTime(),
		time.Second,
	)

	_, err = client.Login(rpcContext, &authv1.LoginRequest{
		Email: email, Password: "wrong-password",
	})
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	_, err = client.Register(rpcContext, &authv1.RegisterRequest{})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	reflectionClient := reflectionv1.NewServerReflectionClient(connection)
	reflectionStream, err := reflectionClient.ServerReflectionInfo(rpcContext)
	require.NoError(t, err)
	require.NoError(t, reflectionStream.Send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_ListServices{},
	}))
	reflectionResponse, err := reflectionStream.Recv()
	require.NoError(t, err)
	requireServiceRegistered(t, reflectionResponse, authv1.AuthService_ServiceDesc.ServiceName)
}

func requireServiceRegistered(
	t *testing.T,
	response *reflectionv1.ServerReflectionResponse,
	serviceName string,
) {
	t.Helper()

	for _, registeredService := range response.GetListServicesResponse().GetService() {
		if registeredService.GetName() == serviceName {
			return
		}
	}

	require.Failf(t, "gRPC service is not registered", "service: %s", serviceName)
}

func cleanupUser(t *testing.T, db database.DB, email string) {
	t.Helper()

	_, err := db.Exec(
		context.Background(),
		"DELETE FROM users WHERE email = $1",
		email,
	)
	require.NoError(t, err)
}
