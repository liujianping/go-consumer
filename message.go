package consumer

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"time"
)

// The number of bytes for a Message.ID
const MsgIDLength = 16

// MessageID is the ASCII encoded hexadecimal message ID
type MessageID [MsgIDLength]byte

// Message is the fundamental data type containing
// the id, body, and metadata
type Message struct {
	ID        MessageID
	Body      []byte
	Timestamp int64
	Attempts  uint16
}

// NewMessage creates a Message, initializes some metadata,
// and returns a pointer
func NewMessage(id MessageID, body []byte) *Message {
	return &Message{
		ID:        id,
		Body:      body,
		Timestamp: time.Now().UnixNano(),
	}
}

// WriteTo implements the WriterTo interface and serializes
// the message into the supplied producer.
//
// It is suggested that the target Writer is buffered to
// avoid performing many system calls.
func (m *Message) WriteTo(w io.Writer) (int64, error) {
	var buf [10]byte
	var total int64

	binary.BigEndian.PutUint64(buf[:8], uint64(m.Timestamp))
	binary.BigEndian.PutUint16(buf[8:10], uint16(m.Attempts))

	n, err := w.Write(buf[:])
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(m.ID[:])
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(m.Body)
	total += int64(n)
	if err != nil {
		return total, err
	}

	return total, nil
}

// DecodeMessage deseralizes data (as []byte) and creates a new Message
func DecodeMessage(b []byte) (*Message, error) {
	var msg Message

	msg.Timestamp = int64(binary.BigEndian.Uint64(b[:8]))
	msg.Attempts = binary.BigEndian.Uint16(b[8:10])

	buf := bytes.NewBuffer(b[10:])

	_, err := io.ReadFull(buf, msg.ID[:])
	if err != nil {
		return nil, err
	}

	msg.Body, err = ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
