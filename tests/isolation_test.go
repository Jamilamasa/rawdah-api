package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rawdah/rawdah-api/internal/repository"
)

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run isolation tests")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}

	return db
}

func seedFamilyAndUser(t *testing.T, ctx context.Context, db *sqlx.DB, familyID, userID uuid.UUID) {
	t.Helper()

	slug := fmt.Sprintf("family-%s", uuid.NewString()[:8])
	email := fmt.Sprintf("%s@example.com", uuid.NewString()[:8])
	passwordHash := "$2a$12$7iQWj7g2SdA3j7g2SdA3jO2N93M4y4qGb3h2v2H5xg8kWgY1bM1xS"

	_, err := db.ExecContext(ctx, `
		INSERT INTO families (id, name, slug, plan)
		VALUES ($1, $2, $3, 'free')`,
		familyID, "Test Family", slug,
	)
	if err != nil {
		t.Fatalf("insert family: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, family_id, role, name, email, password_hash, theme, game_limit_minutes, is_active)
		VALUES ($1, $2, 'child', $3, $4, $5, 'forest', 60, TRUE)`,
		userID, familyID, "Test Child", email, passwordHash,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func cleanupFamilies(t *testing.T, ctx context.Context, db *sqlx.DB, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		if _, err := db.ExecContext(ctx, `DELETE FROM families WHERE id = $1`, id); err != nil {
			t.Fatalf("cleanup family %s: %v", id, err)
		}
	}
}

func TestTaskRepo_CrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewTaskRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	taskID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO tasks (id, family_id, title, assigned_to, created_by, status)
		VALUES ($1, $2, 'Cross tenant task', $3, $4, 'pending')`,
		taskID, familyA, userA, userA,
	)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	if _, err := repo.GetByID(ctx, taskID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch task: %v", err)
	}

	if _, err := repo.GetByID(ctx, taskID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant lookup to fail")
	}
}

func TestRequestRepo_CrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewRequestRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	requestID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO requests (id, family_id, requester_id, title, status)
		VALUES ($1, $2, $3, 'Need help', 'pending')`,
		requestID, familyA, userA,
	)
	if err != nil {
		t.Fatalf("insert request: %v", err)
	}

	if _, err := repo.GetByID(ctx, requestID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch request: %v", err)
	}

	if _, err := repo.GetByID(ctx, requestID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant lookup to fail")
	}
}
