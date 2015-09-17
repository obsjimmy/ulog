package ulog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mgutz/ansi"
)

// C type its a key-value map refering context in function
type C map[string]interface{}

// Stat type its the basic
type Stat struct {
	*log.Logger
	BeginTime time.Time
	Caller    string
	Context   string
	dmode     bool
}

// DefaultLogger its the logger to be used for each new created stat
var DefaultLogger = log.New(os.Stdout, "", log.LstdFlags|log.Ltime|log.Lmicroseconds)

// OutputConfig shows verbose messages
var OutputConfig = struct {
	ShowStarts  bool
	ShowDones   bool
	ShowInfos   bool
	ShowWarns   bool
	ShowErrors  bool
	ShowDetails bool
	ShowDebug   bool
}{true, true, true, true, true, true, true}

var (
	// LInfo shows information messages
	LInfo = ansi.Color("info", "green")

	// LWarn shows warning messages
	LWarn = ansi.Color("warn", "yellow")

	// LErr shows error messages
	LErr = ansi.Color("error", "red")

	// LStart shows start messages
	LStart = ansi.Color("start", "blue+h")

	// LDone shows done messages
	LDone = ansi.Color("done", "blue+h")

	// STime is style for time
	STime = ansi.ColorFunc("magenta+h")

	// SFail is style for faillures
	SFail = ansi.ColorFunc("red+h")

	// SKey is style for keys
	SKey = ansi.ColorFunc("cyan")

	// SVal is style for values
	SVal = ansi.ColorFunc("white+b")
)

func funcName(d int) string {
	pc, _, _, _ := runtime.Caller(d + 1)
	return runtime.FuncForPC(pc).Name()
}

func fileName(d int) (string, int) {
	_, file, line, _ := runtime.Caller(d + 1)

	w := strings.LastIndex(file, "/")
	if w < len(file)-1 && w > 0 {
		file = file[w+1:]
	}

	return file, line
}

// RemoveFormat removes the current format
func RemoveFormat() {
	LInfo = "info"
	LWarn = "warn"
	LErr = "error"
	LStart = "start"
	LDone = "done"
	STime = ansi.ColorFunc("")
	SFail = ansi.ColorFunc("")
	SKey = ansi.ColorFunc("")
	SVal = ansi.ColorFunc("")
}

// String returns a JSON with the values of context
func (c C) String() string {
	var res []string
	for k, v := range c {
		res = append(res, fmt.Sprintf("%s["+SVal("%v")+"]", SKey(k), v))
	}

	return " " + strings.Join(res, ", ")
}

func (s *Stat) logfmt(d int, level, format string, v ...interface{}) {
	f, l := fileName(d + 1)

	if format == "" {
		s.Output(d+1, fmt.Sprintf("%s:%d : %s %s%s", f, l, s.Caller, level, s.Context))
	} else {
		stuff := fmt.Sprintf(format, v...)
		s.Output(d+1, fmt.Sprintf("%s:%d : %s %s%s: %s", f, l, s.Caller, level, s.Context, stuff))
	}
}

// New returns a new Context recording the start time and logging the value.
func New(ctx C) *Stat {
	var s *Stat

	if ctx == nil {
		s = &Stat{DefaultLogger, time.Now(), funcName(1), "", false}
	} else {
		s = &Stat{DefaultLogger, time.Now(), funcName(1), ctx.String(), false}
	}

	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowStarts) {
		s.logfmt(1, LStart, "")
	}

	return s
}

// NewDbg returns a new Context recording the start time and logging the value.
func NewDbg(ctx C) *Stat {
	var s *Stat

	if ctx == nil {
		s = &Stat{DefaultLogger, time.Now(), funcName(1), "", true}
	} else {
		s = &Stat{DefaultLogger, time.Now(), funcName(1), ctx.String(), true}
	}

	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowStarts) {
		s.logfmt(1, LStart, "")
	}

	return s
}

// I formats an info message
func (s *Stat) I(format string, v ...interface{}) {
	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowInfos) {
		s.logfmt(1, LInfo, format, v...)
	}
}

// W formats a warning message
func (s *Stat) W(format string, v ...interface{}) {
	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowWarns) {
		s.logfmt(1, LWarn, format, v...)
	}
}

// E formats an error message
func (s *Stat) E(format string, v ...interface{}) {
	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowErrors) {
		s.logfmt(1, LErr, format, v...)
	}
}

// ED formats an error message with depth
func (s *Stat) ED(d int, format string, v ...interface{}) {
	if OutputConfig.ShowErrors {
		s.logfmt(d+1, LErr, format, v...)
	}
}

// Sub returns a new stat with inherited properties
func (s *Stat) Sub(name string, ctx C) *Stat {
	var n *Stat
	if ctx == nil {
		n = &Stat{DefaultLogger, time.Now(), funcName(1) + "(" + name + ")", "", s.dmode}
	} else {
		n = &Stat{DefaultLogger, time.Now(), funcName(1) + "(" + name + ")", ctx.String(), s.dmode}
	}

	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowStarts) {
		n.logfmt(1, LStart, "")
	}

	return n
}

// Detail prints a pretty JSON into logs
func (s *Stat) Detail(level, head string, v interface{}) {
	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowDetails) {
		v, _ := json.MarshalIndent(v, "", "\t")
		s.logfmt(1, level, "%s -> \n%s", head, v)
	}
}

// Done outputs a done message
func (s *Stat) Done(errs ...interface{}) {
	if (s.dmode && OutputConfig.ShowDebug) || (!s.dmode && OutputConfig.ShowDones) {
		for _, e := range errs {
			if e != nil {
				s.ED(1, SFail("failure")+" : %v", e)
			}
		}

		s.logfmt(1, LDone, "time["+STime("%s")+"]", time.Since(s.BeginTime))
	}
}
