package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var lf *os.File

func LogFile() *os.File {
	return lf
}

func LogConfig() {

	err := os.MkdirAll("logs", 0755)
	if err != nil {
		log.Fatalf("failed to create folder: %v", err)
	}
	timestamp := time.Now().Format("20060102")
	base := fmt.Sprintf("logs/record_%s", timestamp)
	fn := base + "_1.log"
	seq := 1
	for {
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			break
		}
		fn = fmt.Sprintf("%s_%d.log", base, seq)
		seq++
	}

	lf, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	log.SetOutput(lf)
	// log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("this is a custom log entry.")
}
