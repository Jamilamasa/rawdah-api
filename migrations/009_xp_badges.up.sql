CREATE TABLE user_xp (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  family_id  UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  total_xp   INT NOT NULL DEFAULT 0,
  level      INT NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE xp_events (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  family_id  UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  source     TEXT NOT NULL CHECK (source IN ('task', 'hadith_quiz', 'prophet_quiz', 'quran_quiz', 'quran_lesson', 'game')),
  source_id  UUID NOT NULL,
  xp_amount  INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE badges (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug        TEXT UNIQUE NOT NULL,
  name        TEXT NOT NULL,
  description TEXT NOT NULL,
  icon        TEXT NOT NULL,
  xp_reward   INT NOT NULL DEFAULT 0
);

CREATE TABLE user_badges (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  badge_id   UUID NOT NULL REFERENCES badges(id),
  awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, badge_id)
);

INSERT INTO badges (slug, name, description, icon, xp_reward) VALUES
  ('first_task',       'First Step',       'Completed your first task',            '✅', 10),
  ('streak_7',         '7-Day Streak',     'Active 7 days in a row',              '🔥', 50),
  ('streak_30',        '30-Day Streak',    'Active 30 days in a row',             '💎', 200),
  ('first_quiz',       'Quiz Starter',     'Completed your first quiz',            '🧠', 20),
  ('hadith_scholar',   'Hadith Scholar',   'Completed 10 hadith quizzes',          '📜', 100),
  ('prophet_explorer', 'Prophet Explorer', 'Completed quizzes on 10 prophets',     '🌟', 100),
  ('quran_learner',    'Quran Learner',    'Completed 5 Quran lessons',            '📖', 80),
  ('perfect_score',    'Perfect Score',    'Got 100% on any quiz',                 '⭐', 30),
  ('game_player',      'Game Player',      'Played your first game',               '🎮', 10),
  ('messenger',        'Messenger',        'Sent your first message',              '💬', 10);
