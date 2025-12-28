package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitializeDBInstance(ctx context.Context, master DBConfig, slaves *[]DBConfig, nrApp *newrelic.Application) *DbCluster {
	db := getDbInstance(ctx, master, slaves)

	if nrApp != nil {
		setupNewRelicMonitoring(db, nrApp)
	}

	return db
}

func getDbInstance(ctx context.Context, master DBConfig, slaves *[]DBConfig) (instance *DbCluster) {
	slavesCount := len(*slaves)
	instance = &DbCluster{
		master: new(Connection),
		slaves: make([]*Connection, slavesCount),
	}

	instance.master = initDbConnection(ctx, master)

	for i := 0; i < slavesCount; i++ {
		instance.slaves[i] = initDbConnection(ctx, (*slaves)[i])
	}
	return
}

func initDbConnection(ctx context.Context, config DBConfig) *Connection {
	gormLogger := logger.Default
	if config.DebugMode {
		gormLogger = gormLogger.LogMode(logger.Info)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.Username, config.Password, config.Dbname,
	)

	gormDB, err := gorm.Open(postgres.Dialector{
		Config: &postgres.Config{
			DSN: dsn,
		},
	}, &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: config.SkipDefaultTransaction,
		PrepareStmt:            config.PrepareStmt,
	})
	if err != nil {
		panic("Unable to make gorm connection")
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		panic("Unable to get sqlDB from gormDB")
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(config.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	retries := 3
	for retries > 0 {
		err = sqlDB.Ping()
		if err != nil {
			log.Printf("Unable to ping database server: %s, waiting 2 seconds before trying %d more times\n", err.Error(), retries)
			time.Sleep(time.Second * 2)
			retries--
		} else {
			err = nil
			break
		}
	}
	if err != nil {
		panic("Db not initialised")
	}

	EnsureSchema(ctx, sqlDB)

	conn := Connection{db: gormDB, config: config}
	return &conn
}

func setupNewRelicMonitoring(db *DbCluster, app *newrelic.Application) {
	setupConnectionMonitoring(db.master.db, app)

	for _, slave := range db.slaves {
		setupConnectionMonitoring(slave.db, app)
	}
}

func setupConnectionMonitoring(gormDB *gorm.DB, app *newrelic.Application) {
	gormDB.Callback().Query().Before("gorm:query").Register("newrelic:query_before", func(db *gorm.DB) {
		startDBSegment(db, "SELECT")
	})
	gormDB.Callback().Query().After("gorm:after_query").Register("newrelic:query_after", func(db *gorm.DB) {
		endDBSegment(db)
	})

	gormDB.Callback().Create().Before("gorm:before_create").Register("newrelic:create_before", func(db *gorm.DB) {
		startDBSegment(db, "INSERT")
	})
	gormDB.Callback().Create().After("gorm:after_create").Register("newrelic:create_after", func(db *gorm.DB) {
		endDBSegment(db)
	})

	gormDB.Callback().Update().Before("gorm:before_update").Register("newrelic:update_before", func(db *gorm.DB) {
		startDBSegment(db, "UPDATE")
	})
	gormDB.Callback().Update().After("gorm:after_update").Register("newrelic:update_after", func(db *gorm.DB) {
		endDBSegment(db)
	})

	gormDB.Callback().Delete().Before("gorm:before_delete").Register("newrelic:delete_before", func(db *gorm.DB) {
		startDBSegment(db, "DELETE")
	})
	gormDB.Callback().Delete().After("gorm:after_delete").Register("newrelic:delete_after", func(db *gorm.DB) {
		endDBSegment(db)
	})
}

func startDBSegment(db *gorm.DB, operation string) {
	txn := newrelic.FromContext(db.Statement.Context)
	if txn == nil {
		return
	}

	segment := txn.StartSegment("db." + operation)
	segment.AddAttribute("db.operation", operation)
	segment.AddAttribute("db.table", db.Statement.Table)

	db.Statement.Context = context.WithValue(db.Statement.Context, "nr_segment", segment)
}

func endDBSegment(db *gorm.DB) {
	segment, ok := db.Statement.Context.Value("nr_segment").(*newrelic.Segment)
	if !ok || segment == nil {
		return
	}

	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		txn := newrelic.FromContext(db.Statement.Context)
		if txn != nil {
			txn.NoticeError(db.Error)
		}
	}

	segment.End()
}

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	queries := []string{
		// users table
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// game_sessions table
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			score INT NOT NULL,
			game_mode VARCHAR(50) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// leaderboard table
		`CREATE TABLE IF NOT EXISTS leaderboard (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			total_score INT NOT NULL,
			rank INT
		);`,

		// Add unique constraint on user_id (critical for ON CONFLICT to work)
		`DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint 
				WHERE conname = 'leaderboard_user_id_unique'
			) THEN
				ALTER TABLE leaderboard 
				ADD CONSTRAINT leaderboard_user_id_unique UNIQUE (user_id);
			END IF;
		END $$;`,

		// indexes for leaderboard
		`CREATE INDEX IF NOT EXISTS idx_leaderboard_user_id ON leaderboard(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_leaderboard_rank ON leaderboard(rank);`,
		`CREATE INDEX IF NOT EXISTS idx_leaderboard_total_score ON leaderboard(total_score DESC);`,

		// indexes for game_sessions
		`CREATE INDEX IF NOT EXISTS idx_game_sessions_user_id ON game_sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_game_sessions_timestamp ON game_sessions(timestamp DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_game_sessions_score ON game_sessions(score DESC);`,

		// indexes for users
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`,
	}

	for i, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			log.Printf("schema init failed at query %d: %v", i+1, err)
			return err
		}
	}

	log.Println("database schema ensured successfully")
	return nil
}
