package persistence

import (
	"fmt"
	"os"
	"xmpp-bouncer/logger"
)

type dbClient struct {
}

func New() ChatWriter {
	return dbClient{}
}

type ChatWriter interface {
	Write(ID string, from string, subject string, body string) error
}

func (client dbClient) Write(ID string, from string, subject string, body string) error {
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

	_, err = f.WriteString(fmt.Sprintf("%s;%s;%s;%s\n", ID, from, subject, body))
	if err != nil {
		return fmt.Errorf("error writing message to file: %w", err)
	}
	return nil
}
