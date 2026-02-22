CREATE TABLE hadith_quizzes (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id      UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  hadith_id      UUID NOT NULL REFERENCES hadiths(id),
  assigned_to    UUID NOT NULL REFERENCES users(id),
  assigned_by    UUID NOT NULL REFERENCES users(id),
  questions      JSONB NOT NULL,
  answers        JSONB,
  score          INT,
  xp_awarded     INT NOT NULL DEFAULT 0,
  status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN
    ('pending', 'memorizing', 'in_progress', 'completed')),
  memorize_until TIMESTAMPTZ,
  completed_at   TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE prophet_quizzes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id    UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  prophet_id   UUID NOT NULL REFERENCES prophets(id),
  assigned_to  UUID NOT NULL REFERENCES users(id),
  assigned_by  UUID NOT NULL REFERENCES users(id),
  questions    JSONB NOT NULL,
  answers      JSONB,
  score        INT,
  xp_awarded   INT NOT NULL DEFAULT 0,
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN
    ('pending', 'in_progress', 'completed')),
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quran_lessons (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id    UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  verse_id     UUID NOT NULL REFERENCES quran_verses(id),
  assigned_to  UUID NOT NULL REFERENCES users(id),
  assigned_by  UUID NOT NULL REFERENCES users(id),
  reward_id    UUID REFERENCES rewards(id),
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN
    ('pending', 'reading', 'completed')),
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quran_quizzes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id    UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  verse_id     UUID NOT NULL REFERENCES quran_verses(id),
  lesson_id    UUID REFERENCES quran_lessons(id),
  assigned_to  UUID NOT NULL REFERENCES users(id),
  questions    JSONB NOT NULL,
  answers      JSONB,
  score        INT,
  xp_awarded   INT NOT NULL DEFAULT 0,
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN
    ('pending', 'in_progress', 'completed')),
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hadith_quizzes_family   ON hadith_quizzes(family_id);
CREATE INDEX idx_hadith_quizzes_assigned ON hadith_quizzes(assigned_to);
CREATE INDEX idx_prophet_quizzes_family  ON prophet_quizzes(family_id);
CREATE INDEX idx_quran_lessons_family    ON quran_lessons(family_id);
CREATE INDEX idx_quran_quizzes_family    ON quran_quizzes(family_id);
