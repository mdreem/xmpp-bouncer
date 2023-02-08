package persistence

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
	"xmpp-bouncer/logger"
)

type fileWriter struct {
}

type dbClient struct {
	db *sql.DB
}

func NewFileWriter() ChatWriter {
	return fileWriter{}
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

func (client fileWriter) Write(timestamp time.Time, ID string, from string, subject string, body string) error {
	f, err := os.OpenFile("messages.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file to write: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			logger.Sugar.Errorw("error closing file file", "error", err)
		}
	}()

	hashString := createHashString(from, subject, body)

	_, err = f.WriteString(fmt.Sprintf("%s;%s;%s;%s;%s;%s\n", timestamp, ID, hashString, from, subject, body))
	if err != nil {
		return fmt.Errorf("error writing message to file: %w", err)
	}
	return nil
}

func createHashString(from string, subject string, body string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%s", from, subject, body)))
	hashString := fmt.Sprintf("%x", hash[:])
	return hashString
}

func (client dbClient) Write(timestamp time.Time, ID string, from string, subject string, body string) error {
	stmt, err := client.db.Prepare("INSERT INTO chats(msg_timestamp, msg_id, from_address, subject, body) VALUES(?,?,?,?,?)")
	if err != nil {
		return fmt.Errorf("unable to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(timestamp, ID, from, subject, body)
	if err != nil {
		return fmt.Errorf("unable to write to the database: %w", err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("unable to retrieve lastId: %w", err)
	}

	logger.Sugar.Debugw("inserted chat line", "lastId", lastId)
	return nil
}
