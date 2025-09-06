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
	"redis-runner/app/core/unified"
	"redis-runner/app/utils"
)

// 全局变量
var (
	unifiedCommandRouter      *command.CommandRouter
	unifiedManager     unified.UnifiedCommandManager
	useUnifiedManager  bool = true  // 标志：是否使用新的统一管理器
)

func mainUnified() {
	utils.LogConfig()
	defer closeUnifiedLogFile()

	// 初始化命令系统
	if err := initializeUnifiedCommandSystem(); err != nil {
		log.Fatalf("Failed to initialize command system: %v", err)
	}

	// 执行命令
	if err := executeUnifiedCommand(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

// initializeUnifiedCommandSystem 初始化命令系统
func initializeUnifiedCommandSystem() error {
	if useUnifiedManager {
		// 使用新的统一命令管理器
		return initializeUnifiedManager()
	} else {
		// 使用旧的命令路由器（向后兼容）
		return initializeUnifiedCommandRouter()
	}
}

// initializeUnifiedManager 初始化统一命令管理器
func initializeUnifiedManager() error {
	log.Println("Initializing unified command manager...")
	
	// 创建统一命令管理器
	unifiedManager = unified.NewUnifiedCommandManager()
	
	// 注册协议适配器
	if err := registerProtocolAdapters(); err != nil {
		return fmt.Errorf("failed to register protocol adapters: %w", err)
	}
	
	// 设置默认协议
	if err := unifiedManager.SetDefaultProtocol("redis"); err != nil {
		return fmt.Errorf("failed to set default protocol: %w", err)
	}
	
	log.Println("Unified command manager initialized successfully")
	return nil
}

// registerProtocolAdapters 注册协议适配器
func registerProtocolAdapters() error {
	// 这里应该注册实际的协议适配器
	// 暂时使用Mock适配器，实际实现应该引用真实的适配器
	
	// 注册Redis协议适配器
	redisAdapter := &unified.MockRedisAdapter{}
	if err := unifiedManager.RegisterProtocol("redis", redisAdapter); err != nil {
		return fmt.Errorf("failed to register redis adapter: %w", err)
	}
	
	// 注册HTTP协议适配器
	httpAdapter := &unified.MockHttpAdapter{}
	if err := unifiedManager.RegisterProtocol("http", httpAdapter); err != nil {
		return fmt.Errorf("failed to register http adapter: %w", err)
	}
	
	// 注册Kafka协议适配器
	kafkaAdapter := &unified.MockKafkaAdapter{}
	if err := unifiedManager.RegisterProtocol("kafka", kafkaAdapter); err != nil {
		return fmt.Errorf("failed to register kafka adapter: %w", err)
	}
	
	return nil
}

// initializeUnifiedCommandRouter 初始化传统命令路由器（向后兼容）
func initializeUnifiedCommandRouter() error {
	log.Println("Initializing legacy command router...")
	
	unifiedCommandRouter = command.NewCommandRouter()

	// 注册增强版命令
	if err := commands.RegisterEnhancedCommands(unifiedCommandRouter); err != nil {
		return fmt.Errorf("failed to register enhanced commands: %w", err)
	}

	// 注册传统版命令(DEPRECATED)
	if err := command.RegisterLegacyCommands(unifiedCommandRouter); err != nil {
		return fmt.Errorf("failed to register legacy commands: %w", err)
	}

	log.Println("Legacy command router initialized successfully")
	return nil
}

// executeUnifiedCommand 执行命令
func executeUnifiedCommand() error {
	// 处理全局标志
	if handled := handleUnifiedGlobalFlags(); handled {
		return nil
	}

	// 解析命令和参数
	subCmd, args := parseUnifiedCommandArgs()
	if subCmd == "" {
		fmt.Println("Please specify a command")
		showUnifiedGlobalHelp()
		return nil
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 根据配置选择执行方式
	if useUnifiedManager {
		return executeWithUnifiedManager(ctx, subCmd, args)
	} else {
		return executeWithUnifiedCommandRouter(ctx, subCmd, args)
	}
}

// executeWithUnifiedManager 使用统一管理器执行命令
func executeWithUnifiedManager(ctx context.Context, command string, args []string) error {
	log.Printf("Executing command with unified manager: %s", command)
	
	// 使用统一命令管理器执行
	result, err := unifiedManager.ExecuteCommand(ctx, command, args)
	if err != nil {
		return fmt.Errorf("unified manager execution failed: %w", err)
	}
	
	// 输出结果
	printUnifiedResult(result)
	
	return nil
}

// executeWithUnifiedCommandRouter 使用传统路由器执行命令
func executeWithUnifiedCommandRouter(ctx context.Context, subCmd string, args []string) error {
	log.Printf("Executing command with legacy router: %s", subCmd)
	
	// 使用传统命令路由器
	return unifiedCommandRouter.Route(ctx, subCmd, args)
}

// printUnifiedResult 打印统一管理器的执行结果
func printUnifiedResult(result *unified.CommandResult) {
	if result == nil {
		fmt.Println("No result returned")
		return
	}
	
	fmt.Printf("\n=== Command Execution Result ===\n")
	fmt.Printf("Command: %s\n", result.Command)
	fmt.Printf("Protocol: %s\n", result.Protocol)
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Duration: %v\n", result.Duration)
	
	if result.Output != "" {
		fmt.Printf("Output: %s\n", result.Output)
	}
	
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}
	
	if len(result.Suggestions) > 0 {
		fmt.Printf("Suggestions:\n")
		for _, suggestion := range result.Suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}
	
	if result.Metadata != nil && len(result.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for key, value := range result.Metadata {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
	
	fmt.Printf("================================\n\n")
}

// handleUnifiedGlobalFlags 处理全局标志
func handleUnifiedGlobalFlags() bool {
	help := flag.Bool("help", false, "show help information")
	version := flag.Bool("version", false, "show version information")
	unified := flag.Bool("unified", true, "use unified command manager (default: true)")
	legacy := flag.Bool("legacy", false, "use legacy command router")
	flag.Parse()

	// 设置管理器模式
	if *legacy {
		useUnifiedManager = false
		log.Println("Using legacy command router mode")
	} else if *unified {
		useUnifiedManager = true
		log.Println("Using unified command manager mode")
	}

	if *help {
		showUnifiedGlobalHelp()
		return true
	}

	if *version {
		utils.PrintVersion()
		return true
	}

	return false
}

// parseUnifiedCommandArgs 解析命令和参数
func parseUnifiedCommandArgs() (string, []string) {
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
			showUnifiedGlobalHelp()
			return "", nil
		}
		
		// 显示命令帮助
		if useUnifiedManager {
			if help, err := unifiedManager.GetCommandHelp(subCmd); err == nil {
				fmt.Print(help)
			} else {
				fmt.Printf("Unknown command: %s\n\n", subCmd)
				showGlobalHelp()
			}
		} else if commandRouter != nil {
			if info, exists := unifiedCommandRouter.GetCommandInfo(subCmd); exists {
				fmt.Printf("Command: %s\n\n%s\n", subCmd, info.Usage)
			} else {
				fmt.Printf("Unknown command: %s\n\n", subCmd)
				showUnifiedGlobalHelp()
			}
		}
		return "", nil
	}

	return subCmd, args
}

// showUnifiedGlobalHelp 显示全局帮助信息
func showUnifiedGlobalHelp() {
	if useUnifiedManager && unifiedManager != nil {
		showUnifiedHelp()
	} else if unifiedCommandRouter != nil {
		showUnifiedLegacyHelp()
	} else {
		showBasicHelp()
	}
}

// showUnifiedHelp 显示统一管理器的帮助信息
func showUnifiedHelp() {
	fmt.Println("redis-runner - Unified Performance Testing Tool")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  redis-runner <command> [options]")
	fmt.Println()
	
	// 获取可用命令
	commands := unifiedManager.ListCommands()
	
	// 分类显示命令
	enhancedCommands := []unified.CommandInfo{}
	legacyCommands := []unified.CommandInfo{}
	
	for _, cmd := range commands {
		if cmd.Deprecated {
			legacyCommands = append(legacyCommands, cmd)
		} else {
			enhancedCommands = append(enhancedCommands, cmd)
		}
	}
	
	if len(enhancedCommands) > 0 {
		fmt.Println("ENHANCED COMMANDS (Recommended):")
		for _, cmd := range enhancedCommands {
			fmt.Printf("  %-20s %s\n", cmd.Name, cmd.Description)
		}
		fmt.Println()
	}
	
	if len(legacyCommands) > 0 {
		fmt.Println("LEGACY COMMANDS (DEPRECATED):")
		for _, cmd := range legacyCommands {
			replacement := ""
			if cmd.Replacement != "" {
				replacement = fmt.Sprintf(" (use %s)", cmd.Replacement)
			}
			fmt.Printf("  %-20s ⚠️  %s%s\n", cmd.Name, cmd.Description, replacement)
		}
		fmt.Println()
	}
	
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --help, -h       Show help information")
	fmt.Println("  --version, -v    Show version information")
	fmt.Println("  --unified        Use unified command manager (default)")
	fmt.Println("  --legacy         Use legacy command router")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  redis-runner redis-enhanced --config conf/redis.yaml")
	fmt.Println("  redis-runner r -n 1000 -c 10")
	fmt.Println("  redis-runner http-enhanced --url http://localhost:8080")
	fmt.Println()
	fmt.Println("Use \"redis-runner <command> --help\" for more information about a command.")
}

// showUnifiedLegacyHelp 显示传统路由器的帮助信息
func showUnifiedLegacyHelp() {
	if flag.NArg() > 0 {
		subCmd := flag.Arg(0)
		// 显示特定命令的帮助
		if info, exists := unifiedCommandRouter.GetCommandInfo(subCmd); exists {
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
			fmt.Print(unifiedCommandRouter.GenerateHelp())
		}
	} else {
		// 显示总体帮助信息
		fmt.Print(unifiedCommandRouter.GenerateHelp())
	}
}

// showBasicHelp 显示基础帮助信息
func showBasicHelp() {
	fmt.Println("redis-runner - Performance Testing Tool")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  redis-runner <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  redis-enhanced   Redis performance testing (recommended)")
	fmt.Println("  http-enhanced    HTTP load testing (recommended)")
	fmt.Println("  kafka-enhanced   Kafka performance testing (recommended)")
	fmt.Println("  redis            Redis performance testing (deprecated)")
	fmt.Println("  http             HTTP load testing (deprecated)")
	fmt.Println("  kafka            Kafka performance testing (deprecated)")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --help, -h       Show help information")
	fmt.Println("  --version, -v    Show version information")
	fmt.Println()
}

// closeUnifiedLogFile 关闭日志文件
func closeUnifiedLogFile() {
	if utils.LogFile() != nil {
		err := utils.LogFile().Close()
		if err != nil {
			log.Printf("failed to close log file: %v", err)
		}
	}
}

// 兼容性函数 - 保持向后兼容

// commandExeUnified 兼容原有的commandExe函数
func commandExeUnified() {
	if err := executeUnifiedCommand(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
		os.Exit(1)
	}
}