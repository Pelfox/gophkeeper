package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	testDatabaseDSNEnv = "GOPHKEEPER_TEST_DATABASE_DSN"
)

// testPool creates an isolated migrated schema for a repository integration
// test, or skips the test when the database DSN is not configured.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(testDatabaseDSNEnv)
	if dsn == "" {
		t.Skipf("%s is not set", testDatabaseDSNEnv)
	}

	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	if _, err := adminPool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		adminPool.Close()
		t.Fatalf("failed to create test schema: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		adminPool.Close()
		t.Fatalf("failed to parse test database DSN: %v", err)
	}
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		adminPool.Close()
		t.Fatalf("failed to create test database pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		if _, err := adminPool.Exec(
			context.Background(),
			fmt.Sprintf(`DROP SCHEMA IF EXISTS "%s" CASCADE`, schema),
		); err != nil {
			t.Logf("failed to drop test schema %q: %v", schema, err)
		}
		adminPool.Close()
	})

	applyMigrations(t, ctx, pool)
	return pool
}

// applyMigrations applies all up migrations to the current test schema.
func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrations, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("failed to list migrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("no migrations found")
	}
	sort.Strings(migrations)

	for _, path := range migrations {
		sql, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", path, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("failed to apply migration %s: %v", path, err)
		}
	}
}

// createTestUser inserts a user fixture and returns its generated ID.
func createTestUser(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	email string,
) uuid.UUID {
	t.Helper()

	user, err := NewUsersRepository(pool).Create(ctx, CreateUserInput{
		Email:        email,
		PasswordHash: "password-hash",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user.ID
}

// createTestVault inserts a vault fixture and returns its generated ID.
func createTestVault(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	ownerID uuid.UUID,
	name string,
) uuid.UUID {
	t.Helper()

	vault, err := NewVaultsRepository(pool).Create(ctx, CreateVaultInput{
		OwnerID: ownerID,
		Name:    name,
	})
	if err != nil {
		t.Fatalf("failed to create test vault: %v", err)
	}
	return vault.ID
}

// assertPgErrorCode verifies that an error wraps a Postgres error with the
// expected SQLSTATE code.
func assertPgErrorCode(t *testing.T, err error, code string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected Postgres error with code %s, got nil", code)
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected Postgres error with code %s, got %v", code, err)
	}
	if pgErr.Code != code {
		t.Fatalf("unexpected Postgres error code: got %s want %s", pgErr.Code, code)
	}
}
