CREATE TABLE hadiths (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  text_en    TEXT NOT NULL,
  text_ar    TEXT,
  source     TEXT NOT NULL,
  topic      TEXT,
  difficulty TEXT NOT NULL DEFAULT 'medium' CHECK (difficulty IN ('easy', 'medium', 'hard')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE prophets (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name_en       TEXT NOT NULL,
  name_ar       TEXT,
  order_num     INT UNIQUE,
  story_summary TEXT NOT NULL,
  key_miracles  TEXT,
  nation        TEXT,
  quran_refs    TEXT,
  difficulty    TEXT NOT NULL DEFAULT 'medium' CHECK (difficulty IN ('easy', 'medium', 'hard')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quran_verses (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  surah_number     INT NOT NULL,
  ayah_number      INT NOT NULL,
  surah_name_en    TEXT NOT NULL,
  text_ar          TEXT NOT NULL,
  text_en          TEXT NOT NULL,
  transliteration  TEXT,
  tafsir_simple    TEXT NOT NULL,
  life_application TEXT,
  topic            TEXT,
  difficulty       TEXT NOT NULL DEFAULT 'easy' CHECK (difficulty IN ('easy', 'medium', 'hard')),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(surah_number, ayah_number)
);
