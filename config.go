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
	"github.com/go-anyway/framework-clickhouse"
	pkgConfig "github.com/go-anyway/framework-config"
	"github.com/go-anyway/framework-configcenter"
	"github.com/go-anyway/framework-db"
	"github.com/go-anyway/framework-elasticsearch"
	"github.com/go-anyway/framework-email"
	"github.com/go-anyway/framework-log"
	"github.com/go-anyway/framework-messagequeue"
	"github.com/go-anyway/framework-mongodb"
	"github.com/go-anyway/framework-oss"
	"github.com/go-anyway/framework-websocket"
	"github.com/go-anyway/framework-xxljob"
)

// ConfigRegistry 配置注册表接口，所有服务的配置注册表都应实现此接口
type ConfigRegistry interface {
	GetMySQL() (*db.MySQLConfig, error)
	GetRedis() (*db.RedisConfig, error)
	GetClickHouse() (*clickhouse.Config, error)
	GetEmail() (*email.Config, error)
	GetServer() (*ServerConfig, error)
	GetLog() (*log.Config, error)
	GetXxlJob() (*xxljob.Config, error)
	GetElasticsearch() (*elasticsearch.Config, error)
	GetMongoDB() (*mongodb.Config, error)
	GetPostgreSQL() (*db.PostgreSQLConfig, error)
	GetOSS() (*oss.Config, error)
	GetWebSocket() (*websocket.Config, error)
	GetMessageQueue() (*messagequeue.Config, error)
	GetConfigCenter() (*configcenter.Config, error)
	GetServices() (map[string]ServiceConfig, error)
}

// ServerConfig 服务器配置
type ServerConfig struct {
	ServiceName string                   `yaml:"service_name" env:"SERVICE_NAME" required:"true"`
	HTTP        *HTTPConfig              `yaml:"http,omitempty"` // 可选，用于 HTTP 服务
	GRPC        GRPCConfig               `yaml:"grpc"`           // 可选，用于 RPC 服务
	Features    pkgConfig.FeaturesConfig `yaml:"features"`       // 统一使用 FeaturesConfig 支持网关和 RPC 两种模式
}

// GRPCSizeLimitConfig gRPC 消息大小限制配置
type GRPCSizeLimitConfig struct {
	Enabled        bool `yaml:"enabled" env:"GRPC_SIZE_LIMIT_ENABLED" default:"true"`
	MaxRecvMsgSize int  `yaml:"max_recv_msg_size" env:"GRPC_MAX_RECV_MSG_SIZE" default:"4194304"` // 默认 4MB
	MaxSendMsgSize int  `yaml:"max_send_msg_size" env:"GRPC_MAX_SEND_MSG_SIZE" default:"4194304"` // 默认 4MB
}

// Duration 配置中的 Duration 类型（使用 pkgConfig.Duration）
type Duration = pkgConfig.Duration

// GRPCConfig gRPC 服务器配置
type GRPCConfig struct {
	Host      string              `yaml:"host" env:"GRPC_HOST" default:"0.0.0.0"`
	Port      string              `yaml:"port" env:"GRPC_PORT" default:"50051"`
	SizeLimit GRPCSizeLimitConfig `yaml:"size_limit"`
}

// Address 返回 gRPC 服务器地址（带冒号前缀，用于 net.Listen）
func (c *GRPCConfig) Address() string {
	return ":" + c.Port
}

// FullAddress 返回完整的 gRPC 服务器地址
func (c *GRPCConfig) FullAddress() string {
	return c.Host + ":" + c.Port
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Host string `yaml:"host" env:"HTTP_HOST" default:"0.0.0.0"`
	Port string `yaml:"port" env:"HTTP_PORT" default:"8080"`
}

// Address 返回 HTTP 服务器地址（带冒号前缀，用于 net.Listen）
func (c *HTTPConfig) Address() string {
	return ":" + c.Port
}

// FullAddress 返回完整的 HTTP 服务器地址
func (c *HTTPConfig) FullAddress() string {
	return c.Host + ":" + c.Port
}

// ServiceConfig RPC 服务配置
type ServiceConfig struct {
	Address string `yaml:"address" env:"RPC_ADDRESS" required:"true"`
}
