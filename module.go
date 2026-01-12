// Copyright 2025 zampo.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// @contact  zampo3380@gmail.com

package app

// Module 模块接口，定义业务模块的标准接口
// 每个业务模块（如 api_module）都应该实现此接口
type Module interface {
	// Name 返回模块名称
	Name() string

	// Initialize 初始化模块，接收 App 对象作为依赖注入
	// 返回模块注册的服务列表
	// 通过 app 对象可以访问所有依赖：Config, MySQL, Redis, Kafka, GetDependencyContainer() 等
	Initialize(app *App) ([]ServiceRegistration, error)

	// Shutdown 关闭模块，清理资源
	Shutdown() error
}

// ModuleRegistry 模块注册表
type ModuleRegistry struct {
	modules []Module
}

// NewModuleRegistry 创建新的模块注册表
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make([]Module, 0),
	}
}

// Register 注册模块
func (r *ModuleRegistry) Register(module Module) {
	r.modules = append(r.modules, module)
}

// RegisterAll 批量注册模块
func (r *ModuleRegistry) RegisterAll(modules ...Module) {
	r.modules = append(r.modules, modules...)
}

// GetModules 获取所有注册的模块
func (r *ModuleRegistry) GetModules() []Module {
	return r.modules
}

// InitializeAll 初始化所有模块，返回所有服务的注册信息
func (r *ModuleRegistry) InitializeAll(app *App) ([]ServiceRegistration, error) {
	var allServices []ServiceRegistration

	for _, module := range r.modules {
		services, err := module.Initialize(app)
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

// ShutdownAll 关闭所有模块
func (r *ModuleRegistry) ShutdownAll() error {
	var lastErr error
	for _, module := range r.modules {
		if err := module.Shutdown(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
