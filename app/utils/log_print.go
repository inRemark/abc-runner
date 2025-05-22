package utils

import (
	"fmt"
	"log"
)

func PrintConsole(arr []string) {
	for _, str := range arr {
		fmt.Printf(str)
	}
}

func PrintLogs(arr []string) {
	for _, str := range arr {
		log.Printf(str)
	}
}
