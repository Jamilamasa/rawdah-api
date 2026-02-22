CREATE TABLE users (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  family_id          UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  role               TEXT NOT NULL CHECK (role IN ('parent', 'child', 'adult_relative')),
  name               TEXT NOT NULL,
  username           TEXT,
  email              TEXT,
  password_hash      TEXT NOT NULL,
  avatar_url         TEXT,
  theme              TEXT NOT NULL DEFAULT 'forest',
  date_of_birth      DATE,
  game_limit_minutes INT NOT NULL DEFAULT 60,
  is_active          BOOLEAN NOT NULL DEFAULT TRUE,
  created_by         UUID REFERENCES users(id),
  last_login_at      TIMESTAMPTZ,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_family ON users(family_id);
CREATE UNIQUE INDEX idx_users_username_family ON users(family_id, username)
  WHERE username IS NOT NULL;

CREATE TABLE refresh_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
