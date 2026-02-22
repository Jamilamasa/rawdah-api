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

// ── Tasks ────────────────────────────────────────────────────────────────────

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

// ── Requests ─────────────────────────────────────────────────────────────────

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

// ── Rewards ──────────────────────────────────────────────────────────────────

func TestRewardRepo_CrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewRewardRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	rewardID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO rewards (id, family_id, title, value, type, created_by)
		VALUES ($1, $2, 'Screen time', 1.00, 'virtual', $3)`,
		rewardID, familyA, userA,
	)
	if err != nil {
		t.Fatalf("insert reward: %v", err)
	}

	if _, err := repo.GetByID(ctx, rewardID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch reward: %v", err)
	}

	if _, err := repo.GetByID(ctx, rewardID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant reward lookup to fail")
	}
}

// ── Rewards list isolation ────────────────────────────────────────────────────

func TestRewardRepo_ListCrosstenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewRewardRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	_, err := db.ExecContext(ctx, `
		INSERT INTO rewards (family_id, title, value, type, created_by)
		VALUES ($1, 'FamilyA reward', 1.00, 'virtual', $2)`,
		familyA, userA,
	)
	if err != nil {
		t.Fatalf("insert reward: %v", err)
	}

	rewards, err := repo.List(ctx, familyB.String())
	if err != nil {
		t.Fatalf("list rewards: %v", err)
	}
	if len(rewards) != 0 {
		t.Fatalf("expected 0 rewards for familyB, got %d", len(rewards))
	}
}

// ── Messages ─────────────────────────────────────────────────────────────────

func TestMessageRepo_CrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewMessageRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA1, userA2 := uuid.New(), uuid.New()
	userB1 := uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA1)

	// second user in familyA
	emailA2 := fmt.Sprintf("%s@example.com", uuid.NewString()[:8])
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, family_id, role, name, email, password_hash, theme, game_limit_minutes, is_active)
		VALUES ($1, $2, 'parent', 'Parent A', $3, 'hash', 'forest', 60, TRUE)`,
		userA2, familyA, emailA2,
	)
	if err != nil {
		t.Fatalf("insert userA2: %v", err)
	}

	seedFamilyAndUser(t, ctx, db, familyB, userB1)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	_, err = db.ExecContext(ctx, `
		INSERT INTO messages (family_id, sender_id, recipient_id, content)
		VALUES ($1, $2, $3, 'Hello from A')`,
		familyA, userA1, userA2,
	)
	if err != nil {
		t.Fatalf("insert message: %v", err)
	}

	// FamilyB should see zero messages in its thread
	msgs, err := repo.GetThread(ctx, userB1.String(), userA1.String(), familyB.String())
	if err != nil {
		t.Fatalf("get thread: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages for familyB thread, got %d", len(msgs))
	}

	// FamilyB conversations should be empty
	convos, err := repo.Conversations(ctx, userB1.String(), familyB.String())
	if err != nil {
		t.Fatalf("conversations: %v", err)
	}
	if len(convos) != 0 {
		t.Fatalf("expected 0 conversations for familyB, got %d", len(convos))
	}
}

// ── Notifications ─────────────────────────────────────────────────────────────

func TestNotificationRepo_CrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewNotificationRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	_, err := db.ExecContext(ctx, `
		INSERT INTO notifications (user_id, type, title, body)
		VALUES ($1, 'task_assigned', 'New Task', 'You have a task')`,
		userA,
	)
	if err != nil {
		t.Fatalf("insert notification: %v", err)
	}

	// userB should see only their own notifications (zero from userA)
	notifs, err := repo.List(ctx, userB.String())
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notifs) != 0 {
		t.Fatalf("expected 0 notifications for userB, got %d", len(notifs))
	}
}

// ── Game Sessions ─────────────────────────────────────────────────────────────

func TestGameRepo_CrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewGameRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	_, err := db.ExecContext(ctx, `
		INSERT INTO game_sessions (user_id, family_id, game_name, game_type, started_at)
		VALUES ($1, $2, 'names-match', 'islamic', NOW())`,
		userA, familyA,
	)
	if err != nil {
		t.Fatalf("insert game session: %v", err)
	}

	sessions, err := repo.ListSessions(ctx, familyB.String(), "")
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions for familyB, got %d", len(sessions))
	}
}

// ── Hadith Quizzes ────────────────────────────────────────────────────────────

func TestQuizRepo_HadithCrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewQuizRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	// Need a hadith to reference
	hadithID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO hadiths (id, text_en, source, difficulty)
		VALUES ($1, 'Actions are judged by intentions.', 'Bukhari', 'easy')`,
		hadithID,
	)
	if err != nil {
		t.Fatalf("insert hadith: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM hadiths WHERE id = $1`, hadithID)

	quizID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO hadith_quizzes (id, family_id, hadith_id, assigned_to, assigned_by, questions, status)
		VALUES ($1, $2, $3, $4, $4, '[]'::jsonb, 'pending')`,
		quizID, familyA, hadithID, userA,
	)
	if err != nil {
		t.Fatalf("insert hadith quiz: %v", err)
	}

	if _, err := repo.GetHadithQuizByID(ctx, quizID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch quiz: %v", err)
	}

	if _, err := repo.GetHadithQuizByID(ctx, quizID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant hadith quiz lookup to fail")
	}
}

// ── Prophet Quizzes ───────────────────────────────────────────────────────────

func TestQuizRepo_ProphetCrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewQuizRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	prophetID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO prophets (id, name_en, story_summary, difficulty)
		VALUES ($1, 'Musa (Test)', 'Led the Israelites out of Egypt.', 'medium')`,
		prophetID,
	)
	if err != nil {
		t.Fatalf("insert prophet: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM prophets WHERE id = $1`, prophetID)

	quizID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO prophet_quizzes (id, family_id, prophet_id, assigned_to, assigned_by, questions, status)
		VALUES ($1, $2, $3, $4, $4, '[]'::jsonb, 'pending')`,
		quizID, familyA, prophetID, userA,
	)
	if err != nil {
		t.Fatalf("insert prophet quiz: %v", err)
	}

	if _, err := repo.GetProphetQuizByID(ctx, quizID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch quiz: %v", err)
	}

	if _, err := repo.GetProphetQuizByID(ctx, quizID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant prophet quiz lookup to fail")
	}
}

// ── Quran Quizzes ─────────────────────────────────────────────────────────────

func TestQuizRepo_QuranCrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewQuizRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	verseID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO quran_verses (id, surah_number, ayah_number, surah_name_en, text_ar, text_en, tafsir_simple, difficulty)
		VALUES ($1, 999, 1, 'Test Surah', 'بسم الله', 'In the name of Allah', 'Opening invocation.', 'easy')`,
		verseID,
	)
	if err != nil {
		t.Fatalf("insert verse: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM quran_verses WHERE id = $1`, verseID)

	quizID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO quran_quizzes (id, family_id, verse_id, assigned_to, questions, status)
		VALUES ($1, $2, $3, $4, '[]'::jsonb, 'pending')`,
		quizID, familyA, verseID, userA,
	)
	if err != nil {
		t.Fatalf("insert quran quiz: %v", err)
	}

	if _, err := repo.GetQuranQuizByID(ctx, quizID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch quiz: %v", err)
	}

	if _, err := repo.GetQuranQuizByID(ctx, quizID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant quran quiz lookup to fail")
	}
}

// ── Quran Lessons ─────────────────────────────────────────────────────────────

func TestLessonRepo_QuranCrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewLessonRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	verseID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO quran_verses (id, surah_number, ayah_number, surah_name_en, text_ar, text_en, tafsir_simple, difficulty)
		VALUES ($1, 998, 1, 'Lesson Test Surah', 'بسم الله', 'In the name of Allah', 'Opening invocation.', 'easy')`,
		verseID,
	)
	if err != nil {
		t.Fatalf("insert verse: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM quran_verses WHERE id = $1`, verseID)

	lessonID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO quran_lessons (id, family_id, verse_id, assigned_to, assigned_by, status)
		VALUES ($1, $2, $3, $4, $4, 'pending')`,
		lessonID, familyA, verseID, userA,
	)
	if err != nil {
		t.Fatalf("insert quran lesson: %v", err)
	}

	if _, err := repo.GetLesson(ctx, lessonID.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch lesson: %v", err)
	}

	if _, err := repo.GetLesson(ctx, lessonID.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant lesson lookup to fail")
	}
}

// ── Learn Content ─────────────────────────────────────────────────────────────

func TestLessonRepo_LearnContentCrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewLessonRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	contentID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO learn_content (id, family_id, title, content_type, content, created_by)
		VALUES ($1, $2, 'Islamic Article', 'text', 'Content here...', $3)`,
		contentID, familyA, userA,
	)
	if err != nil {
		t.Fatalf("insert learn content: %v", err)
	}

	// FamilyB listing should not include FamilyA content
	contents, err := repo.GetLearnContent(ctx, familyB.String())
	if err != nil {
		t.Fatalf("list learn content: %v", err)
	}
	for _, c := range contents {
		if c.ID == contentID {
			t.Fatalf("familyB can see familyA learn content — isolation failure")
		}
	}
}

// ── Family Members ────────────────────────────────────────────────────────────

func TestFamilyRepo_MemberCrossTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := repository.NewFamilyRepo(db)

	familyA, familyB := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedFamilyAndUser(t, ctx, db, familyA, userA)
	seedFamilyAndUser(t, ctx, db, familyB, userB)
	defer cleanupFamilies(t, ctx, db, familyA, familyB)

	// userA belongs to familyA — familyB must not be able to fetch them
	if _, err := repo.GetMemberByID(ctx, userA.String(), familyA.String()); err != nil {
		t.Fatalf("expected owner family to fetch member: %v", err)
	}

	if _, err := repo.GetMemberByID(ctx, userA.String(), familyB.String()); err == nil {
		t.Fatalf("expected cross-tenant member lookup to fail")
	}
}
