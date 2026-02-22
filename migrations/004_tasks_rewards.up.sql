CREATE TABLE rewards (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id   UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT,
  value       NUMERIC(10,2) NOT NULL DEFAULT 0,
  type        TEXT NOT NULL DEFAULT 'virtual' CHECK (type IN ('virtual', 'real', 'monetary')),
  icon        TEXT,
  created_by  UUID NOT NULL REFERENCES users(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tasks (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id    UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  title        TEXT NOT NULL,
  description  TEXT,
  assigned_to  UUID NOT NULL REFERENCES users(id),
  created_by   UUID NOT NULL REFERENCES users(id),
  reward_id    UUID REFERENCES rewards(id),
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
    'pending', 'in_progress', 'completed',
    'reward_requested', 'reward_approved', 'reward_declined'
  )),
  due_date     DATE,
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_family   ON tasks(family_id);
CREATE INDEX idx_tasks_assigned ON tasks(assigned_to);
CREATE INDEX idx_tasks_status   ON tasks(status);
