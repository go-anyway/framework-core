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
	"net"
	"time"

	"github.com/go-anyway/framework-interceptor"
	"github.com/go-anyway/framework-log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server 服务器结构体
type Server struct {
	grpcServer *grpc.Server
	app        *App
}

// StartRPCServer 启动 gRPC 服务器（仅 RPC，不启动 HTTP）
func (a *App) StartRPCServer() (*Server, error) {
	serverCfg, err := a.Config.GetServer()
	if err != nil {
		return nil, err
	}

	// 启动 gRPC 服务器
	grpcServer, err := a.startGRPCServer(serverCfg.GRPC)
	if err != nil {
		return nil, err
	}

	server := &Server{
		grpcServer: grpcServer,
		app:        a,
	}

	// 注册关闭函数
	a.lifecycle.RegisterShutdown(func() error {
		grpcServer.GracefulStop()
		log.Info("gRPC server stopped", zap.String("service", a.ServiceName))
		return nil
	})

	a.lifecycle.SetStarted()
	return server, nil
}

// startGRPCServer 启动 gRPC 服务器
func (a *App) startGRPCServer(grpcCfg GRPCConfig) (*grpc.Server, error) {
	address := grpcCfg.Address()
	lc := &net.ListenConfig{}
	lis, err := lc.Listen(context.Background(), "tcp", address)
	if err != nil {
		return nil, err
	}

	// 获取服务器配置以检查功能开关
	serverCfg, err := a.Config.GetServer()
	if err != nil {
		serverCfg = nil // 如果获取失败，使用默认值
	}

	// 根据配置决定是否启用拦截器
	var interceptors []grpc.UnaryServerInterceptor

	// 追踪拦截器（始终启用，用于 traceID/requestID）
	interceptors = append(interceptors, interceptor.TraceUnaryInterceptor())

	// Metrics 拦截器（根据配置）
	if serverCfg != nil && serverCfg.Features.Metrics.Enabled {
		interceptors = append(interceptors, interceptor.MetricsUnaryInterceptor())
	} else if serverCfg == nil {
		// 如果配置不存在，默认启用 metrics
		interceptors = append(interceptors, interceptor.MetricsUnaryInterceptor())
	}

	// 创建 gRPC 服务器选项
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),
	}

	// 添加消息大小限制（如果配置存在且启用）
	if serverCfg != nil && serverCfg.GRPC.SizeLimit.Enabled {
		serverOpts = append(serverOpts,
			grpc.MaxRecvMsgSize(serverCfg.GRPC.SizeLimit.MaxRecvMsgSize),
			grpc.MaxSendMsgSize(serverCfg.GRPC.SizeLimit.MaxSendMsgSize),
		)
	} else if serverCfg == nil {
		// 如果配置不存在，使用默认值（4MB）
		serverOpts = append(serverOpts,
			grpc.MaxRecvMsgSize(4*1024*1024), // 4MB
			grpc.MaxSendMsgSize(4*1024*1024), // 4MB
		)
	}

	// 创建 gRPC 服务器
	s := grpc.NewServer(serverOpts...)

	// 批量注册所有服务到 gRPC 服务器
	a.serviceRegistry.RegisterToGRPCServer(s)

	serviceCount := len(a.serviceRegistry.GetServices())
	log.Info("gRPC server listening",
		zap.String("service", a.ServiceName),
		zap.String("address", grpcCfg.FullAddress()),
		zap.Int("service_count", serviceCount),
	)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	return s, nil
}

// GracefulStop 优雅停止服务器
func (s *Server) GracefulStop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// GracefulStopWithTimeout 带超时的优雅停止服务器
func (s *Server) GracefulStopWithTimeout(timeout time.Duration) {
	if s.grpcServer == nil {
		return
	}

	// 使用 channel 来等待 GracefulStop 完成
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	// 等待完成或超时
	select {
	case <-done:
		log.Info("gRPC server stopped gracefully", zap.String("service", s.app.ServiceName))
	case <-time.After(timeout):
		log.Warn("gRPC server graceful stop timeout, forcing stop",
			zap.String("service", s.app.ServiceName),
			zap.Duration("timeout", timeout))
		s.grpcServer.Stop()
	}
}
