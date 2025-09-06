package integration

import (
	"os"
	"strings"
	"testing"

	"redis-runner/app/commands"
	"redis-runner/app/core/command"
)

// TestCommandIntegration 集成测试验证新旧版本功能对等性
func TestCommandIntegration(t *testing.T) {
	// 设置测试环境
	setupTestEnvironment(t)

	// 创建命令路由器
	router := command.NewCommandRouter()

	// 注册所有命令
	if err := commands.RegisterEnhancedCommands(router); err != nil {
		t.Fatalf("Failed to register enhanced commands: %v", err)
	}

	if err := command.RegisterLegacyCommands(router); err != nil {
		t.Fatalf("Failed to register legacy commands: %v", err)
	}

	// 测试命令注册完整性
	testCommandRegistration(t, router)

	// 测试帮助系统
	testHelpSystem(t, router)

	// 测试弃用警告
	testDeprecationWarnings(t, router)

	// 测试参数验证
	testArgumentValidation(t, router)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("REDIS_RUNNER_TEST_MODE", "true")
	os.Setenv("REDIS_RUNNER_LOG_LEVEL", "debug")

	// 确保测试目录存在
	if err := os.MkdirAll("test/fixtures", 0755); err != nil {
		t.Logf("Warning: Could not create test directory: %v", err)
	}
}

// testCommandRegistration 测试命令注册完整性
func testCommandRegistration(t *testing.T, router *command.CommandRouter) {
	t.Run("CommandRegistration", func(t *testing.T) {
		// 验证所有预期的命令都已注册
		expectedEnhanced := []string{
			"redis-enhanced",
			"http-enhanced",
			"kafka-enhanced",
		}

		expectedLegacy := []string{
			"redis",
			"http",
			"kafka",
		}

		registry := router.GetRegistry()

		// 检查enhanced命令
		enhancedCommands := registry.ListEnhanced()
		for _, expected := range expectedEnhanced {
			found := false
			for _, cmd := range enhancedCommands {
				if cmd == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Enhanced command '%s' not found in registry", expected)
			}
		}

		// 检查legacy命令
		legacyCommands := registry.ListLegacy()
		for _, expected := range expectedLegacy {
			found := false
			for _, cmd := range legacyCommands {
				if cmd == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Legacy command '%s' not found in registry", expected)
			}
		}

		// 验证命令版本信息
		for _, cmd := range expectedEnhanced {
			info, exists := router.GetCommandInfo(cmd)
			if !exists {
				t.Errorf("Command info not found for '%s'", cmd)
				continue
			}

			if info.Version != command.Enhanced {
				t.Errorf("Command '%s' should be Enhanced version, got %v", cmd, info.Version)
			}

			if info.Deprecated {
				t.Errorf("Enhanced command '%s' should not be deprecated", cmd)
			}
		}

		for _, cmd := range expectedLegacy {
			info, exists := router.GetCommandInfo(cmd)
			if !exists {
				t.Errorf("Command info not found for '%s'", cmd)
				continue
			}

			if info.Version != command.Legacy {
				t.Errorf("Command '%s' should be Legacy version, got %v", cmd, info.Version)
			}

			if !info.Deprecated {
				t.Errorf("Legacy command '%s' should be deprecated", cmd)
			}
		}
	})
}

// testHelpSystem 测试帮助系统
func testHelpSystem(t *testing.T, router *command.CommandRouter) {
	t.Run("HelpSystem", func(t *testing.T) {
		// 测试全局帮助
		help := router.GenerateHelp()
		if help == "" {
			t.Error("Global help should not be empty")
		}

		// 验证帮助内容包含必要信息
		expectedContent := []string{
			"Enhanced Commands (Recommended)",
			"Legacy Commands (DEPRECATED)",
			"redis-enhanced",
			"http-enhanced",
			"kafka-enhanced",
			"Migration guide",
		}

		for _, content := range expectedContent {
			if !strings.Contains(help, content) {
				t.Errorf("Help should contain '%s', but it's missing", content)
			}
		}

		// 测试特定命令帮助
		commands := []string{"redis-enhanced", "http-enhanced", "kafka-enhanced"}
		for _, cmd := range commands {
			info, exists := router.GetCommandInfo(cmd)
			if !exists {
				t.Errorf("Command info not found for '%s'", cmd)
				continue
			}

			if info.Usage == "" {
				t.Errorf("Usage should not be empty for command '%s'", cmd)
			}

			if !strings.Contains(info.Usage, cmd) {
				t.Errorf("Usage for '%s' should contain the command name", cmd)
			}
		}
	})
}

// testDeprecationWarnings 测试弃用警告
func testDeprecationWarnings(t *testing.T, router *command.CommandRouter) {
	t.Run("DeprecationWarnings", func(t *testing.T) {
		// 注意：这个测试只验证调用不会出错
		// 实际的弃用警告是通过标准输出显示的
		legacyCommands := []string{"redis", "http", "kafka"}

		for _, cmd := range legacyCommands {
			// 验证legacy命令可以被识别
			_, exists := router.GetRegistry().Get(cmd)
			if !exists {
				t.Errorf("Legacy command '%s' should be registered", cmd)
				continue
			}

			// 验证命令被标记为deprecated
			info, exists := router.GetCommandInfo(cmd)
			if !exists {
				t.Errorf("Command info not found for legacy command '%s'", cmd)
				continue
			}

			if !info.Deprecated {
				t.Errorf("Legacy command '%s' should be marked as deprecated", cmd)
			}

			if info.Version != command.Legacy {
				t.Errorf("Command '%s' should have Legacy version, got %v", cmd, info.Version)
			}
		}
	})
}

// testArgumentValidation 测试参数验证
func testArgumentValidation(t *testing.T, router *command.CommandRouter) {
	t.Run("ArgumentValidation", func(t *testing.T) {
		// 测试enhanced命令的参数验证
		enhancedCommands := map[string][]string{
			"redis-enhanced": {"-h", "localhost", "-p", "6379", "-n", "100"},
			"http-enhanced":  {"--url", "http://example.com", "-n", "100"},
			"kafka-enhanced": {"--broker", "localhost:9092", "--topic", "test", "-n", "100"},
		}

		for cmd, validArgs := range enhancedCommands {
			handler, exists := router.GetRegistry().Get(cmd)
			if !exists {
				t.Errorf("Command '%s' not found", cmd)
				continue
			}

			// 测试有效参数
			err := handler.ValidateArgs(validArgs)
			if err != nil {
				t.Errorf("Valid arguments for '%s' should not cause validation error: %v", cmd, err)
			}

			// 测试空参数（对于enhanced命令，空参数应该导致显示帮助而不是错误）
			err = handler.ValidateArgs([]string{})
			if err != nil {
				t.Logf("Empty arguments for '%s' returned: %v (this may be expected)", cmd, err)
			}
		}

		// 测试legacy命令的参数验证（更宽松）
		legacyCommands := map[string][]string{
			"redis": {"-h", "localhost", "-p", "6379"},
			"http":  {"--url", "http://example.com"},
			"kafka": {"--broker", "localhost:9092"},
		}

		for cmd, args := range legacyCommands {
			handler, exists := router.GetRegistry().Get(cmd)
			if !exists {
				t.Errorf("Legacy command '%s' not found", cmd)
				continue
			}

			// Legacy命令的验证应该更宽松
			err := handler.ValidateArgs(args)
			if err != nil {
				t.Logf("Legacy command '%s' validation returned: %v", cmd, err)
			}
		}
	})
}

// TestCommandVersionMigration 测试版本迁移
func TestCommandVersionMigration(t *testing.T) {
	t.Run("VersionMigration", func(t *testing.T) {
		router := command.NewCommandRouter()

		// 注册命令
		if err := commands.RegisterEnhancedCommands(router); err != nil {
			t.Fatalf("Failed to register enhanced commands: %v", err)
		}

		if err := command.RegisterLegacyCommands(router); err != nil {
			t.Fatalf("Failed to register legacy commands: %v", err)
		}

		// 验证每个legacy命令都有对应的enhanced版本
		migrationPairs := map[string]string{
			"redis": "redis-enhanced",
			"http":  "http-enhanced",
			"kafka": "kafka-enhanced",
		}

		for legacy, enhanced := range migrationPairs {
			// 验证legacy命令存在
			_, legacyExists := router.GetRegistry().Get(legacy)
			if !legacyExists {
				t.Errorf("Legacy command '%s' should exist", legacy)
			}

			// 验证enhanced命令存在
			_, enhancedExists := router.GetRegistry().Get(enhanced)
			if !enhancedExists {
				t.Errorf("Enhanced command '%s' should exist", enhanced)
			}

			// 验证两个命令都可以获取信息
			legacyInfo, legacyInfoExists := router.GetCommandInfo(legacy)
			enhancedInfo, enhancedInfoExists := router.GetCommandInfo(enhanced)

			if !legacyInfoExists {
				t.Errorf("Legacy command info for '%s' should exist", legacy)
			}

			if !enhancedInfoExists {
				t.Errorf("Enhanced command info for '%s' should exist", enhanced)
			}

			// 验证版本标记正确
			if legacyInfoExists && !legacyInfo.Deprecated {
				t.Errorf("Legacy command '%s' should be deprecated", legacy)
			}

			if enhancedInfoExists && enhancedInfo.Deprecated {
				t.Errorf("Enhanced command '%s' should not be deprecated", enhanced)
			}
		}
	})
}

// TestCommandExecutionParity 测试命令执行对等性
func TestCommandExecutionParity(t *testing.T) {
	t.Run("ExecutionParity", func(t *testing.T) {
		// 这个测试验证新旧版本在相似参数下的行为一致性
		// 由于需要实际的服务连接，这里只做基本的接口验证

		router := command.NewCommandRouter()

		// 注册命令
		if err := commands.RegisterEnhancedCommands(router); err != nil {
			t.Fatalf("Failed to register enhanced commands: %v", err)
		}

		if err := command.RegisterLegacyCommands(router); err != nil {
			t.Fatalf("Failed to register legacy commands: %v", err)
		}

		// 验证命令接口的一致性
		commandPairs := map[string]string{
			"redis": "redis-enhanced",
			"http":  "http-enhanced",
			"kafka": "kafka-enhanced",
		}

		for legacy, enhanced := range commandPairs {
			legacyHandler, legacyExists := router.GetRegistry().Get(legacy)
			enhancedHandler, enhancedExists := router.GetRegistry().Get(enhanced)

			if !legacyExists || !enhancedExists {
				t.Errorf("Both legacy (%v) and enhanced (%v) commands should exist for %s",
					legacyExists, enhancedExists, legacy)
				continue
			}

			// 验证两个处理器都实现了相同的接口
			if legacyHandler.GetCommandName() == "" {
				t.Errorf("Legacy handler for '%s' should have a valid command name", legacy)
			}

			if enhancedHandler.GetCommandName() == "" {
				t.Errorf("Enhanced handler for '%s' should have a valid command name", enhanced)
			}

			// 验证描述不为空
			if legacyHandler.GetDescription() == "" {
				t.Errorf("Legacy handler for '%s' should have a description", legacy)
			}

			if enhancedHandler.GetDescription() == "" {
				t.Errorf("Enhanced handler for '%s' should have a description", enhanced)
			}

			// 验证使用说明不为空
			if legacyHandler.GetUsage() == "" {
				t.Errorf("Legacy handler for '%s' should have usage information", legacy)
			}

			if enhancedHandler.GetUsage() == "" {
				t.Errorf("Enhanced handler for '%s' should have usage information", enhanced)
			}
		}
	})
}
