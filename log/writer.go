package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type LogMarshaler func(m *LogMessage) ([]byte, error)

type writer struct {
	w  io.Writer
	fn LogMarshaler
}

func NewWriter(w io.Writer, fn LogMarshaler) io.Writer {
	return &writer{w, fn}
}

func (w *writer) WriteLogMessage(m *LogMessage) error {
	var msg []byte
	if v, err := w.fn(m); err != nil {
		return err
	} else {
		msg = v
	}
	_, err := w.w.Write(msg)
	return err
}

func (w *writer) Write(p []byte) (int, error) {
	m := new(LogMessage)
	if er := m.unmarshal(bytes.NewBuffer(p)); er != nil {
		// 解码失败
		return w.w.Write(p)
	}
	return len(p), w.WriteLogMessage(m)
}

func NewTextWriter(w io.Writer) io.Writer {
	return NewWriter(w, TextMarshaler)
}

func NewJsonWriter(w io.Writer) io.Writer {
	return NewWriter(w, func(m *LogMessage) ([]byte, error) {
		return json.Marshal(m)
	})
}

const TimeFormat = "2006-01-02 15:04:05.000"

func TextMarshaler(m *LogMessage) ([]byte, error) {
	var msg string
	if len(m.File) > 0 {
		msg = fmt.Sprintf("%s [%-5s] %s:%d %s", m.Time.Format(TimeFormat), m.Level, m.File, m.Line, m.Message)
	} else {
		msg = fmt.Sprintf("%s [%-5s] %s", m.Time.Format(TimeFormat), m.Level, m.Message)
	}
	return []byte(msg), nil
}
