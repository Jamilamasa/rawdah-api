CREATE TABLE game_sessions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  family_id        UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  game_name        TEXT NOT NULL,
  game_type        TEXT NOT NULL CHECK (game_type IN ('islamic', 'general')),
  started_at       TIMESTAMPTZ NOT NULL,
  ended_at         TIMESTAMPTZ,
  duration_seconds INT
);

CREATE INDEX idx_game_sessions_user   ON game_sessions(user_id);
CREATE INDEX idx_game_sessions_family ON game_sessions(family_id);

CREATE TABLE notifications (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type       TEXT NOT NULL,
  title      TEXT NOT NULL,
  body       TEXT NOT NULL,
  data       JSONB,
  is_read    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user   ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;

CREATE TABLE push_subscriptions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  endpoint   TEXT NOT NULL UNIQUE,
  p256dh     TEXT NOT NULL,
  auth       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE learn_content (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id    UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  assigned_to  UUID REFERENCES users(id),
  title        TEXT NOT NULL,
  content_type TEXT NOT NULL CHECK (content_type IN ('text', 'link', 'video', 'pdf')),
  content      TEXT NOT NULL,
  reward_id    UUID REFERENCES rewards(id),
  created_by   UUID NOT NULL REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE learn_progress (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content_id   UUID NOT NULL REFERENCES learn_content(id) ON DELETE CASCADE,
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  completed_at TIMESTAMPTZ,
  UNIQUE(content_id, user_id)
);
