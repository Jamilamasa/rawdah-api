CREATE TABLE topic_quizzes (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id      UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  assigned_to    UUID NOT NULL REFERENCES users(id),
  assigned_by    UUID NOT NULL REFERENCES users(id),
  category       TEXT NOT NULL CHECK (category IN ('hadith', 'quran', 'science', 'fun_facts', 'custom', 'general')),
  topic          TEXT NOT NULL,
  lesson_content TEXT NOT NULL,
  flashcards     JSONB NOT NULL,
  questions      JSONB NOT NULL,
  answers        JSONB,
  score          INT,
  xp_awarded     INT NOT NULL DEFAULT 0,
  status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
  completed_at   TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_topic_quizzes_family   ON topic_quizzes(family_id);
CREATE INDEX idx_topic_quizzes_assigned ON topic_quizzes(assigned_to);
CREATE INDEX idx_topic_quizzes_status   ON topic_quizzes(status);
