package lib

import (
	"log"
	"time"
)

// ErrorLogger is error log for self
func ErrorLogger(errAgs error) {
	timeMark := time.Now()
	log.Println(timeMark.Format("[2006/01/02 15:04:05]"), errAgs)
}
