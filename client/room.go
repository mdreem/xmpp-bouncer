package client

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"os"
	"xmpp-bouncer/logger"
)

type Room struct {
	RoomAddress string `yaml:"room_address"`
	RoomPass    string `yaml:"room_pass"`
}

type Rooms struct {
	Rooms map[string]Room `yaml:"rooms"`
}

func JoinRoom(ctx context.Context, connection Connection, room string, roomPass string) error {
	roomID, err := jid.Parse(room)
	if err != nil {
		return fmt.Errorf("error parsing room %s: %v", room, err)
	}

	logger.Sugar.Infow("joining the chat room", "chatroom", roomID.String())
	username := connection.Jid.Localpart()
	roomJID, err := roomID.WithResource(username)
	if err != nil {
		return fmt.Errorf("error adding resource part %s: %v", username, err)
	}

	opts := []muc.Option{muc.MaxBytes(0)}
	if roomPass != "" {
		opts = append(opts, muc.Password(roomPass))
	}
	opts = append(opts, muc.MaxBytes(10*1024))

	_, err = connection.Client.Join(ctx, roomJID, connection.Session, opts...)
	if err != nil {
		return fmt.Errorf("error joining MUC %s: %v", room, err)
	}
	return nil
}

func JoinRooms(ctx context.Context, connection Connection) {
	if _, err := os.Stat("rooms.yaml"); errors.Is(err, os.ErrNotExist) {
		logger.Sugar.Info("no 'rooms.yaml' present")
		return
	}

	yamlFile, err := os.ReadFile("rooms.yaml")
	if err != nil {
		logger.Sugar.Fatalw("unable to open 'rooms.yaml'", "error", err)
	}

	var roomData Rooms
	err = yaml.Unmarshal(yamlFile, &roomData)
	if err != nil {
		logger.Sugar.Fatalw("unable to unmarshal 'rooms.yaml'", "error", err)
	}

	for roomName, roomInfo := range roomData.Rooms {
		logger.Sugar.Infow("joining room", "room", roomName)
		err = JoinRoom(ctx, connection, roomInfo.RoomAddress, roomInfo.RoomPass)
		if err != nil {
			logger.Sugar.Fatalw("failed to join room", "room", roomName, "error", err)
		}
	}
}
