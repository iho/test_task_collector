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

	// Wait for DB
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

	// Clean references (if any from previous runs) and apply schema
	// For simplicity, we'll just drop tables and re-create.
	// In a real app we'd use migrate, but for this task iterating quickly is key.
	// We read schema.sql manually.
	schemaBytes, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	// Simple split by ; isn't robust for all SQL but works for our simple schema
	// Or just exec the whole blob if driver supports it (pgx usually does).
	// Drop old tables first to ensure clean state
	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS measurements CASCADE; DROP TABLE IF EXISTS sensors CASCADE; DROP TABLE IF EXISTS rooms CASCADE; DROP TYPE IF EXISTS sensor_type;")
	if err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}

	_, err = pool.Exec(ctx, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	log.Println("Schema applied.")

	// Seed Data
	// Room A
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

	// Seed Measurements
	baseTime := time.Date(2025, 6, 30, 10, 0, 0, 0, time.UTC)

	// Helper to generic insert
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

	// CASE 1: Standard data for Room A at T+0s
	ts1 := baseTime
	insertM(sA_V1.ID, 220, ts1)
	insertM(sA_R1.ID, 10, ts1)
	insertM(sA_R2.ID, 12, ts1)

	// CASE 2: Room B at T+1s
	ts2 := baseTime.Add(1 * time.Second)
	insertM(sB_V1.ID, 220, ts2)
	insertM(sB_V2.ID, 230, ts2)
	insertM(sB_R1.ID, 100, ts2)
	insertM(sB_R2.ID, 150, ts2)
	insertM(sB_R3.ID, 50, ts2)

	// CASE 3: Room A at T+2s
	ts3 := baseTime.Add(2 * time.Second)
	insertM(sA_V1.ID, 230, ts3)
	insertM(sA_R1.ID, 11.5, ts3)

	// Run Analysis
	results, err := queries.GetAnalysis(ctx)
	if err != nil {
		return fmt.Errorf("get analysis: %w", err)
	}

	// Print Header
	fmt.Printf("%-10s | %-30s | %-10s | %-10s | %-10s\n", "ROOM", "TIMESTAMP", "I", "V", "R")
	fmt.Println("------------------------------------------------------------------------------------------")
	for _, r := range results {
		tStr := r.Timestamp.Time.Format(time.RFC3339)

		fmt.Printf("%-10s | %-30s | %-10.4f | %-10.2f | %-10.2f\n",
			r.Room,
			tStr,
			r.IVal,
			r.VVal,
			r.RVal,
		)
	}

	return nil
}
