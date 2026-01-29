package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iho/test_task_collector/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	dbURL := "postgres://user:password@localhost:5432/telemetry?sslmode=disable"

	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 30; i++ {
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
		}
		log.Println("Waiting for Postgres...")
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	schemaBytes, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS measurements CASCADE; DROP TABLE IF EXISTS sensors CASCADE; DROP TABLE IF EXISTS rooms CASCADE; DROP TYPE IF EXISTS sensor_type; DROP DOMAIN IF EXISTS nullable_float8 CASCADE;")
	if err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}

	_, err = pool.Exec(ctx, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	log.Println("Schema applied.")

	roomA, err := queries.InsertRoom(ctx, "room_A")
	if err != nil {
		return err
	}

	sA_V1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomA.ID, Name: "sA_V1", Type: db.SensorTypeV})
	if err != nil {
		return err
	}
	sA_R1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomA.ID, Name: "sA_R1", Type: db.SensorTypeR})
	if err != nil {
		return err
	}
	sA_R2, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomA.ID, Name: "sA_R2", Type: db.SensorTypeR})
	if err != nil {
		return err
	}

	// Room B
	roomB, err := queries.InsertRoom(ctx, "room_B")
	if err != nil {
		return err
	}
	// 2 sensors for V, 3 sensors for R
	sB_V1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_V1", Type: db.SensorTypeV})
	if err != nil {
		return err
	}
	sB_V2, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_V2", Type: db.SensorTypeV})
	if err != nil {
		return err
	}
	sB_R1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_R1", Type: db.SensorTypeR})
	if err != nil {
		return err
	}
	sB_R2, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_R2", Type: db.SensorTypeR})
	if err != nil {
		return err
	}
	sB_R3, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_R3", Type: db.SensorTypeR})
	if err != nil {
		return err
	}

	baseTime := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)

	insertM := func(sensorID int32, val float64, t time.Time) {
		_, err := queries.InsertMeasurement(ctx, db.InsertMeasurementParams{
			SensorID:  sensorID,
			Value:     val,
			Timestamp: pgtype.Timestamp{Time: t, Valid: true},
		})
		if err != nil {
			log.Printf("Failed to insert measurement: %v", err)
		}
	}

	ts1 := baseTime
	insertM(sA_V1.ID, 220, ts1)
	insertM(sA_R1.ID, 10, ts1)
	insertM(sA_R2.ID, 12, ts1)

	ts2 := baseTime.Add(1 * time.Second)
	insertM(sB_V1.ID, 220, ts2)
	insertM(sB_V2.ID, 230, ts2)
	insertM(sB_R1.ID, 100, ts2)
	insertM(sB_R2.ID, 150, ts2)
	insertM(sB_R3.ID, 50, ts2)

	ts3 := baseTime.Add(2 * time.Second)
	insertM(sA_V1.ID, 230, ts3)
	insertM(sA_R1.ID, 11.5, ts3)

	ts4 := baseTime.Add(3 * time.Second)
	insertM(sA_V1.ID, 240, ts4)

	roomC, err := queries.InsertRoom(ctx, "room_C")
	if err != nil {
		return err
	}

	sC_V1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomC.ID, Name: "sC_V1", Type: db.SensorTypeV})
	if err != nil {
		return err
	}

	insertM(sC_V1.ID, 300, ts1)

	results, err := queries.GetAnalysis(ctx)
	if err != nil {
		return fmt.Errorf("get analysis: %w", err)
	}

	fmt.Printf("%-10s | %-30s | %-10s | %-10s | %-10s\n", "ROOM", "TIMESTAMP", "I", "V", "R")
	fmt.Println("------------------------------------------------------------------------------------------")
	for _, r := range results {
		tStr := r.Timestamp.Time.Format(time.RFC3339)

		fmt.Printf("%-10s | %-30s | %-10s | %-10s | %-10s\n",
			r.Room,
			tStr,
			getVal(r.IVal),
			getVal(r.VVal),
			getVal(r.RVal),
		)
	}

	return nil
}

func getVal(f pgtype.Float8) string {
	if !f.Valid {
		return "NULL"
	}
	return fmt.Sprintf("%.2f", f.Float64)
}
