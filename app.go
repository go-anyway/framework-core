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
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	clickhouse "github.com/go-anyway/framework-clickhouse"
	elasticsearch "github.com/go-anyway/framework-elasticsearch"
	email "github.com/go-anyway/framework-email"
	log "github.com/go-anyway/framework-log"
	messagequeue "github.com/go-anyway/framework-messagequeue"
	mongodb "github.com/go-anyway/framework-mongodb"
	oss "github.com/go-anyway/framework-oss"
	ratelimit "github.com/go-anyway/framework-ratelimit"
	xxljob "github.com/go-anyway/framework-xxljob"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用主结构体，显式声明所有依赖
type App struct {
	// 服务名（从配置中读取）
	ServiceName string

	// 配置（接口类型）
	Config ConfigRegistry

	// 基础设施依赖
	MySQL         *gorm.DB
	PostgreSQL    *gorm.DB
	Redis         *redis.Client
	MessageQueue  *messagequeue.Client
	ClickHouse    *clickhouse.ClickHouseClient
	Elasticsearch *elasticsearch.ElasticsearchClient
	Email         *email.EmailClient
	XxlJob        xxljob.Executor
	MongoDB       *mongodb.MongoDBClient
	OSS           *oss.StoreFactory

	// 模块注册表
	moduleRegistry *ModuleRegistry

	// 服务注册表
	serviceRegistry *ServiceRegistry

	// 生命周期管理
	lifecycle *Lifecycle

	// 依赖清单
	dependencies DependencyList

	// 类型安全的依赖容器
	dependencyContainer *DependencyContainer

	// HTTP 服务器相关（用于网关或 HTTP 服务）
	Router     *gin.Engine
	Handler    interface{} // gateway.GatewayHandler，使用 interface{} 避免循环依赖
	HTTPServer *http.Server

	// 可热更新的组件（限流器、熔断器等）
	RateLimitManager ratelimit.RateLimitManager // 业务限流管理器
}

// GetDependencies 获取依赖清单
func (a *App) GetDependencies() DependencyList {
	return a.dependencies
}

// Shutdown 关闭应用，清理所有资源
func (a *App) Shutdown() error {
	return a.lifecycle.Shutdown()
}

// GetMySQL 获取 MySQL 数据库连接
func (a *App) GetMySQL() *gorm.DB {
	return a.MySQL
}

// GetRedis 获取 Redis 客户端连接
func (a *App) GetRedis() *redis.Client {
	return a.Redis
}

// GetMessageQueue 获取消息队列客户端
func (a *App) GetMessageQueue() *messagequeue.Client {
	return a.MessageQueue
}

// GetClickHouse 获取 ClickHouse 客户端连接
func (a *App) GetClickHouse() *clickhouse.ClickHouseClient {
	return a.ClickHouse
}

// GetElasticsearch 获取 Elasticsearch 客户端连接
func (a *App) GetElasticsearch() *elasticsearch.ElasticsearchClient {
	return a.Elasticsearch
}

// GetEmail 获取邮件客户端连接
func (a *App) GetEmail() *email.EmailClient {
	return a.Email
}

// GetXxlJob 获取 XXL-JOB 执行器
func (a *App) GetXxlJob() xxljob.Executor {
	return a.XxlJob
}

// GetMongoDB 获取 MongoDB 客户端连接
func (a *App) GetMongoDB() *mongodb.MongoDBClient {
	return a.MongoDB
}

// GetPostgreSQL 获取 PostgreSQL 数据库连接
func (a *App) GetPostgreSQL() *gorm.DB {
	return a.PostgreSQL
}

// GetOSS 获取对象存储工厂
func (a *App) GetOSS() *oss.StoreFactory {
	return a.OSS
}

// GetModuleRegistry 获取模块注册表
func (a *App) GetModuleRegistry() *ModuleRegistry {
	return a.moduleRegistry
}

// GetServiceRegistry 获取服务注册表
func (a *App) GetServiceRegistry() *ServiceRegistry {
	return a.serviceRegistry
}

// GetDependencyContainer 获取依赖容器
func (a *App) GetDependencyContainer() *DependencyContainer {
	if a.dependencyContainer == nil {
		a.dependencyContainer = NewDependencyContainer()
	}
	return a.dependencyContainer
}

// GetRouter 获取路由引擎
func (a *App) GetRouter() *gin.Engine {
	return a.Router
}

// StartHTTPServer 启动 HTTP 服务器
func (a *App) StartHTTPServer() error {
	serverCfg, err := a.Config.GetServer()
	if err != nil {
		return err
	}

	if serverCfg.HTTP == nil {
		return errors.New("HTTP config not found")
	}

	if a.Router == nil {
		return errors.New("router not initialized, please call WithHTTP first")
	}

	address := serverCfg.HTTP.Address()
	log.Info("HTTP server starting", zap.String("address", serverCfg.HTTP.FullAddress()))

	// 先尝试监听端口，确保启动错误在主线程中处理
	listener, err := net.Listen("tcp", address)
	if err != nil {
		// 检查是否是端口占用错误
		if isPortInUseError(err) {
			log.Fatal("Port is already in use",
				zap.String("address", serverCfg.HTTP.FullAddress()),
				zap.String("port", serverCfg.HTTP.Port),
				zap.Error(err),
				zap.String("hint", "Please check if another process is using this port or change the port in config"))
		}
		return err
	}

	server := &http.Server{
		Addr:              address,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second, // 防止 Slowloris 攻击
	}

	a.HTTPServer = server

	// 注册关闭函数
	a.lifecycle.RegisterShutdown(func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error("HTTP server shutdown timeout, forcing close", zap.Error(err))
				_ = server.Close()
				return err
			}
			log.Error("Failed to shutdown HTTP server", zap.Error(err))
			return err
		}
		log.Info("HTTP server stopped gracefully")
		return nil
	})

	// 启动服务器（在 goroutine 中运行）
	go func() {
		log.Info("HTTP server listening", zap.String("address", serverCfg.HTTP.FullAddress()))
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error", zap.Error(err))
		}
	}()

	log.Info("HTTP server started successfully, waiting for shutdown signal...")
	a.lifecycle.SetStarted()

	return nil
}

// isPortInUseError 检查错误是否是端口占用错误
func isPortInUseError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	keywords := []string{
		"address already in use",
		"bind: address already in use",
		"port is already allocated",
		"only one usage of each socket address",
		"address is already in use",
		"listen tcp",
	}
	for _, keyword := range keywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}
