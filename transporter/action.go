package transporter

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

type HostAction struct {
	Host map[string]string
	mtx  sync.Mutex
}

type VisitMode string

const (
	Block    VisitMode = "0"
	Upstream VisitMode = "1"
	Direct   VisitMode = "2"
)

const (
	ModeWhite = "white"
	ModeAll   = "all"
)

func NewActionConf() *HostAction {
	return &HostAction{
		Host: map[string]string{},
		mtx:  sync.Mutex{},
	}
}

var strM = map[VisitMode]string{
	Upstream: "upstream",
	Block:    "block",
	Direct:   "direct",
}

func (vm VisitMode) String() string {
	if v, ok := strM[vm]; ok {
		return v
	}
	return string(vm)
}

var mStr = map[string]VisitMode{
	"upstream": Upstream,
	"block":    Block,
	"direct":   Direct,
}

func (h *HostAction) Add(name, action string) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	if v, ok := mStr[action]; ok {
		action = string(v)
	}
	h.Host[name] = action
}

func (h *HostAction) Detect(name string) VisitMode {
	n := name
	for {
		if a, ok := h.Host[n]; ok {
			return VisitMode(a)
		}
		if i := strings.Index(n, "."); i > 0 {
			n = n[i+1:]
		} else {
			return ""
		}
	}
}

func (h *HostAction) Load(p string) error {
	if r, err := os.OpenFile(p, os.O_RDONLY, os.ModePerm); err != nil {
		return err
	} else {
		br := bufio.NewReader(r)
		for {
			line, _, err := br.ReadLine()
			if io.EOF == err {
				break
			}
			if err != nil {
				return err
			}
			ln := strings.TrimSpace(string(line))
			if len(ln) > 1 && ln[0] != '#' {
				fd := strings.Fields(ln)
				if len(fd) >= 2 {
					h.Add(fd[0], strings.Join(fd[1:], ";"))
				} else {
					// 使用默认方式
					h.Add(fd[0], "")
				}
			}
		}
		return nil
	}
}
