package locater

import (
	"os"

	"github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}[%{level:.6s}] ▶ %{time:2006-01-02 15:04:09} %{shortfile} %{message} %{color:reset}`,
)

// SetLogger is printing out log message to STDOUT and LOGFILE
func SetLogger(f *os.File) {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(f, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)
}
