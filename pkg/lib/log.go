package lib

import (
	"log"
	"os"
	"strings"
	"time"
)

// InitLog is init log info
func InitLog() {
	log.SetOutput(os.Stdout)
}

// CustomLogger is custom log for self
func CustomLogger(args ...string) {
	logInfo := strings.Join(args, " ")
	timeMark := time.Now()
	log.Println(timeMark.Format("[2006/01/02 15:04:05]"), logInfo)
}
