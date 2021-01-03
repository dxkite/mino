// +build !windows

package log

func (w *colorWriter) ColorWrite(level LogLevel, msg []byte) (int, error) {
	var tpl = "%s"
	switch level {
	case Lerror:
		tpl = "\x1b[31;1m%s\x1b[0m"
	case Lwarn:
		tpl = "\x1b[33;1m%s\x1b[0m"
	case Linfo:
		tpl = "\x1b[36;1m%s\x1b[0m"
	case Ldebug:
	}
	n, err := w.w.Write([]byte(fmt.Sprintf(tpl, string(msg))))
	return n, err
}
