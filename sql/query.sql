-- name: InsertRoom :one
INSERT INTO rooms (name) VALUES ($1) RETURNING *;

-- name: GetRoom :one
SELECT * FROM rooms WHERE name = $1 LIMIT 1;

-- name: InsertSensor :one
INSERT INTO sensors (room_id, name, type) VALUES ($1, $2, $3) RETURNING *;

-- name: GetSensor :one
SELECT * FROM sensors WHERE name = $1 LIMIT 1;

-- name: InsertMeasurement :one
INSERT INTO measurements (sensor_id, value, timestamp) VALUES ($1, $2, $3) RETURNING *;

-- name: GetAnalysis :many
WITH bucketed_measurements AS (
    SELECT
        r.name AS room,
        date_trunc('second', m.timestamp)::TIMESTAMP AS bucket,
        s.type,
        AVG(m.value)::DOUBLE PRECISION AS avg_val
    FROM measurements m
    JOIN sensors s ON m.sensor_id = s.id
    JOIN rooms r ON s.room_id = r.id
    GROUP BY 1, 2, 3
),
pivoted AS (
    SELECT
        room,
        bucket,
        MAX(CASE WHEN type = 'V' THEN avg_val END)::DOUBLE PRECISION AS v_raw,
        MAX(CASE WHEN type = 'R' THEN avg_val END)::DOUBLE PRECISION AS r_raw
    FROM bucketed_measurements
    GROUP BY 1, 2
),
filled AS (
    SELECT
        room,
        bucket,
        COALESCE(v_raw, MAX(v_raw) OVER (PARTITION BY room ORDER BY bucket ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)) AS v_val,
        COALESCE(r_raw, MAX(r_raw) OVER (PARTITION BY room ORDER BY bucket ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)) AS r_val
    FROM pivoted
)
SELECT
    room,
    bucket AS timestamp,
    (CASE 
        WHEN r_val = 0 THEN 0 
        WHEN v_val IS NULL OR r_val IS NULL THEN 0
        ELSE v_val / r_val 
    END)::nullable_float8 AS i_val,
    CASE WHEN true THEN v_val ELSE NULL END::nullable_float8 AS v_val,
    CASE WHEN true THEN r_val ELSE NULL END::nullable_float8 AS r_val
FROM filled
ORDER BY room, bucket;
