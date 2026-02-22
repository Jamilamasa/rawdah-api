CREATE TABLE recurring_tasks (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id   UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT,
  assigned_to UUID NOT NULL REFERENCES users(id),
  created_by  UUID NOT NULL REFERENCES users(id),
  reward_id   UUID REFERENCES rewards(id),
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_recurring_tasks_family ON recurring_tasks(family_id);
CREATE INDEX idx_recurring_tasks_active ON recurring_tasks(is_active) WHERE is_active = TRUE;
