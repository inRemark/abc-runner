package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"redis-runner/app/commands"
	"redis-runner/app/core/command"
	"redis-runner/app/core/unified"
	"redis-runner/app/utils"
)

// 全局变量
var (
	commandRouter      *command.CommandRouter
	unifiedManager     unified.UnifiedCommandManager
	useUnifiedManager  bool = true  // 标志：是否使用新的统一管理器
)

func main() {
	utils.LogConfig()
	defer closeLogFile()

	// 初始化命令系统
	if err := initializeCommandSystem(); err != nil {
		log.Fatalf("Failed to initialize command system: %v", err)
	}

	// 执行命令
	if err := executeCommand(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

// initializeCommandSystem 初始化命令系统
func initializeCommandSystem() error {
	if useUnifiedManager {
		// 使用新的统一命令管理器
		return initializeUnifiedManager()
	} else {
		// 使用旧的命令路由器（向后兼容）
		return initializeCommandRouter()
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

// initializeCommandRouter 初始化传统命令路由器（向后兼容）
func initializeCommandRouter() error {
	log.Println("Initializing legacy command router...")
	
	commandRouter = command.NewCommandRouter()

	// 注册增强版命令
	if err := commands.RegisterEnhancedCommands(commandRouter); err != nil {
		return fmt.Errorf("failed to register enhanced commands: %w", err)
	}

	// 注册传统版命令(DEPRECATED)
	if err := command.RegisterLegacyCommands(commandRouter); err != nil {
		return fmt.Errorf("failed to register legacy commands: %w", err)
	}

	log.Println("Legacy command router initialized successfully")
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

	// 根据配置选择执行方式
	if useUnifiedManager {
		return executeWithUnifiedManager(ctx, subCmd, args)
	} else {
		return executeWithCommandRouter(ctx, subCmd, args)
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

// executeWithCommandRouter 使用传统路由器执行命令
func executeWithCommandRouter(ctx context.Context, subCmd string, args []string) error {
	log.Printf("Executing command with legacy router: %s", subCmd)
	
	// 使用传统命令路由器
	return commandRouter.Route(ctx, subCmd, args)
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

// handleGlobalFlags 处理全局标志
func handleGlobalFlags() bool {
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
		
		// 显示命令帮助
		if useUnifiedManager && unifiedManager != nil {
			if help, err := unifiedManager.GetCommandHelp(subCmd); err == nil {
				fmt.Print(help)
			} else {
				fmt.Printf("Unknown command: %s\n\n", subCmd)
				showGlobalHelp()
			}
		} else if commandRouter != nil {
			if info, exists := commandRouter.GetCommandInfo(subCmd); exists {
				fmt.Printf("Command: %s\n\n%s\n", subCmd, info.Usage)
			} else {
				fmt.Printf("Unknown command: %s\n\n", subCmd)
				showGlobalHelp()
			}
		}
		return "", nil
	}

	return subCmd, args
}

// showGlobalHelp 显示全局帮助信息
func showGlobalHelp() {
	if useUnifiedManager && unifiedManager != nil {
		showUnifiedHelp()
	} else if commandRouter != nil {
		showLegacyHelp()
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
		fmt.Println("LEGACY COMMANDS (REMOVED):")
		for _, cmd := range legacyCommands {
			replacement := ""
			if cmd.Replacement != "" {
				replacement = fmt.Sprintf(" (use %s)", cmd.Replacement)
			}
			fmt.Printf("  %-20s ❌  %s%s\n", cmd.Name, cmd.Description, replacement)
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

// showLegacyHelp 显示传统路由器的帮助信息
func showLegacyHelp() {
	if commandRouter != nil {
		fmt.Print(commandRouter.GenerateHelp())
	} else {
		showBasicHelp()
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
	fmt.Println("  redis            Redis performance testing (REMOVED)")
	fmt.Println("  http             HTTP load testing (REMOVED)")
	fmt.Println("  kafka            Kafka performance testing (REMOVED)")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --help, -h       Show help information")
	fmt.Println("  --version, -v    Show version information")
	fmt.Println()
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