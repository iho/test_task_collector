// Package main implements the Task 2 verification tool.
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

	pool, err := initDB(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := db.New(pool)

	if err := applySchema(ctx, pool); err != nil {
		return err
	}

	if err := seedData(ctx, queries); err != nil {
		return err
	}

	return printAnalysis(ctx, queries)
}

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	dbURL := "postgres://user:password@localhost:5432/telemetry?sslmode=disable"
	var pool *pgxpool.Pool
	var err error

	// Retry loop for Docker startup
	for i := 0; i < 30; i++ {
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}
		}
		log.Println("Waiting for Postgres...")
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to db: %w", err)
}

func applySchema(ctx context.Context, pool *pgxpool.Pool) error {
	schemaBytes, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	// Drop artifacts to ensure clean state
	cleanupSQL := `
		DROP TABLE IF EXISTS measurements CASCADE;
		DROP TABLE IF EXISTS sensors CASCADE;
		DROP TABLE IF EXISTS rooms CASCADE;
		DROP TYPE IF EXISTS sensor_type;
		DROP DOMAIN IF EXISTS nullable_float8 CASCADE;
	`
	if _, err := pool.Exec(ctx, cleanupSQL); err != nil {
		return fmt.Errorf("drop tables: %w", err)
	}

	if _, err := pool.Exec(ctx, string(schemaBytes)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	log.Println("Schema applied.")
	return nil
}

func seedData(ctx context.Context, queries *db.Queries) error {
	// Helper for insertions
	insertM := func(sensorID int32, val float64, t time.Time) error {
		_, err := queries.InsertMeasurement(ctx, db.InsertMeasurementParams{
			SensorID:  sensorID,
			Value:     val,
			Timestamp: pgtype.Timestamp{Time: t, Valid: true},
		})
		return err
	}

	// --- Room A ---
	roomA, err := queries.InsertRoom(ctx, "room_A")
	if err != nil {
		return err
	}

	sAV1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomA.ID, Name: "sA_V1", Type: db.SensorTypeV})
	if err != nil {
		return err
	}
	sAR1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomA.ID, Name: "sA_R1", Type: db.SensorTypeR})
	if err != nil {
		return err
	}
	sAR2, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomA.ID, Name: "sA_R2", Type: db.SensorTypeR})
	if err != nil {
		return err
	}

	// --- Room B ---
	roomB, err := queries.InsertRoom(ctx, "room_B")
	if err != nil {
		return err
	}

	sBV1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_V1", Type: db.SensorTypeV})
	if err != nil {
		return err
	}
	sBV2, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_V2", Type: db.SensorTypeV})
	if err != nil {
		return err
	}
	sBR1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_R1", Type: db.SensorTypeR})
	if err != nil {
		return err
	}
	sBR2, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_R2", Type: db.SensorTypeR})
	if err != nil {
		return err
	}
	sBR3, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomB.ID, Name: "sB_R3", Type: db.SensorTypeR})
	if err != nil {
		return err
	}

	// --- Room C ---
	roomC, err := queries.InsertRoom(ctx, "room_C")
	if err != nil {
		return err
	}
	sCV1, err := queries.InsertSensor(ctx, db.InsertSensorParams{RoomID: roomC.ID, Name: "sC_V1", Type: db.SensorTypeV})
	if err != nil {
		return err
	}

	baseTime := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)

	// Case 1: Room A, T+0
	if err := insertM(sAV1.ID, 220, baseTime); err != nil {
		return err
	}
	if err := insertM(sAR1.ID, 10, baseTime); err != nil {
		return err
	}
	if err := insertM(sAR2.ID, 12, baseTime); err != nil {
		return err
	}

	// Case 2: Room B, T+1
	ts2 := baseTime.Add(1 * time.Second)
	if err := insertM(sBV1.ID, 220, ts2); err != nil {
		return err
	}
	if err := insertM(sBV2.ID, 230, ts2); err != nil {
		return err
	}
	if err := insertM(sBR1.ID, 100, ts2); err != nil {
		return err
	}
	if err := insertM(sBR2.ID, 150, ts2); err != nil {
		return err
	}
	if err := insertM(sBR3.ID, 50, ts2); err != nil {
		return err
	}

	// Case 3: Room A, T+2
	ts3 := baseTime.Add(2 * time.Second)
	if err := insertM(sAV1.ID, 230, ts3); err != nil {
		return err
	}
	if err := insertM(sAR1.ID, 11.5, ts3); err != nil {
		return err
	}

	// Case 4: Gap Filling (Room A, T+3)
	ts4 := baseTime.Add(3 * time.Second)
	if err := insertM(sAV1.ID, 240, ts4); err != nil {
		return err
	}

	// Case 5: Partial Data (Room C, T+0)
	if err := insertM(sCV1.ID, 300, baseTime); err != nil {
		return err
	}

	return nil
}

func printAnalysis(ctx context.Context, queries *db.Queries) error {
	results, err := queries.GetAnalysis(ctx)
	if err != nil {
		return fmt.Errorf("get analysis: %w", err)
	}

	fmt.Printf("%-10s | %-30s | %-10s | %-10s | %-10s\n", "ROOM", "TIMESTAMP", "I", "V", "R")
	fmt.Println("------------------------------------------------------------------------------------------")
	for _, r := range results {
		fmt.Printf("%-10s | %-30s | %-10s | %-10s | %-10s\n",
			r.Room,
			r.Timestamp.Time.Format(time.RFC3339),
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
