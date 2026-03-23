CREATE TABLE IF NOT EXISTS devices (
    id          SERIAL PRIMARY KEY,
    hostname    TEXT NOT NULL,
    ip          TEXT NOT NULL,
    location    TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS configs (
    id          SERIAL PRIMARY KEY,
    device_id   INTEGER NOT NULL REFERENCES devices(id),
    config_text TEXT NOT NULL,
    version     TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS logs (
    id          SERIAL PRIMARY KEY,
    device_id   INTEGER NOT NULL REFERENCES devices(id),
    level       TEXT NOT NULL DEFAULT 'info',  -- info, warning, error
    message     TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS devices_active ON devices(is_active);
CREATE INDEX IF NOT EXISTS configs_device ON configs(device_id);
CREATE INDEX IF NOT EXISTS logs_device_ts ON logs(device_id, created_at DESC);
