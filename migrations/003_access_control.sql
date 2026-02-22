CREATE TABLE family_access_control (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id   UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  grantor_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  grantee_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  permissions JSONB NOT NULL DEFAULT '[]',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(grantor_id, grantee_id)
);
