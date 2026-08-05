package main

import (
	"fmt"
	"log"
	"mineral/data"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (app *Config) initDB() *gorm.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panic("can't connect to database")
	}

	// Auto-migrate the schema using actual model structs, not interfaces
	if err := conn.AutoMigrate(
		&data.User{},
		&data.Income{},
		&data.Expense{},
		&data.InventoryItem{},
		&data.MineSiteInfo{},
		&data.MineSiteCertification{},
		&data.CoCLot{},
		&data.LotComposition{},
		&data.ExportShipment{},
		&data.DueDiligenceReport{},
		&data.ThirdPartyAudit{},
		&data.ComplianceDocument{},
		// Enhanced traceability models
		&data.GPSLocation{},
		&data.TransportRecord{},
		&data.ProcessingRecord{},
		&data.TrackingAlert{},
		&data.RealTimeTracking{},
		&data.CustodyTransfer{},
		&data.PhotoRecord{},
		&data.Stakeholder{},
	); err != nil {
		log.Panic("failed to migrate database:", err)
	}
	log.Println("Database migration completed successfully")

	// Backfill public verification codes for any lots created before the QR
	// verification feature existed. Idempotent and additive — safe on live data.
	if n, err := data.NewComplianceRepository(conn).BackfillVerifyCodes(); err != nil {
		log.Println("warning: verify-code backfill failed:", err)
	} else if n > 0 {
		log.Printf("Backfilled verification codes for %d existing CoC lot(s)", n)
	}

	return conn
}

func connectToDB() *gorm.DB {
	counts := 0

	// Get database connection details from environment variables or use defaults
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "miningFinanceDB_2025"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "mining_data"
	}

	// Construct the DSN string
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
	}

	log.Printf("Attempting to connect to database with DSN: %s", dsn)

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
			log.Printf("Connection error: %v", err)
		} else {
			log.Print("connected to database!")
			return connection
		}

		if counts > 10 {
			return nil
		}

		log.Print("Backing off for 1 second")
		time.Sleep(1 * time.Second)
		counts++
	}
}

func openDB(dsn string) (*gorm.DB, error) {
	config := &gorm.Config{
		PrepareStmt: true, // cache prepared statements — reduces parse overhead per query
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	// Get the underlying *sql.DB instance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool from environment with sensible defaults.
	// Tune MAX_DB_OPEN_CONNS and MAX_DB_IDLE_CONNS in .env for your server size:
	//   Small VPS (2 CPU):  OPEN=50,  IDLE=10
	//   Medium (4 CPU):     OPEN=100, IDLE=20
	//   Large (8+ CPU):     OPEN=200, IDLE=40
	maxOpen := getEnvInt("MAX_DB_OPEN_CONNS", 200)
	maxIdle := getEnvInt("MAX_DB_IDLE_CONNS", 20)
	connMaxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	connMaxIdleTime := getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// Test the connection
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// getEnvInt reads an integer from the environment with a fallback default.
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

// getEnvDuration reads a duration (in seconds) from the environment with a fallback.
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return defaultVal
}
