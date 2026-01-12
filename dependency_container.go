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
	"reflect"
)

// DependencyContainer 类型安全的依赖容器
// 通过类型来注册和获取依赖，提供编译时类型检查
type DependencyContainer struct {
	deps map[reflect.Type]interface{}
}

// register 内部注册方法（非泛型，供 RegisterDependency 使用）
func (c *DependencyContainer) register(instance interface{}) {
	typ := reflect.TypeOf(instance)
	if typ == nil {
		return
	}
	if c.deps == nil {
		c.deps = make(map[reflect.Type]interface{})
	}
	c.deps[typ] = instance
}

// NewDependencyContainer 创建新的依赖容器
func NewDependencyContainer() *DependencyContainer {
	return &DependencyContainer{
		deps: make(map[reflect.Type]interface{}),
	}
}

// Register 注册依赖实例（类型安全）
// 使用示例: Register(container, myService)
func Register[T any](c *DependencyContainer, instance T) {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	c.deps[typ] = instance
}

// Get 获取依赖实例（类型安全）
// 使用示例: service, ok := Get[MyService](container)
func Get[T any](c *DependencyContainer) (T, bool) {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	dep, ok := c.deps[typ]
	if !ok {
		var zero T
		return zero, false
	}
	instance, ok := dep.(T)
	return instance, ok
}

// MustGet 获取依赖实例，如果不存在则 panic
// 使用示例: service := MustGet[MyService](container)
func MustGet[T any](c *DependencyContainer) T {
	instance, ok := Get[T](c)
	if !ok {
		var zero T
		typ := reflect.TypeOf(zero)
		panic(fmt.Errorf("dependency of type %s not found", typ))
	}
	return instance
}

// Has 检查依赖是否存在
// 使用示例: if Has[MyService](container) { ... }
func Has[T any](c *DependencyContainer) bool {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	_, ok := c.deps[typ]
	return ok
}

// GetAll 获取所有已注册的依赖类型
func (c *DependencyContainer) GetAll() []reflect.Type {
	types := make([]reflect.Type, 0, len(c.deps))
	for typ := range c.deps {
		types = append(types, typ)
	}
	return types
}
