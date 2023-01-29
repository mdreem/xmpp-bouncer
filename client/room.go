package client

import (
	"context"
	"fmt"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"xmpp-bouncer/logger"
)

func JoinRoom(ctx context.Context, connection Connection, room string, roomPass string) error {
	roomId, err := jid.Parse(room)
	if err != nil {
		return fmt.Errorf("error parsing room %s: %v", room, err)
	}

	logger.Sugar.Infow("joining the chat room", "chatroom", roomId.String())
	username := connection.Jid.Localpart()
	roomJID, err := roomId.WithResource(username)
	if err != nil {
		return fmt.Errorf("error adding resource part %s: %v", username, err)
	}

	opts := []muc.Option{muc.MaxBytes(0)}
	if roomPass != "" {
		opts = append(opts, muc.Password(roomPass))
	}

	_, err = connection.Client.Join(ctx, roomJID, connection.Session, opts...)
	if err != nil {
		return fmt.Errorf("error joining MUC %s: %v", room, err)
	}
	return nil
}
