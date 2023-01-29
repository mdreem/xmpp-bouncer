package persistence

import (
	"encoding/xml"
	"io"
	"mellium.im/xmlstream"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
	"xmpp-bouncer/logger"
)

type messageBody struct {
	stanza.Message
	Subject string `xml:"subject,omitempty"`
	Thread  string `xml:"thread,omitempty"`
	Body    string `xml:"body"`
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

		err = chatWriter.Write(msg.ID, from.String(), msg.Subject, msg.Body)
		if err != nil {
			logger.Sugar.Errorw("error persisting message", "error", err)
			return nil
		}

		logger.Sugar.Infow("message", "from", from, "subject", msg.Subject, "body", msg.Body)
		return nil
	}
}
