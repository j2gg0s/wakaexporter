SELECT 'CREATE DATABASE wakatime'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wakatime')\gexec

\c wakatime

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS heartbeat (
    id VARCHAR NOT NULL,

    user_id VARCHAR NOT NULL,
    project VARCHAR,
    language VARCHAR,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    branch VARCHAR,
    category VARCHAR,
    cursorpos VARCHAR,
    dependencies VARCHAR [],
    entity VARCHAR,
    is_write BOOLEAN,
    lineno INTEGER,
    lines INTEGER,
    machine_name_id VARCHAR,
    time NUMERIC,
    type VARCHAR,
    user_agent_id VARCHAR,

    PRIMARY KEY (id, user_id)
);

SELECT 'CREATE TRIGGER heartbeat_set_timestamp
BEFORE UPDATE ON heartbeat
FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp()'
WHERE NOT EXISTS (
    SELECT FROM information_schema.triggers where trigger_name = 'heartbeat_set_timestamp'
)\gexec

CREATE INDEX IF NOT EXISTS idx_heartbeat_created_user_project ON heartbeat (created_at, user_id, project);
CREATE INDEX IF NOT EXISTS idx_heartbeat_created_user_language ON heartbeat (created_at, user_id, language);
