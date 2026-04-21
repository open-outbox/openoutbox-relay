//go:build integration

package integration

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	dockercontainer "github.com/moby/moby/api/types/container"
	nats_go "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nats"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestPostgres(t *testing.T) (*sql.DB, string) {
	t.Helper()
	ctx := context.Background()

	schemaPath, err := filepath.Abs("../../schema/postgres/open-outbox.sql")
	require.NoError(t, err)

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("outbox_db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		postgres.WithInitScripts(schemaPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(20*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		pgContainer.Terminate(context.Background())
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	return db, connStr
}

func setupNats(t *testing.T) (*nats_go.Conn, string) {
	t.Helper()
	ctx := context.Background()

	natsContainer, err := nats.Run(ctx, "nats:2.10-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("Server is ready")),
		testcontainers.WithConfigModifier(func(conf *dockercontainer.Config) {
			conf.Cmd = []string{"-js"}
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { natsContainer.Terminate(context.Background()) })

	url, _ := natsContainer.ConnectionString(ctx)
	nc, err := nats_go.Connect(url)
	require.NoError(t, err)

	return nc, url
}
