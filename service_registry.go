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
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// GRPCServiceRegister gRPC 服务注册函数类型
// 用于注册 gRPC 服务到 gRPC 服务器
type GRPCServiceRegister func(server *grpc.Server)

// GatewayServiceRegister 网关服务注册函数类型
// 用于注册 gRPC 服务到 HTTP 网关
// 函数签名与 gateway.RegisterService 兼容
type GatewayServiceRegister func(ctx context.Context, mux *runtime.ServeMux, grpcAddr string, opts []grpc.DialOption) error

// ServiceRegistration 服务注册信息
type ServiceRegistration struct {
	// Name 服务名称
	Name string
	// GRPCRegister gRPC 服务注册函数
	GRPCRegister GRPCServiceRegister
	// GatewayRegister 网关服务注册函数（可选）
	GatewayRegister GatewayServiceRegister
}

// ServiceRegistry 服务注册表
type ServiceRegistry struct {
	services []ServiceRegistration
}

// NewServiceRegistry 创建新的服务注册表
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make([]ServiceRegistration, 0),
	}
}

// Register 注册服务
func (r *ServiceRegistry) Register(reg ServiceRegistration) {
	r.services = append(r.services, reg)
}

// RegisterAll 批量注册服务
func (r *ServiceRegistry) RegisterAll(regs ...ServiceRegistration) {
	r.services = append(r.services, regs...)
}

// GetServices 获取所有注册的服务
func (r *ServiceRegistry) GetServices() []ServiceRegistration {
	return r.services
}

// RegisterToGRPCServer 将所有服务注册到 gRPC 服务器
func (r *ServiceRegistry) RegisterToGRPCServer(server *grpc.Server) {
	for _, reg := range r.services {
		if reg.GRPCRegister != nil {
			reg.GRPCRegister(server)
		}
	}
}

// RegisterToGateway 将所有服务注册到 HTTP 网关
// 使用 gateway.GatewayHandler 的 RegisterService 方法
func (r *ServiceRegistry) RegisterToGateway(ctx context.Context, gwHandler interface {
	RegisterService(ctx context.Context, grpcAddr string, registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error, opts ...grpc.DialOption) error
}, grpcAddr string, opts []grpc.DialOption) error {
	for _, reg := range r.services {
		if reg.GatewayRegister != nil {
			// 创建一个适配器函数，将 GatewayServiceRegister 转换为 gateway.RegisterService 期望的格式
			registerFunc := func(ctx context.Context, mux *runtime.ServeMux, grpcAddr string, opts []grpc.DialOption) error {
				return reg.GatewayRegister(ctx, mux, grpcAddr, opts)
			}
			if err := gwHandler.RegisterService(ctx, grpcAddr, registerFunc, opts...); err != nil {
				return err
			}
		}
	}
	return nil
}
