package client

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"mellium.im/sasl"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/muc"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
	"time"
	"xmpp-bouncer/logger"
)

type Connection struct {
	Mux     *mux.ServeMux
	Client  *muc.Client
	Session *xmpp.Session
	Jid     jid.JID
}

type logWriter struct {
	logType string
}

func (w logWriter) Write(p []byte) (int, error) {
	logger.Sugar.Infow("XML", "type", w.logType, "payload", string(p))
	return len(p), nil
}

func createSession(ctx context.Context, jid jid.JID, password string) (*xmpp.Session, error) {
	dialCtx, dialCtxCancel := context.WithTimeout(ctx, 30*time.Second)
	defer dialCtxCancel()

	conn, err := dial.Client(dialCtx, "tcp", jid)
	if err != nil {
		return nil, fmt.Errorf("error dialing session: %w", err)
	}

	negotiator := xmpp.NewNegotiator(func(*xmpp.Session, *xmpp.StreamConfig) xmpp.StreamConfig {
		return xmpp.StreamConfig{
			Lang: "en",
			Features: []xmpp.StreamFeature{
				xmpp.BindResource(),
				xmpp.StartTLS(&tls.Config{
					ServerName: jid.Domain().String(),
					MinVersion: tls.VersionTLS12,
				}),
				xmpp.SASL("", password, sasl.ScramSha1Plus, sasl.ScramSha1, sasl.Plain),
			},
			TeeIn:  logWriter{logType: "IN"},
			TeeOut: logWriter{logType: "OUT"},
		}
	})
	session, err := xmpp.NewSession(dialCtx, jid.Domain(), jid, conn, 0, negotiator)
	if err != nil {
		return nil, fmt.Errorf("error creating a new session: %w", err)
	}

	return session, nil
}

func Connect(ctx context.Context, address string, password string, messageHandlerFunc mux.MessageHandlerFunc) (Connection, error) {
	jabberId, err := jid.Parse(address)
	if err != nil {
		return Connection{}, fmt.Errorf("error parsing address %q: %w", address, err)
	}

	session, err := createSession(ctx, jabberId, password)
	if err != nil {
		return Connection{}, fmt.Errorf("error creating session: %w", err)
	}

	err = session.Send(ctx, stanza.Presence{Type: stanza.AvailablePresence}.Wrap(nil))
	if err != nil {
		return Connection{}, fmt.Errorf("error sending available presence: %w", err)
	}

	mucClient := &muc.Client{}
	jabberMux := mux.New(
		stanza.NSClient,
		mux.Message(stanza.ChatMessage, xml.Name{Local: "body"}, messageHandlerFunc),
		mux.Message(stanza.GroupChatMessage, xml.Name{Local: "body"}, messageHandlerFunc),
		muc.HandleClient(mucClient),
	)

	return Connection{
		Mux:     jabberMux,
		Client:  mucClient,
		Session: session,
		Jid:     jabberId,
	}, nil
}
