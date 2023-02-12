package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mysqlContainer struct {
	testcontainers.Container
}

func setupMysql(ctx context.Context) (*mysqlContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8",
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "database",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("port: 3306  MySQL Community Server - GPL"),
			wait.ForListeningPort("3306/tcp"),
		),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &mysqlContainer{Container: container}, nil
}

type chatLine struct {
	msgTimestamp time.Time
	msgId        string
	fromAddress  string
	subject      string
	body         string
}

func queryData(connectionString string) ([]chatLine, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return []chatLine{}, fmt.Errorf("unable to open connection: %w", err)
	}
	defer db.Close()

	results, err := db.Query("SELECT msg_timestamp, msg_id, from_address, subject, body FROM chats")
	if err != nil {
		return []chatLine{}, fmt.Errorf("unable to query data: %w", err)
	}

	chatLines := make([]chatLine, 0)
	for results.Next() {
		var curChatLine chatLine
		err = results.Scan(&curChatLine.msgTimestamp, &curChatLine.msgId, &curChatLine.fromAddress, &curChatLine.subject, &curChatLine.body)
		if err != nil {
			return []chatLine{}, fmt.Errorf("unable to scan data: %w", err)
		}

		chatLines = append(chatLines, curChatLine)
	}

	return chatLines, nil
}

func Test_dbClient_Write(t *testing.T) {
	ctx := context.Background()

	container, err := setupMysql(ctx)
	if err != nil {
		t.Fatal(err)
	}
	host, _ := container.Host(ctx)
	p, _ := container.MappedPort(ctx, "3306/tcp")
	port := p.Int()

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=skip-verify&parseTime=true", "root", "password", host, port, "database")

	writer := NewDBWriter(connectionString)
	writer.Migrate(filepath.Join("..", "migrations"))

	if err != nil {
		t.Fatalf("error creating table: %v", err)
	}

	referenceTime := time.Date(2023, 1, 1, 11, 22, 33, 0, time.UTC)

	for i := 0; i < 2; i++ {
		err = writer.Write(referenceTime, "coffee", "the.arm@blacklodge", "fire", "fire walk with me")
		if err != nil {
			t.Fatalf("error writing to DB: %v", err)
		}
	}

	data, err := queryData(connectionString)
	if err != nil {
		t.Fatalf("error fetching created data: %v", err)
	}

	if len(data) != 1 {
		t.Errorf("number of created rows was '%d' and not '1'", len(data))
	}

	testData := []chatLine{{
		msgTimestamp: referenceTime,
		msgId:        "coffee",
		fromAddress:  "the.arm@blacklodge",
		subject:      "fire",
		body:         "fire walk with me",
	}}

	if !(reflect.DeepEqual(data, testData)) {
		t.Errorf("%v - %v", testData[0].msgTimestamp, data[0].msgTimestamp)
		t.Errorf("expected '%v' but got '%v'", testData, data)
	}
}
