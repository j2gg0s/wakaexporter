SELECT 'CREATE DATABASE wakatime'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wakatime')\gexec

\c wakatime

SET client_min_messages TO WARNING;

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

-- from pg_prometheus plugin
CREATE OR REPLACE FUNCTION insert_view_normal()
    RETURNS TRIGGER LANGUAGE PLPGSQL AS
$BODY$
DECLARE
    metric_labels_id  INTEGER;
    labels_table      NAME;
    values_table      NAME;
BEGIN
    IF TG_NARGS != 2 THEN
        RAISE EXCEPTION 'insert_view_normal requires 2 parameters';
    END IF;

    values_table := TG_ARGV[0];
    labels_table := TG_ARGV[1];

    -- Insert labels
    EXECUTE format('SELECT id FROM %I l WHERE %L = l.labels AND %L = l.metric_name',
          labels_table, New.labels, New.name) INTO metric_labels_id;

    IF metric_labels_id IS NULL THEN
      EXECUTE format(
          $$
          INSERT INTO %I (metric_name, labels) VALUES (%L, %L) RETURNING id
          $$,
          labels_table,
          New.name,
          New.labels
      ) INTO STRICT metric_labels_id;
    END IF;

    EXECUTE format('INSERT INTO %I (time, value, labels_id) VALUES (%L, %L, %L) ON CONFLICT(time, labels_id) DO NOTHING',
          values_table, New.time, New.value, metric_labels_id);

    RETURN NULL;
END
$BODY$;

CREATE OR REPLACE FUNCTION create_prometheus_table(
       metrics_view_name NAME,
       metrics_values_table_name NAME = NULL,
       metrics_labels_table_name NAME = NULL,
       use_timescaledb BOOL = NULL,
       chunk_time_interval INTERVAL = interval '1 day'
)
    RETURNS VOID LANGUAGE PLPGSQL VOLATILE AS
$BODY$
DECLARE
    timescaledb_ext_relid OID = NULL;
BEGIN
    SELECT oid FROM pg_extension
    WHERE extname = 'timescaledb'
    INTO timescaledb_ext_relid;

    IF use_timescaledb IS NULL THEN
      IF timescaledb_ext_relid IS NULL THEN
        use_timescaledb := FALSE;
      ELSE
        use_timescaledb := TRUE;
      END IF;
    END IF;

    IF use_timescaledb AND  timescaledb_ext_relid IS NULL THEN
      RAISE 'TimescaleDB not installed';
    END IF;

    IF metrics_view_name IS NULL THEN
       RAISE EXCEPTION 'Invalid table name';
    END IF;

    IF metrics_values_table_name IS NULL THEN
       metrics_values_table_name := format('%I_values', metrics_view_name);
    END IF;

    IF metrics_labels_table_name IS NULL THEN
       metrics_labels_table_name := format('%I_labels', metrics_view_name);
    END IF;

    -- Create labels table
    EXECUTE format(
        $$
        CREATE TABLE IF NOT EXISTS %I (
              id SERIAL PRIMARY KEY,
              metric_name TEXT NOT NULL,
              labels jsonb,
              UNIQUE(metric_name, labels)
        )
        $$,
        metrics_labels_table_name
    );
    -- Add a GIN index on labels
    EXECUTE format(
        $$
        CREATE INDEX IF NOT EXISTS %I_labels_idx ON %1$I USING GIN (labels)
        $$,
        metrics_labels_table_name
    );

     -- Add a index on metric name
    EXECUTE format(
        $$
        CREATE INDEX IF NOT EXISTS %I_metric_name_idx ON %1$I USING BTREE (metric_name)
        $$,
        metrics_labels_table_name
    );

    -- Create normalized metrics table
    IF use_timescaledb THEN
      --does not support foreign  references
      EXECUTE format(
          $$
          CREATE TABLE IF NOT EXISTS %I (time TIMESTAMPTZ, value FLOAT8, labels_id INTEGER, UNIQUE(time, labels_id))
          $$,
          metrics_values_table_name
      );
    ELSE
      EXECUTE format(
          $$
          CREATE TABLE IF NOT EXISTS %I (time TIMESTAMPTZ, value FLOAT8, labels_id INTEGER REFERENCES %I(id))
          $$,
          metrics_values_table_name,
          metrics_labels_table_name
      );
    END IF;

    -- Make metrics table a hypertable if the TimescaleDB extension is present
    IF use_timescaledb THEN
       PERFORM create_hypertable(metrics_values_table_name::regclass, 'time',
               chunk_time_interval => _timescaledb_internal.interval_to_usec(chunk_time_interval));
    END IF;

    -- Create labels ID column index
    EXECUTE format(
        $$
        CREATE INDEX IF NOT EXISTS %I_labels_id_idx ON %1$I USING BTREE (labels_id, time desc)
        $$,
        metrics_values_table_name
    );

    -- Create a view for the metrics
    EXECUTE format(
        $$
        CREATE VIEW %I AS 
        SELECT m.time AS time, l.metric_name AS name,  m.value AS value, l.labels AS labels
        FROM %I AS m
        INNER JOIN %I l ON (m.labels_id = l.id)
        $$,
        metrics_view_name,
        metrics_values_table_name,
        metrics_labels_table_name
    );

    EXECUTE format(
        $$
        DROP TRIGGER IF EXISTS insert_trigger ON public.%I
        $$,
        metrics_view_name
    );

    EXECUTE format(
        $$
        CREATE TRIGGER insert_trigger INSTEAD OF INSERT ON %I
        FOR EACH ROW EXECUTE PROCEDURE insert_view_normal(%L, %L)
        $$,
        metrics_view_name,
        metrics_values_table_name,
        metrics_labels_table_name
    );
END
$BODY$;
-- from pg_prometheus plugin

SELECT create_prometheus_table('metric')
WHERE NOT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'metric_view');
