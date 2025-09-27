package di

import (
	"go.uber.org/dig"

	"abc-runner/app/core/config"
	"abc-runner/app/core/utils"
)

// Container 依赖注入容器
type Container struct {
	container *dig.Container
}

// NewContainer 创建依赖注入容器
func NewContainer() *Container {
	container := dig.New()

	// 注册核心组件
	container.Provide(config.NewConfigManager)
	container.Provide(utils.NewOperationRegistry)
	container.Provide(utils.NewDefaultKeyGenerator)

	// 注意：AdapterFactory 和 MetricsCollector 由应用层在运行时注册

	return &Container{
		container: container,
	}
}

// Provide 注册依赖
func (c *Container) Provide(constructor interface{}, opts ...dig.ProvideOption) error {
	return c.container.Provide(constructor, opts...)
}

// Invoke 调用函数并注入依赖
func (c *Container) Invoke(function interface{}, opts ...dig.InvokeOption) error {
	return c.container.Invoke(function, opts...)
}
