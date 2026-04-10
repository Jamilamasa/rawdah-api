package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Family struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Slug      string    `db:"slug"       json:"slug"`
	LogoURL   *string   `db:"logo_url"   json:"logo_url"`
	Plan      string    `db:"plan"       json:"plan"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type User struct {
	ID               uuid.UUID  `db:"id"                  json:"id"`
	FamilyID         uuid.UUID  `db:"family_id"           json:"family_id"`
	Role             string     `db:"role"                json:"role"`
	Name             string     `db:"name"                json:"name"`
	Username         *string    `db:"username"            json:"username"`
	Email            *string    `db:"email"               json:"email"`
	PasswordHash     string     `db:"password_hash"       json:"-"`
	AvatarURL        *string    `db:"avatar_url"          json:"avatar_url"`
	Theme            string     `db:"theme"               json:"theme"`
	DateOfBirth      *time.Time `db:"date_of_birth"       json:"date_of_birth"`
	GameLimitMinutes int        `db:"game_limit_minutes"  json:"game_limit_minutes"`
	IsActive         bool       `db:"is_active"           json:"is_active"`
	CreatedBy        *uuid.UUID `db:"created_by"          json:"created_by"`
	LastLoginAt      *time.Time `db:"last_login_at"       json:"last_login_at"`
	CreatedAt        time.Time  `db:"created_at"          json:"created_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
}

type FamilyAccessControl struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	FamilyID    uuid.UUID `db:"family_id"   json:"family_id"`
	GrantorID   uuid.UUID `db:"grantor_id"  json:"grantor_id"`
	GranteeID   uuid.UUID `db:"grantee_id"  json:"grantee_id"`
	Permissions []string  `db:"permissions" json:"permissions"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
}

type Reward struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	FamilyID    uuid.UUID `db:"family_id"   json:"family_id"`
	Title       string    `db:"title"       json:"title"`
	Description *string   `db:"description" json:"description"`
	Value       float64   `db:"value"       json:"value"`
	Type        string    `db:"type"        json:"type"`
	Icon        *string   `db:"icon"        json:"icon"`
	CreatedBy   uuid.UUID `db:"created_by"  json:"created_by"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
}

type Task struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	FamilyID    uuid.UUID  `db:"family_id"    json:"family_id"`
	Title       string     `db:"title"        json:"title"`
	Description *string    `db:"description"  json:"description"`
	AssignedTo  uuid.UUID  `db:"assigned_to"  json:"assigned_to"`
	CreatedBy   uuid.UUID  `db:"created_by"   json:"created_by"`
	RewardID    *uuid.UUID `db:"reward_id"    json:"reward_id"`
	Status      string     `db:"status"       json:"status"`
	DueDate     *time.Time `db:"due_date"     json:"due_date"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

type DueReward struct {
	TaskID            uuid.UUID  `db:"task_id"             json:"task_id"`
	TaskTitle         string     `db:"task_title"          json:"task_title"`
	TaskDescription   *string    `db:"task_description"    json:"task_description"`
	TaskStatus        string     `db:"task_status"         json:"task_status"`
	TaskDueDate       *time.Time `db:"task_due_date"       json:"task_due_date"`
	TaskCompletedAt   *time.Time `db:"task_completed_at"   json:"task_completed_at"`
	TaskCreatedAt     time.Time  `db:"task_created_at"     json:"task_created_at"`
	ChildID           uuid.UUID  `db:"child_id"            json:"child_id"`
	ChildName         string     `db:"child_name"          json:"child_name"`
	RewardID          uuid.UUID  `db:"reward_id"           json:"reward_id"`
	RewardTitle       string     `db:"reward_title"        json:"reward_title"`
	RewardDescription *string    `db:"reward_description"  json:"reward_description"`
	RewardValue       float64    `db:"reward_value"        json:"reward_value"`
	RewardType        string     `db:"reward_type"         json:"reward_type"`
	RewardIcon        *string    `db:"reward_icon"         json:"reward_icon"`
}

type Hadith struct {
	ID         uuid.UUID `db:"id"         json:"id"`
	TextEn     string    `db:"text_en"    json:"text_en"`
	TextAr     *string   `db:"text_ar"    json:"text_ar"`
	Source     string    `db:"source"     json:"source"`
	Topic      *string   `db:"topic"      json:"topic"`
	Difficulty string    `db:"difficulty" json:"difficulty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Prophet struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	NameEn       string    `db:"name_en"       json:"name_en"`
	NameAr       *string   `db:"name_ar"       json:"name_ar"`
	OrderNum     *int      `db:"order_num"     json:"order_num"`
	StorySummary string    `db:"story_summary" json:"story_summary"`
	KeyMiracles  *string   `db:"key_miracles"  json:"key_miracles"`
	Nation       *string   `db:"nation"        json:"nation"`
	QuranRefs    *string   `db:"quran_refs"    json:"quran_refs"`
	Difficulty   string    `db:"difficulty"    json:"difficulty"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type QuranVerse struct {
	ID              uuid.UUID `db:"id"               json:"id"`
	SurahNumber     int       `db:"surah_number"     json:"surah_number"`
	AyahNumber      int       `db:"ayah_number"      json:"ayah_number"`
	SurahNameEn     string    `db:"surah_name_en"    json:"surah_name_en"`
	TextAr          string    `db:"text_ar"          json:"text_ar"`
	TextEn          string    `db:"text_en"          json:"text_en"`
	Transliteration *string   `db:"transliteration"  json:"transliteration"`
	TafsirSimple    string    `db:"tafsir_simple"    json:"tafsir_simple"`
	LifeApplication *string   `db:"life_application" json:"life_application"`
	Topic           *string   `db:"topic"            json:"topic"`
	Difficulty      string    `db:"difficulty"       json:"difficulty"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
}

type QuizQuestion struct {
	ID            string            `json:"id"`
	Question      string            `json:"question"`
	Options       map[string]string `json:"options"`
	CorrectAnswer string            `json:"correct_answer"`
	Explanation   string            `json:"explanation"`
}

type Flashcard struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

type QuizAnswer struct {
	QuestionID     string `json:"question_id"`
	SelectedAnswer string `json:"selected_answer"`
	IsCorrect      bool   `json:"is_correct"`
}

type HadithQuiz struct {
	ID            uuid.UUID      `db:"id"             json:"id"`
	FamilyID      uuid.UUID      `db:"family_id"      json:"family_id"`
	HadithID      uuid.UUID      `db:"hadith_id"      json:"hadith_id"`
	AssignedTo    uuid.UUID      `db:"assigned_to"    json:"assigned_to"`
	AssignedBy    uuid.UUID      `db:"assigned_by"    json:"assigned_by"`
	Questions     []QuizQuestion `db:"questions"      json:"questions"`
	Answers       []QuizAnswer   `db:"answers"        json:"answers,omitempty"`
	Score         *int           `db:"score"          json:"score"`
	XPAwarded     int            `db:"xp_awarded"     json:"xp_awarded"`
	Status        string         `db:"status"         json:"status"`
	MemorizeUntil *time.Time     `db:"memorize_until" json:"memorize_until"`
	CompletedAt   *time.Time     `db:"completed_at"   json:"completed_at"`
	CreatedAt     time.Time      `db:"created_at"     json:"created_at"`
}

type ProphetQuiz struct {
	ID          uuid.UUID      `db:"id"          json:"id"`
	FamilyID    uuid.UUID      `db:"family_id"   json:"family_id"`
	ProphetID   uuid.UUID      `db:"prophet_id"  json:"prophet_id"`
	AssignedTo  uuid.UUID      `db:"assigned_to" json:"assigned_to"`
	AssignedBy  uuid.UUID      `db:"assigned_by" json:"assigned_by"`
	Questions   []QuizQuestion `db:"questions"   json:"questions"`
	Answers     []QuizAnswer   `db:"answers"     json:"answers,omitempty"`
	Score       *int           `db:"score"       json:"score"`
	XPAwarded   int            `db:"xp_awarded"  json:"xp_awarded"`
	Status      string         `db:"status"      json:"status"`
	CompletedAt *time.Time     `db:"completed_at" json:"completed_at"`
	CreatedAt   time.Time      `db:"created_at"  json:"created_at"`
}

type QuranLesson struct {
	ID          uuid.UUID  `db:"id"          json:"id"`
	FamilyID    uuid.UUID  `db:"family_id"   json:"family_id"`
	VerseID     uuid.UUID  `db:"verse_id"    json:"verse_id"`
	AssignedTo  uuid.UUID  `db:"assigned_to" json:"assigned_to"`
	AssignedBy  uuid.UUID  `db:"assigned_by" json:"assigned_by"`
	RewardID    *uuid.UUID `db:"reward_id"   json:"reward_id"`
	Status      string     `db:"status"      json:"status"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`
	CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
}

type QuranQuiz struct {
	ID          uuid.UUID      `db:"id"          json:"id"`
	FamilyID    uuid.UUID      `db:"family_id"   json:"family_id"`
	VerseID     uuid.UUID      `db:"verse_id"    json:"verse_id"`
	LessonID    *uuid.UUID     `db:"lesson_id"   json:"lesson_id"`
	AssignedTo  uuid.UUID      `db:"assigned_to" json:"assigned_to"`
	Questions   []QuizQuestion `db:"questions"   json:"questions"`
	Answers     []QuizAnswer   `db:"answers"     json:"answers,omitempty"`
	Score       *int           `db:"score"       json:"score"`
	XPAwarded   int            `db:"xp_awarded"  json:"xp_awarded"`
	Status      string         `db:"status"      json:"status"`
	CompletedAt *time.Time     `db:"completed_at" json:"completed_at"`
	CreatedAt   time.Time      `db:"created_at"  json:"created_at"`
}

type TopicQuiz struct {
	ID            uuid.UUID      `db:"id"             json:"id"`
	FamilyID      uuid.UUID      `db:"family_id"      json:"family_id"`
	AssignedTo    uuid.UUID      `db:"assigned_to"    json:"assigned_to"`
	AssignedBy    uuid.UUID      `db:"assigned_by"    json:"assigned_by"`
	Category      string         `db:"category"       json:"category"`
	Topic         string         `db:"topic"          json:"topic"`
	LessonContent string         `db:"lesson_content" json:"lesson_content"`
	Flashcards    []Flashcard    `db:"flashcards"     json:"flashcards"`
	Questions     []QuizQuestion `db:"questions"      json:"questions"`
	Answers       []QuizAnswer   `db:"answers"        json:"answers,omitempty"`
	Score         *int           `db:"score"          json:"score"`
	XPAwarded     int            `db:"xp_awarded"     json:"xp_awarded"`
	Status        string         `db:"status"         json:"status"`
	CompletedAt   *time.Time     `db:"completed_at"   json:"completed_at"`
	CreatedAt     time.Time      `db:"created_at"     json:"created_at"`
}

type Message struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	FamilyID    uuid.UUID  `db:"family_id"    json:"family_id"`
	SenderID    uuid.UUID  `db:"sender_id"    json:"sender_id"`
	RecipientID uuid.UUID  `db:"recipient_id" json:"recipient_id"`
	Content     string     `db:"content"      json:"content"`
	ReadAt      *time.Time `db:"read_at"      json:"read_at"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

type Rant struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	UserID       uuid.UUID `db:"user_id"       json:"user_id"`
	Title        *string   `db:"title"         json:"title"`
	Content      string    `db:"content"       json:"content,omitempty"`
	PasswordHash *string   `db:"password_hash" json:"-"`
	IsLocked     bool      `json:"is_locked"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

type Request struct {
	ID              uuid.UUID  `db:"id"               json:"id"`
	FamilyID        uuid.UUID  `db:"family_id"        json:"family_id"`
	RequesterID     uuid.UUID  `db:"requester_id"     json:"requester_id"`
	TargetID        *uuid.UUID `db:"target_id"        json:"target_id"`
	Title           string     `db:"title"            json:"title"`
	Description     *string    `db:"description"      json:"description"`
	Status          string     `db:"status"           json:"status"`
	ResponseMessage *string    `db:"response_message" json:"response_message"`
	RespondedBy     *uuid.UUID `db:"responded_by"     json:"responded_by"`
	RespondedAt     *time.Time `db:"responded_at"     json:"responded_at"`
	CreatedAt       time.Time  `db:"created_at"       json:"created_at"`
}

type GameSession struct {
	ID              uuid.UUID  `db:"id"               json:"id"`
	UserID          uuid.UUID  `db:"user_id"          json:"user_id"`
	FamilyID        uuid.UUID  `db:"family_id"        json:"family_id"`
	GameName        string     `db:"game_name"        json:"game_name"`
	GameType        string     `db:"game_type"        json:"game_type"`
	StartedAt       time.Time  `db:"started_at"       json:"started_at"`
	EndedAt         *time.Time `db:"ended_at"         json:"ended_at"`
	DurationSeconds *int       `db:"duration_seconds" json:"duration_seconds"`
}

type Notification struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	Type      string    `db:"type"       json:"type"`
	Title     string    `db:"title"      json:"title"`
	Body      string    `db:"body"       json:"body"`
	Data      *string   `db:"data"       json:"data"`
	IsRead    bool      `db:"is_read"    json:"is_read"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type DuaSelectedName struct {
	Transliteration string `json:"transliteration"`
	DuaForm         string `json:"dua_form"`
	Arabic          string `json:"arabic"`
	Meaning         string `json:"meaning"`
	Explanation     string `json:"explanation"`
}

type DuaHistory struct {
	ID            uuid.UUID         `db:"id"             json:"id"`
	FamilyID      uuid.UUID         `db:"family_id"      json:"family_id"`
	UserID        uuid.UUID         `db:"user_id"        json:"user_id"`
	AskingFor     string            `db:"asking_for"     json:"asking_for"`
	HeavyOnHeart  string            `db:"heavy_on_heart" json:"heavy_on_heart"`
	AfraidOf      string            `db:"afraid_of"      json:"afraid_of"`
	IfAnswered    string            `db:"if_answered"    json:"if_answered"`
	OutputStyle   string            `db:"output_style"   json:"output_style"`
	Depth         string            `db:"depth"          json:"depth"`
	Tone          string            `db:"tone"           json:"tone"`
	SelectedNames []DuaSelectedName `db:"selected_names" json:"selected_names"`
	DuaText       string            `db:"dua_text"       json:"dua_text"`
	CreatedAt     time.Time         `db:"created_at"     json:"created_at"`
}

type PushSubscription struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	Endpoint  string    `db:"endpoint"   json:"endpoint"`
	P256dh    string    `db:"p256dh"     json:"p256dh"`
	Auth      string    `db:"auth"       json:"auth"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UserXP struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	FamilyID  uuid.UUID `db:"family_id"  json:"family_id"`
	TotalXP   int       `db:"total_xp"   json:"total_xp"`
	Level     int       `db:"level"      json:"level"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type XPEvent struct {
	ID        uuid.UUID `db:"id"        json:"id"`
	UserID    uuid.UUID `db:"user_id"   json:"user_id"`
	FamilyID  uuid.UUID `db:"family_id" json:"family_id"`
	Source    string    `db:"source"    json:"source"`
	SourceID  uuid.UUID `db:"source_id" json:"source_id"`
	XPAmount  int       `db:"xp_amount" json:"xp_amount"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Badge struct {
	ID          uuid.UUID `db:"id"          json:"id"`
	Slug        string    `db:"slug"        json:"slug"`
	Name        string    `db:"name"        json:"name"`
	Description string    `db:"description" json:"description"`
	Icon        string    `db:"icon"        json:"icon"`
	XPReward    int       `db:"xp_reward"   json:"xp_reward"`
}

type UserBadge struct {
	ID        uuid.UUID `db:"id"        json:"id"`
	UserID    uuid.UUID `db:"user_id"   json:"user_id"`
	BadgeID   uuid.UUID `db:"badge_id"  json:"badge_id"`
	AwardedAt time.Time `db:"awarded_at" json:"awarded_at"`
}

type LearnContent struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	FamilyID    uuid.UUID  `db:"family_id"    json:"family_id"`
	AssignedTo  *uuid.UUID `db:"assigned_to"  json:"assigned_to"`
	Title       string     `db:"title"        json:"title"`
	ContentType string     `db:"content_type" json:"content_type"`
	Content     string     `db:"content"      json:"content"`
	RewardID    *uuid.UUID `db:"reward_id"    json:"reward_id"`
	CreatedBy   uuid.UUID  `db:"created_by"   json:"created_by"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

type LearnProgress struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	ContentID   uuid.UUID  `db:"content_id"   json:"content_id"`
	UserID      uuid.UUID  `db:"user_id"      json:"user_id"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`
}

type RecurringTask struct {
	ID          uuid.UUID  `db:"id"          json:"id"`
	FamilyID    uuid.UUID  `db:"family_id"   json:"family_id"`
	Title       string     `db:"title"       json:"title"`
	Description *string    `db:"description" json:"description"`
	AssignedTo  uuid.UUID  `db:"assigned_to" json:"assigned_to"`
	CreatedBy   uuid.UUID  `db:"created_by"  json:"created_by"`
	RewardID    *uuid.UUID `db:"reward_id"   json:"reward_id"`
	IsActive    bool       `db:"is_active"   json:"is_active"`
	CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
}

// JWT context keys
type ContextKey string

const (
	ContextKeyUserID   ContextKey = "user_id"
	ContextKeyFamilyID ContextKey = "family_id"
	ContextKeyRole     ContextKey = "role"
)

// Level thresholds
var LevelThresholds = []struct {
	Level int
	XP    int
	Title string
}{
	{1, 0, "Seedling"},
	{2, 100, "Sprout"},
	{3, 300, "Learner"},
	{4, 600, "Explorer"},
	{5, 1000, "Scholar"},
	{6, 1500, "Hadith Student"},
	{7, 2500, "Quran Reader"},
	{8, 4000, "Prophet Historian"},
	{9, 6000, "Knowledge Seeker"},
	{10, 10000, "Garden Master"},
}

// sql.NullString helpers
var _ = sql.NullString{}
