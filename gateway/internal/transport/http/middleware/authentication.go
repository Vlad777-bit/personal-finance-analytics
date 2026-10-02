package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/token"
	httptransport "github.com/Vlad777-bit/personal-finance-analytics/gateway/internal/transport/http"
)

type identityContextKey struct{}

func Authenticate(verifier token.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeUnauthorized(w)

				return
			}

			identity, err := verifier.Verify(r.Context(), value)
			if err != nil {
				writeUnauthorized(w)

				return
			}

			ctx := context.WithValue(r.Context(), identityContextKey{}, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func IdentityFromContext(ctx context.Context) (token.Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(token.Identity)

	return identity, ok
}

func WithIdentity(ctx context.Context, identity token.Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func writeUnauthorized(w http.ResponseWriter) {
	httptransport.WriteError(
		w,
		http.StatusUnauthorized,
		"unauthorized",
		"valid bearer token is required",
	)
}
