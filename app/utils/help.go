package utils

import "fmt"

var AppName = "redis-runner"
var Version = "0.0.1"
var ReleaseAt = "2025-05-21"

func PrintVersion() {
	fmt.Printf("Version: %s \n", Version)
	fmt.Printf("Release date: %s\n", ReleaseAt)
}

func PrintHelpUsage() {
	fmt.Printf("Usage: %s [--help] [--version] <command> [arguments]", AppName)
	fmt.Printf("\n")
	fmt.Println("Sub Commands:")
	fmt.Printf("\n")
	fmt.Println("  redis      Redis benchmark ")
	fmt.Println("  http       HTTP benchmark")
	fmt.Printf("\n")
	fmt.Println("Use tool help <command> for more information about a command.")
	fmt.Printf("\n")
	PrintRedisUsage()
	fmt.Printf("\n")
	PrintHttpUsage()
}

func PrintRedisUsage() {
	fmt.Printf("Usage: %s redis [arguments]", AppName)
	fmt.Printf("\n")
	fmt.Println("  --cluster	redis cluster mode (default false)")
	fmt.Println("  -h       	Server hostname (default 127.0.0.1) ")
	fmt.Println("  -p       	Server port (default 6379)")
	fmt.Println("  -a       	Password for Redis Auth")
	fmt.Println("  -n       	Total number of requests (default 100000)")
	fmt.Println("  -c       	Number of parallel connections (default 50)")
	fmt.Println("  -d       	Data size of SET/GET value in bytes (default 3)")
	fmt.Println("  -r       	Read operation percent (default 100%)")
	fmt.Println("  -ttl     	TTL in seconds (default 300)")
}

func PrintHttpUsage() {
	fmt.Printf("Usage: %s http [arguments]", AppName)
	fmt.Printf("\n")
	fmt.Println("  -url       	Server url (default http://localhost:8080)")
	fmt.Println("  -m       	Method get/post (default get)")
	fmt.Println("  -n       	Total number of requests (default 10000)")
	fmt.Println("  -c       	Number of parallel connections (default 50)")
}
