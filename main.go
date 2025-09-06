package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"redis-runner/app/commands"
	"redis-runner/app/core/command"
	"redis-runner/app/utils"
)

// 全局变量
var (
	commandRouter *command.CommandRouter
)

func main() {
	utils.LogConfig()
	defer closeLogFile()

	// 初始化命令路由器
	if err := initializeCommandRouter(); err != nil {
		log.Fatalf("Failed to initialize command router: %v", err)
	}

	// 执行命令
	if err := executeCommand(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

// initializeCommandRouter 初始化命令路由器
func initializeCommandRouter() error {
	commandRouter = command.NewCommandRouter()

	// 注册增强版命令
	if err := commands.RegisterEnhancedCommands(commandRouter); err != nil {
		return fmt.Errorf("failed to register enhanced commands: %w", err)
	}

	// 注册传统版命令(DEPRECATED)
	if err := command.RegisterLegacyCommands(commandRouter); err != nil {
		return fmt.Errorf("failed to register legacy commands: %w", err)
	}

	log.Println("Command router initialized successfully")
	return nil
}

// executeCommand 执行命令
func executeCommand() error {
	// 处理全局标志
	if handled := handleGlobalFlags(); handled {
		return nil
	}

	// 解析命令和参数
	subCmd, args := parseCommandArgs()
	if subCmd == "" {
		fmt.Println("Please specify a command")
		showGlobalHelp()
		return nil
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 路由到具体命令
	return commandRouter.Route(ctx, subCmd, args)
}

// handleGlobalFlags 处理全局标志
func handleGlobalFlags() bool {
	help := flag.Bool("help", false, "show help information")
	version := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *help {
		showGlobalHelp()
		return true
	}

	if *version {
		utils.PrintVersion()
		return true
	}

	return false
}

// parseCommandArgs 解析命令和参数
func parseCommandArgs() (string, []string) {
	if flag.NArg() < 1 {
		return "", nil
	}

	subCmd := flag.Arg(0)
	args := flag.Args()[1:]

	// 对于传统命令，如果没有参数也允许执行
	legacyCommands := []string{"redis", "http", "kafka"}
	for _, legacy := range legacyCommands {
		if subCmd == legacy {
			return subCmd, args
		}
	}

	// 对于增强版命令，检查最少参数要求
	if len(args) < 1 {
		// 检查是否是请求帮助
		if subCmd == "help" || subCmd == "-h" || subCmd == "--help" {
			showGlobalHelp()
			return "", nil
		}
		// 对于增强版命令，没有参数时显示该命令的帮助
		if info, exists := commandRouter.GetCommandInfo(subCmd); exists {
			fmt.Printf("Command: %s\n\n%s\n", subCmd, info.Usage)
			return "", nil
		}
	}

	return subCmd, args
}

// showGlobalHelp 显示全局帮助信息
func showGlobalHelp() {
	if flag.NArg() > 0 {
		subCmd := flag.Arg(0)
		// 显示特定命令的帮助
		if info, exists := commandRouter.GetCommandInfo(subCmd); exists {
			fmt.Printf("Command: %s\n\n%s\n", subCmd, info.Usage)
			return
		}
		// 传统方式显示帮助
		if subCmd == "redis" {
			utils.PrintRedisUsage()
		} else if subCmd == "http" {
			utils.PrintHttpUsage()
		} else {
			fmt.Printf("Unknown command: %s\n\n", subCmd)
			fmt.Print(commandRouter.GenerateHelp())
		}
	} else {
		// 显示总体帮助信息
		fmt.Print(commandRouter.GenerateHelp())
	}
}

// closeLogFile 关闭日志文件
func closeLogFile() {
	if utils.LogFile() != nil {
		err := utils.LogFile().Close()
		if err != nil {
			log.Printf("failed to close log file: %v", err)
		}
	}
}

// 兼容性函数 - 保持向后兼容

// commandExe 兼容原有的commandExe函数
func commandExe() {
	if err := executeCommand(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
		os.Exit(1)
	}
}
