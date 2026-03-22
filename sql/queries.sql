
-- Активные устройства + количество конфигов

SELECT
    d.id,
    d.hostname,
    d.ip,
    d.location,
    COUNT(c.id) AS config_count
FROM devices d
LEFT JOIN configs c ON c.device_id = d.id
WHERE d.is_active = 1
GROUP BY d.id
ORDER BY config_count DESC;

-- Последние N логов для устройства

SELECT
    l.id,
    l.level,
    l.message,
    l.created_at
FROM logs l
WHERE l.device_id = ?
ORDER BY l.created_at DESC
LIMIT ?;


-- Тяжёлый запрос — все устройства + дата последнего лога + количество конфигов

SELECT
    d.id,
    d.hostname,
    d.ip,
    d.is_active,
    COUNT(DISTINCT c.id)  AS config_count,
    MAX(l.created_at)     AS last_log_time
FROM devices d
LEFT JOIN configs c ON c.device_id = d.id
LEFT JOIN logs l    ON l.device_id = d.id
GROUP BY d.id
ORDER BY last_log_time DESC;

-- проблема: два JOIN дают декартово произведение, COUNT(DISTINCT) тяжёлый
-- оптимизация — агрегируем до джоина:

SELECT
    d.id, d.hostname, d.ip, d.is_active,
    COALESCE(cc.cnt, 0) AS config_count,
    ll.last_log
FROM devices d
LEFT JOIN (
    SELECT device_id, COUNT(*) AS cnt
    FROM configs GROUP BY device_id
) cc ON cc.device_id = d.id
LEFT JOIN (
    SELECT device_id, MAX(created_at) AS last_log
    FROM logs GROUP BY device_id
) ll ON ll.device_id = d.id
ORDER BY ll.last_log DESC;


-- Индексы

CREATE INDEX configs_device ON configs(device_id);
CREATE INDEX logs_device_ts ON logs(device_id, created_at DESC);
CREATE INDEX devices_active ON devices(is_active);

-- configs_device — ускоряет JOIN по device_id, без него full scan
-- logs_device_ts — составной, покрывает фильтр + сортировку по дате
-- devices_active — ускоряет WHERE is_active = 1 если активных мало
