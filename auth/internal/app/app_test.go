package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vlad777-bit/personal-finance-analytics/auth/internal/database"
)

type fakeDatabase struct {
	close func()
}

func (*fakeDatabase) QueryRow(context.Context, string, ...any) database.Row {
	return nil
}

func (*fakeDatabase) Exec(context.Context, string, ...any) (database.Result, error) {
	return nil, nil
}

func (f *fakeDatabase) Close() {
	f.close()
}

func TestApp_Run(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := (&App{}).Run(ctx)
	require.NoError(t, err)
}

func TestApp_Shutdown(t *testing.T) {
	t.Parallel()

	closed := make(chan struct{})
	application := &App{database: &fakeDatabase{close: func() { close(closed) }}}

	err := application.Shutdown(t.Context())
	require.NoError(t, err)
	requireClosed(t, closed)
}

func TestApp_Shutdown_Timeout(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	closed := make(chan struct{})
	application := &App{database: &fakeDatabase{close: func() {
		<-release
		close(closed)
	}}}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := application.Shutdown(ctx)
	require.ErrorIs(t, err, context.Canceled)
	close(release)
	requireClosed(t, closed)
}

func requireClosed(t *testing.T, channel <-chan struct{}) {
	t.Helper()

	select {
	case <-channel:
	case <-t.Context().Done():
		require.Fail(t, "database was not closed")
	}
}
