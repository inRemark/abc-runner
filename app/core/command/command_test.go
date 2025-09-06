package command

import (
	"context"
	"testing"
	"time"
)

// MockCommandHandler 用于测试的模拟命令处理器
type MockCommandHandler struct {
	*BaseCommandHandler
	executeCalled bool
	executeError  error
}

func NewMockCommandHandler(name, description string, version CommandVersion, deprecated bool) *MockCommandHandler {
	baseHandler := NewBaseCommandHandler(name, description, version, deprecated, nil, nil)
	return &MockCommandHandler{
		BaseCommandHandler: baseHandler,
	}
}

func (m *MockCommandHandler) ExecuteCommand(ctx context.Context, args []string) error {
	m.executeCalled = true
	return m.executeError
}

func (m *MockCommandHandler) SetExecuteError(err error) {
	m.executeError = err
}

func (m *MockCommandHandler) WasExecuteCalled() bool {
	return m.executeCalled
}

// TestCommandRegistry 测试命令注册器
func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	// 测试注册命令
	handler := NewMockCommandHandler("test-cmd", "Test command", Enhanced, false)
	err := registry.Register("test-cmd", handler)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// 测试获取命令
	retrievedHandler, exists := registry.Get("test-cmd")
	if !exists {
		t.Fatal("Command not found after registration")
	}

	if retrievedHandler.GetCommandName() != "test-cmd" {
		t.Errorf("Expected command name 'test-cmd', got '%s'", retrievedHandler.GetCommandName())
	}

	// 测试重复注册
	err = registry.Register("test-cmd", handler)
	if err == nil {
		t.Error("Expected error when registering duplicate command")
	}

	// 测试列出命令
	commands := registry.List()
	if len(commands) != 1 || commands[0] != "test-cmd" {
		t.Errorf("Expected ['test-cmd'], got %v", commands)
	}

	// 测试注销命令
	registry.Unregister("test-cmd")
	_, exists = registry.Get("test-cmd")
	if exists {
		t.Error("Command should not exist after unregistration")
	}
}

// TestCommandRouter 测试命令路由器
func TestCommandRouter(t *testing.T) {
	router := NewCommandRouter()

	// 注册测试命令
	enhancedHandler := NewMockCommandHandler("enhanced-cmd", "Enhanced command", Enhanced, false)
	legacyHandler := NewMockCommandHandler("legacy-cmd", "Legacy command", Legacy, true)

	err := router.RegisterCommand("enhanced-cmd", enhancedHandler)
	if err != nil {
		t.Fatalf("Failed to register enhanced command: %v", err)
	}

	err = router.RegisterCommand("legacy-cmd", legacyHandler)
	if err != nil {
		t.Fatalf("Failed to register legacy command: %v", err)
	}

	// 测试路由到enhanced命令
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = router.Route(ctx, "enhanced-cmd", []string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("Failed to route to enhanced command: %v", err)
	}

	if !enhancedHandler.WasExecuteCalled() {
		t.Error("Enhanced command execute was not called")
	}

	// 测试路由到legacy命令（应该显示弃用警告）
	err = router.Route(ctx, "legacy-cmd", []string{"arg1"})
	if err != nil {
		t.Errorf("Failed to route to legacy command: %v", err)
	}

	if !legacyHandler.WasExecuteCalled() {
		t.Error("Legacy command execute was not called")
	}

	// 测试路由到不存在的命令
	err = router.Route(ctx, "unknown-cmd", []string{})
	if err == nil {
		t.Error("Expected error when routing to unknown command")
	}
}

// TestBaseCommandHandler 测试基础命令处理器
func TestBaseCommandHandler(t *testing.T) {
	handler := NewBaseCommandHandler(
		"test-command",
		"Test description",
		Enhanced,
		false,
		nil,
		nil,
	)

	// 测试基本属性
	if handler.GetCommandName() != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", handler.GetCommandName())
	}

	if handler.GetDescription() != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", handler.GetDescription())
	}

	if handler.GetVersion() != Enhanced {
		t.Errorf("Expected version Enhanced, got %v", handler.GetVersion())
	}

	if handler.IsDeprecated() {
		t.Error("Command should not be deprecated")
	}

	// 测试弃用的命令
	deprecatedHandler := NewBaseCommandHandler(
		"deprecated-cmd",
		"Deprecated command",
		Legacy,
		true,
		nil,
		nil,
	)

	if !deprecatedHandler.IsDeprecated() {
		t.Error("Command should be deprecated")
	}

	if deprecatedHandler.GetVersion() != Legacy {
		t.Errorf("Expected version Legacy, got %v", deprecatedHandler.GetVersion())
	}
}

// TestCommandRouterHelp 测试命令路由器帮助功能
func TestCommandRouterHelp(t *testing.T) {
	router := NewCommandRouter()

	// 注册测试命令
	enhancedHandler := NewMockCommandHandler("redis-enhanced", "Redis enhanced testing", Enhanced, false)
	legacyHandler := NewMockCommandHandler("redis", "Redis testing (deprecated)", Legacy, true)

	router.RegisterCommand("redis-enhanced", enhancedHandler)
	router.RegisterCommand("redis", legacyHandler)

	// 测试生成帮助信息
	help := router.GenerateHelp()
	if help == "" {
		t.Error("Help should not be empty")
	}

	// 帮助信息应该包含enhanced和legacy命令
	if !contains(help, "redis-enhanced") {
		t.Error("Help should contain enhanced command")
	}

	if !contains(help, "DEPRECATED") {
		t.Error("Help should indicate deprecated commands")
	}

	// 测试获取命令信息
	info, exists := router.GetCommandInfo("redis-enhanced")
	if !exists {
		t.Error("Should find command info for enhanced command")
	}

	if info.Version != Enhanced {
		t.Errorf("Expected Enhanced version, got %v", info.Version)
	}

	if info.Deprecated {
		t.Error("Enhanced command should not be deprecated")
	}

	// 测试获取不存在命令的信息
	_, exists = router.GetCommandInfo("nonexistent")
	if exists {
		t.Error("Should not find info for nonexistent command")
	}
}

// TestCommandVersions 测试命令版本处理
func TestCommandVersions(t *testing.T) {
	registry := NewCommandRegistry()

	// 注册不同版本的命令
	enhancedHandler := NewMockCommandHandler("test-enhanced", "Enhanced test", Enhanced, false)
	legacyHandler := NewMockCommandHandler("test-legacy", "Legacy test", Legacy, true)

	registry.Register("test-enhanced", enhancedHandler)
	registry.Register("test-legacy", legacyHandler)

	// 测试列出enhanced命令
	enhanced := registry.ListEnhanced()
	if len(enhanced) != 1 || enhanced[0] != "test-enhanced" {
		t.Errorf("Expected ['test-enhanced'], got %v", enhanced)
	}

	// 测试列出legacy命令
	legacy := registry.ListLegacy()
	if len(legacy) != 1 || legacy[0] != "test-legacy" {
		t.Errorf("Expected ['test-legacy'], got %v", legacy)
	}

	// 测试获取命令信息
	commandsInfo := registry.GetCommandsInfo()
	if len(commandsInfo) != 2 {
		t.Errorf("Expected 2 commands info, got %d", len(commandsInfo))
	}

	enhancedInfo, exists := commandsInfo["test-enhanced"]
	if !exists {
		t.Error("Should find enhanced command info")
	}

	if enhancedInfo.Version != Enhanced || enhancedInfo.Deprecated {
		t.Error("Enhanced command info is incorrect")
	}

	legacyInfo, exists := commandsInfo["test-legacy"]
	if !exists {
		t.Error("Should find legacy command info")
	}

	if legacyInfo.Version != Legacy || !legacyInfo.Deprecated {
		t.Error("Legacy command info is incorrect")
	}
}

// TestCommandValidation 测试命令参数验证
func TestCommandValidation(t *testing.T) {
	handler := NewMockCommandHandler("test-cmd", "Test command", Enhanced, false)

	// 测试默认验证（应该总是通过）
	err := handler.ValidateArgs([]string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("Default validation should pass, got error: %v", err)
	}

	// 测试空参数验证
	err = handler.ValidateArgs([]string{})
	if err != nil {
		t.Errorf("Default validation should pass for empty args, got error: %v", err)
	}

	// 测试nil参数验证
	err = handler.ValidateArgs(nil)
	if err != nil {
		t.Errorf("Default validation should pass for nil args, got error: %v", err)
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && indexOf(s, substr) >= 0))
}

// indexOf 查找子字符串位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
