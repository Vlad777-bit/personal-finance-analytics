package app

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
)

type fakeDatabase struct {
	closed chan struct{}
}

func (*fakeDatabase) QueryRow(context.Context, string, ...any) database.Row {
	return nil
}

func (*fakeDatabase) Exec(context.Context, string, ...any) (database.Result, error) {
	return nil, nil
}

func (f *fakeDatabase) Close() {
	close(f.closed)
}

func TestApp_Lifecycle(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	databaseClosed := make(chan struct{})
	application := &App{
		database: &fakeDatabase{closed: databaseClosed},
		server:   grpc.NewServer(),
		listener: listener,
	}
	ctx, cancel := context.WithCancel(t.Context())
	runErrors := make(chan error, 1)
	go func() {
		runErrors <- application.Run(ctx)
	}()

	cancel()
	require.NoError(t, <-runErrors)
	require.NoError(t, application.Shutdown(t.Context()))

	select {
	case <-databaseClosed:
	case <-t.Context().Done():
		require.Fail(t, "database was not closed")
	}
}
