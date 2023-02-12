package persistence

import (
	"encoding/xml"
	"io"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
	"time"
	"xmpp-bouncer/logger"
)

type messageBody struct {
	stanza.Message
	Subject  string   `xml:"subject,omitempty"`
	Thread   string   `xml:"thread,omitempty"`
	Body     string   `xml:"body"`
	Delay    delay    `xml:"delay"`
	StanzaID stanzaID `xml:"stanza-id"`
}

type stanzaID struct {
	By string `xml:"by,attr,omitempty"`
	ID string `xml:"id,attr,omitempty"`
}

type delay struct {
	Stamp time.Time `xml:"stamp,attr,omitempty"`
	From  string    `xml:"from,attr,omitempty"`
}

func ReceiveMessage(chatWriter ChatWriter) mux.MessageHandlerFunc {
	return func(m stanza.Message, t xmlstream.TokenReadEncoder) error {
		d := xml.NewTokenDecoder(t)
		from := m.From
		if m.Type != stanza.GroupChatMessage {
			from = m.From.Bare()
		}

		msg := messageBody{}
		err := d.Decode(&msg)
		if err != nil && err != io.EOF {
			logger.Sugar.Errorw("error decoding message", "error", err)
			return nil
		}

		var timestamp = msg.Delay.Stamp
		if (timestamp == time.Time{}) {
			timestamp = time.Now().UTC()
		}

		err = chatWriter.Write(timestamp, msg.StanzaID.ID, from.String(), msg.Subject, msg.Body)
		if err != nil {
			logger.Sugar.Errorw("error persisting message", "error", err)
			return nil
		}

		logger.Sugar.Infow("message", "ID", msg.StanzaID.ID, "from", from, "subject", msg.Subject, "body", msg.Body, "msg_timestamp", timestamp)
		return nil
	}
}
