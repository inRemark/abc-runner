package main

import (
	"flag"
	"fmt"
	"log"
	httpCases "redis-runner/app/http_cases"
	kafkaCases "redis-runner/app/kafka_cases"
	redisCases "redis-runner/app/redis_cases"
	"redis-runner/app/utils"
)

func main() {
	utils.LogConfig()
	commandExe()
	if utils.LogFile() != nil {
		err := utils.LogFile().Close()
		if err != nil {
			log.Printf("failed to close log file: %v", err)
		}
	}
}

func commandExe() {
	help := flag.Bool("help", false, "help info")
	version := flag.Bool("version", false, "version info")
	flag.Parse()

	if *help {
		showHelpUsage()
		return
	}

	if *version {
		utils.PrintVersion()
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Please specify a command")
		utils.PrintHelpUsage()
		return
	}

	subCmd := flag.Arg(0)
	args := flag.Args()[1:]
	if len(args) < 1 {
		fmt.Println("Please specify a argument.")
		utils.PrintHelpUsage()
	} else {
		subCommandExe(subCmd, args)
	}
}

func subCommandExe(subCmd string, args []string) bool {
	switch subCmd {
	case "redis":
		redisCases.RedisCommand(args)
	case "http":
		httpCases.HttpCommand(args)
	case "kafka":
		kafkaCases.KafkaCommand(args)
	default:
		fmt.Printf("Unknown sub command: %s\n", subCmd)
	}
	return true
}

func showHelpUsage() {
	if flag.NArg() > 0 {
		subCmd := flag.Arg(0)
		if subCmd == "redis" {
			utils.PrintRedisUsage()
		} else if subCmd == "http" {
			utils.PrintHttpUsage()
		}
	} else {
		utils.PrintHelpUsage()
	}
}
