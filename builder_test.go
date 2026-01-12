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

import (
	"testing"
)

// TestAppBuilder_CollectInitializedServices 测试收集已初始化服务列表
func TestAppBuilder_CollectInitializedServices(t *testing.T) {
	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "test-service",
		},
	}

	builder := NewAppBuilder(cfg)

	// 模拟一些已初始化的服务
	builder.app.MySQL = nil // 设置为 nil 表示未初始化
	builder.app.Redis = nil

	// 注册一些依赖类型
	builder.app.dependencies = append(builder.app.dependencies, Dependency{
		Type:        DependencyWebSocket,
		Name:        "WebSocket",
		Initialized: true,
	})

	services := builder.collectInitializedServices()

	// 应该包含 WebSocket（通过依赖检查）
	found := false
	for _, s := range services {
		if s == "WebSocket" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected WebSocket to be in initialized services")
	}
}
