package log

import (
	"io"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	w := io.MultiWriter(NewTextWriter(os.Stdout))
	log := New(w, true)
	if err := log.Output(1, "default", Linfo, "information\n"); err != nil {
		t.Error(err)
	}

	log.Println(Ldebug, Application("user"), "user info", "some", 1, "@", "11")
	log.Println(Application("user"), Lwarn, "user info", "some", 1, "@", "11")
	log.Println(Lerror, "user info", "some", 1, "@", "11")

	SetLogCaller(true)
	SetOutput(NewTextWriter(os.Stdout))
	Warn("user info", "some", 1, "@", "11")
	Error(Application("user"), "user info", "some", 1, "@", "11")
	_ = Output(1, "app", Linfo, "message\n")
}
