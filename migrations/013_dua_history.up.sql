CREATE TABLE dua_history (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id      UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  asking_for     TEXT NOT NULL,
  heavy_on_heart TEXT NOT NULL,
  afraid_of      TEXT NOT NULL,
  if_answered    TEXT NOT NULL,
  output_style   TEXT NOT NULL,
  depth          TEXT NOT NULL,
  tone           TEXT NOT NULL,
  selected_names JSONB NOT NULL DEFAULT '[]'::jsonb,
  dua_text       TEXT NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dua_history_user_created ON dua_history(user_id, created_at DESC);
CREATE INDEX idx_dua_history_family_user ON dua_history(family_id, user_id);
