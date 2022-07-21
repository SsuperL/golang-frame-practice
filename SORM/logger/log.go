// 自定义log，输出行号和日志级别，不同日志级别颜色不同
// info为蓝色，error为红色
package logger

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	errlog  = log.New(os.Stdout, "\033[31m[error]\033[0m", log.Lshortfile|log.LstdFlags)
	infolog = log.New(os.Stdout, "\033[34m[info]\033[0m", log.Lshortfile|log.LstdFlags)
	loggers = []*log.Logger{errlog, infolog}
	mu      sync.Mutex
)

var (
	Error  = errlog.Fatal
	Errorf = errlog.Fatalf
	Info   = infolog.Println
	Infof  = infolog.Printf
)

const (
	// Infolevel ...
	Infolevel = iota + 1
	// Errorlevel ...
	Errorlevel
	// Disable ...
	Disable
)

// SetLevel set level of log
func SetLevel(level int) {
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if level > Infolevel {
		infolog.SetOutput(ioutil.Discard)
	}

	if level > Errorlevel {
		errlog.SetOutput(ioutil.Discard)
	}

}
