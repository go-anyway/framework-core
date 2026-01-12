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
	"fmt"
	"sync"
)

// Lifecycle 生命周期管理器
type Lifecycle struct {
	shutdownFuncs []func() error
	mu            sync.Mutex
	started       bool
}

// NewLifecycle 创建新的生命周期管理器
func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		shutdownFuncs: make([]func() error, 0),
	}
}

// RegisterShutdown 注册关闭函数
func (l *Lifecycle) RegisterShutdown(fn func() error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.shutdownFuncs = append(l.shutdownFuncs, fn)
}

// Shutdown 执行所有关闭函数
func (l *Lifecycle) Shutdown() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error
	// 逆序执行关闭函数（后注册的先关闭）
	for i := len(l.shutdownFuncs) - 1; i >= 0; i-- {
		if err := l.shutdownFuncs[i](); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// SetStarted 标记应用已启动
func (l *Lifecycle) SetStarted() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.started = true
}

// IsStarted 检查应用是否已启动
func (l *Lifecycle) IsStarted() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.started
}
