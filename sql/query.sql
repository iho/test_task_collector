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
        MAX(CASE WHEN type = 'V' THEN avg_val END)::DOUBLE PRECISION AS v_val,
        MAX(CASE WHEN type = 'R' THEN avg_val END)::DOUBLE PRECISION AS r_val
    FROM bucketed_measurements
    GROUP BY 1, 2
)
SELECT
    room,
    bucket AS timestamp,
    (CASE 
        WHEN r_val = 0 THEN 0 
        WHEN v_val IS NULL OR r_val IS NULL THEN 0
        ELSE v_val / r_val 
    END)::DOUBLE PRECISION AS i_val,
    v_val,
    r_val
FROM pivoted
ORDER BY room, bucket;
