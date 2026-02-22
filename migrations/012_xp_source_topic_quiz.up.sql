DO $$
DECLARE
  source_check_name TEXT;
BEGIN
  SELECT c.conname
    INTO source_check_name
    FROM pg_constraint c
   WHERE c.conrelid = 'xp_events'::regclass
     AND c.contype = 'c'
     AND pg_get_constraintdef(c.oid) ILIKE '%source IN%';

  IF source_check_name IS NOT NULL THEN
    EXECUTE format('ALTER TABLE xp_events DROP CONSTRAINT %I', source_check_name);
  END IF;
END $$;

ALTER TABLE xp_events
  ADD CONSTRAINT xp_events_source_check
  CHECK (
    source IN (
      'task',
      'hadith_quiz',
      'prophet_quiz',
      'quran_quiz',
      'topic_quiz',
      'quran_lesson',
      'game'
    )
  );
