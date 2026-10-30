-- Existing TIMESTAMP values were written in the server's default time zone (UTC).
ALTER TABLE members
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
UPDATE members SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE members SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;
ALTER TABLE members
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE devices
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
UPDATE devices SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE devices SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;
ALTER TABLE devices
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE scan_runs
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
UPDATE scan_runs SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE scan_runs SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;
ALTER TABLE scan_runs
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE member_attendances
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
UPDATE member_attendances SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE member_attendances SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;
ALTER TABLE member_attendances
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE device_observations
  ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
  ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
UPDATE device_observations SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;
UPDATE device_observations SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;
ALTER TABLE device_observations
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL;
