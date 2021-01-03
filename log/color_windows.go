// +build windows

package log

func (w *colorWriter) ColorWrite(level LogLevel, p []byte) (int, error) {
	return w.w.Write(p)
}
