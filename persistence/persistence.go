package persistence

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"xmpp-bouncer/logger"
)

type dbClient struct {
	db *sql.DB
}

func NewDBWriter(connectionString string) ChatWriter {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)

	return dbClient{
		db: db,
	}
}

type ChatWriter interface {
	Write(timestamp time.Time, ID string, from string, subject string, body string) error
}

func createHashString(from string, subject string, body string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", from, subject, body)))
	hashString := fmt.Sprintf("%x", hash[:])
	return hashString
}

func (client dbClient) Write(timestamp time.Time, ID string, from string, subject string, body string) error {
	stmt, err := client.db.Prepare("INSERT INTO chats(msg_timestamp, msg_id, msg_hash, from_address, subject, body) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return fmt.Errorf("unable to prepare statement: %w", err)
	}
	defer stmt.Close()

	hashString := createHashString(from, subject, body)
	res, err := stmt.Exec(timestamp, ID, hashString, from, subject, body)
	if err != nil {
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number == 1062 {
				logger.Sugar.Debugw("ignoring duplicate entry", "msg_id", ID, "msg_hash", hashString)
				return nil
			}
		}
		return fmt.Errorf("unable to write to the database: %w", err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("unable to retrieve lastId: %w", err)
	}

	logger.Sugar.Debugw("inserted chat line", "lastId", lastId)
	return nil
}
