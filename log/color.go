package log

import (
	"bytes"
	"io"
	"os"
)

type colorWriter struct {
	writer
}

func NewColorWriter() io.Writer {
	return &colorWriter{writer{os.Stdout, TextMarshaler}}
}

func (w *colorWriter) WriteLogMessage(m *LogMessage) error {
	var msg []byte
	if v, err := w.fn(m); err != nil {
		return err
	} else {
		msg = v
	}
	_, err := w.ColorWrite(m.Level, msg)
	return err
}

func (w *colorWriter) Write(p []byte) (int, error) {
	m := new(LogMessage)
	if er := m.unmarshal(bytes.NewBuffer(p)); er != nil {
		// 解码失败
		return w.ColorWrite(Ldebug, p)
	}
	return len(p), w.WriteLogMessage(m)
}
