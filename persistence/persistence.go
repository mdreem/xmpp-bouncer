package persistence

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"
	"xmpp-bouncer/logger"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migrate_mysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file" // needed to load migration files
)

type dbClient struct {
	db *sql.DB
}

func NewDBWriter(connectionString string) MigrateableChatWriter {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		logger.Sugar.Fatalw("sql open failed", "error", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)

	stmt, err := db.Prepare("SELECT 1 from dual")
	if err != nil {
		logger.Sugar.Fatalw("initial sql prepare failed", "error", err)
	}
	_, err = stmt.Exec()
	if err != nil {
		logger.Sugar.Fatalw("initial sql exec failed", "error", err)
	}

	return dbClient{
		db: db,
	}
}

func (client dbClient) Migrate(location string) error {
	driver, _ := migrate_mysql.WithInstance(client.db, &migrate_mysql.Config{})
	migrationsLocation := fmt.Sprintf("file://%s", location)
	logger.Sugar.Infow("migrations are located here", "location", migrationsLocation)

	migration, err := migrate.NewWithDatabaseInstance(
		migrationsLocation,
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("unable to initialize migrations: %w", err)
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("unable to migrate: %w", err)
	}
	if err == migrate.ErrNoChange {
		logger.Sugar.Info("no change during migration")
	}
	return nil
}

type MigrateableChatWriter interface {
	ChatWriter
	Migrate(location string) error
}

type ChatWriter interface {
	Write(timestamp time.Time, id string, from string, subject string, body string) error
}

func createHashString(from string, subject string, body string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", from, subject, body)))
	hashString := fmt.Sprintf("%x", hash[:])
	return hashString
}

func (client dbClient) Write(timestamp time.Time, id string, from string, subject string, body string) error {
	stmt, err := client.db.Prepare("INSERT INTO chats(msg_timestamp, msg_id, msg_hash, from_address, subject, body) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return fmt.Errorf("unable to prepare statement: %w", err)
	}
	defer stmt.Close()

	hashString := createHashString(from, subject, body)
	res, err := stmt.Exec(timestamp, id, hashString, from, subject, body)
	if err != nil {
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number == 1062 {
				logger.Sugar.Debugw("ignoring duplicate entry", "msg_id", id, "msg_hash", hashString)
				return nil
			}
		}
		return fmt.Errorf("unable to write to the database: %w", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("unable to retrieve lastId: %w", err)
	}

	logger.Sugar.Debugw("inserted chat line", "lastId", lastID)
	return nil
}
