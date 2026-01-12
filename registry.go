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

// DefaultConfigRegistry 配置注册表，管理所有配置模块
// 实现 ConfigRegistry 接口
// 这是通用的实现，可以被多个服务复用
type DefaultConfigRegistry struct {
	Mysql         *db.MySQLConfig          `yaml:"mysql"`
	Redis         *db.RedisConfig          `yaml:"redis"`
	ClickHouse    *clickhouse.Config       `yaml:"clickhouse"`
	Email         *email.Config            `yaml:"email"`
	Server        *ServerConfig            `yaml:"server"`
	Log           *log.Config              `yaml:"log"`
	XxlJob        *xxljob.Config           `yaml:"xxl_job"`
	Elasticsearch *elasticsearch.Config    `yaml:"elasticsearch"`
	MongoDB       *mongodb.Config          `yaml:"mongodb"`
	PostgreSQL    *db.PostgreSQLConfig     `yaml:"postgresql"`
	OSS           *oss.Config              `yaml:"oss"`
	WebSocket     *websocket.Config        `yaml:"websocket"`
	MessageQueue  *messagequeue.Config     `yaml:"message_queue"`
	ConfigCenter  *configcenter.Config     `yaml:"config_center"`
	Services      map[string]ServiceConfig `yaml:"services"` // RPC 服务地址配置
}

// 确保 DefaultConfigRegistry 实现 ConfigRegistry 接口
var _ ConfigRegistry = (*DefaultConfigRegistry)(nil)

// GetMySQL 获取 MySQL 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetMySQL() (*db.MySQLConfig, error) {
	if r.Mysql == nil {
		return nil, fmt.Errorf("mysql config not loaded")
	}
	return r.Mysql, nil
}

// MustGetMySQL 获取 MySQL 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetMySQL() *db.MySQLConfig {
	cfg, err := r.GetMySQL()
	if err != nil {
		panic(err)
	}
	return cfg
}

// GetRedis 获取 Redis 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetRedis() (*db.RedisConfig, error) {
	if r.Redis == nil {
		return nil, fmt.Errorf("redis config not loaded")
	}
	return r.Redis, nil
}

// MustGetRedis 获取 Redis 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetRedis() *db.RedisConfig {
	cfg, err := r.GetRedis()
	if err != nil {
		panic(err)
	}
	return cfg
}

// GetServer 获取服务器配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetServer() (*ServerConfig, error) {
	if r.Server == nil {
		return nil, fmt.Errorf("server config not loaded")
	}
	return r.Server, nil
}

// MustGetServer 获取服务器配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetServer() *ServerConfig {
	cfg, err := r.GetServer()
	if err != nil {
		panic(err)
	}
	return cfg
}

// GetLog 获取日志配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetLog() (*log.Config, error) {
	if r.Log == nil {
		return nil, fmt.Errorf("log config not loaded")
	}
	return r.Log, nil
}

// MustGetLog 获取日志配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetLog() *log.Config {
	cfg, err := r.GetLog()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasMySQL 检查是否已加载 MySQL 配置
func (r *DefaultConfigRegistry) HasMySQL() bool {
	return r.Mysql != nil
}

// HasRedis 检查是否已加载 Redis 配置
func (r *DefaultConfigRegistry) HasRedis() bool {
	return r.Redis != nil
}

// HasServer 检查是否已加载服务器配置
func (r *DefaultConfigRegistry) HasServer() bool {
	return r.Server != nil
}

// HasLog 检查是否已加载日志配置
func (r *DefaultConfigRegistry) HasLog() bool {
	return r.Log != nil
}

// GetClickHouse 获取 ClickHouse 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetClickHouse() (*clickhouse.Config, error) {
	if r.ClickHouse == nil {
		return nil, fmt.Errorf("clickhouse config not loaded")
	}
	return r.ClickHouse, nil
}

// MustGetClickHouse 获取 ClickHouse 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetClickHouse() *clickhouse.Config {
	cfg, err := r.GetClickHouse()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasClickHouse 检查是否已加载 ClickHouse 配置
func (r *DefaultConfigRegistry) HasClickHouse() bool {
	return r.ClickHouse != nil
}

// GetEmail 获取邮件配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetEmail() (*email.Config, error) {
	if r.Email == nil {
		return nil, fmt.Errorf("email config not loaded")
	}
	return r.Email, nil
}

// MustGetEmail 获取邮件配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetEmail() *email.Config {
	cfg, err := r.GetEmail()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasEmail 检查是否已加载邮件配置
func (r *DefaultConfigRegistry) HasEmail() bool {
	return r.Email != nil
}

// GetXxlJob 获取 XXL-JOB 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetXxlJob() (*xxljob.Config, error) {
	if r.XxlJob == nil {
		return nil, fmt.Errorf("xxl_job config not loaded")
	}
	return r.XxlJob, nil
}

// MustGetXxlJob 获取 XXL-JOB 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetXxlJob() *xxljob.Config {
	cfg, err := r.GetXxlJob()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasXxlJob 检查是否已加载 XXL-JOB 配置
func (r *DefaultConfigRegistry) HasXxlJob() bool {
	return r.XxlJob != nil
}

// GetElasticsearch 获取 Elasticsearch 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetElasticsearch() (*elasticsearch.Config, error) {
	if r.Elasticsearch == nil {
		return nil, fmt.Errorf("elasticsearch config not loaded")
	}
	return r.Elasticsearch, nil
}

// GetMongoDB 获取 MongoDB 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetMongoDB() (*mongodb.Config, error) {
	if r.MongoDB == nil {
		return nil, fmt.Errorf("mongodb config not loaded")
	}
	return r.MongoDB, nil
}

// MustGetMongoDB 获取 MongoDB 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetMongoDB() *mongodb.Config {
	cfg, err := r.GetMongoDB()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasMongoDB 检查是否已加载 MongoDB 配置
func (r *DefaultConfigRegistry) HasMongoDB() bool {
	return r.MongoDB != nil
}

// GetPostgreSQL 获取 PostgreSQL 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetPostgreSQL() (*db.PostgreSQLConfig, error) {
	if r.PostgreSQL == nil {
		return nil, fmt.Errorf("postgresql config not loaded")
	}
	return r.PostgreSQL, nil
}

// MustGetPostgreSQL 获取 PostgreSQL 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetPostgreSQL() *db.PostgreSQLConfig {
	cfg, err := r.GetPostgreSQL()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasPostgreSQL 检查是否已加载 PostgreSQL 配置
func (r *DefaultConfigRegistry) HasPostgreSQL() bool {
	return r.PostgreSQL != nil
}

// GetOSS 获取对象存储配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetOSS() (*oss.Config, error) {
	if r.OSS == nil {
		return nil, fmt.Errorf("oss config not found")
	}
	return r.OSS, nil
}

// MustGetOSS 获取对象存储配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetOSS() *oss.Config {
	cfg, err := r.GetOSS()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasOSS 检查是否已加载对象存储配置
func (r *DefaultConfigRegistry) HasOSS() bool {
	return r.OSS != nil
}

// GetWebSocket 获取 WebSocket 配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetWebSocket() (*websocket.Config, error) {
	if r.WebSocket == nil {
		return nil, fmt.Errorf("websocket config not found")
	}
	return r.WebSocket, nil
}

// MustGetWebSocket 获取 WebSocket 配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetWebSocket() *websocket.Config {
	cfg, err := r.GetWebSocket()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasWebSocket 检查是否已加载 WebSocket 配置
func (r *DefaultConfigRegistry) HasWebSocket() bool {
	return r.WebSocket != nil
}

// GetMessageQueue 获取消息队列配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetMessageQueue() (*messagequeue.Config, error) {
	if r.MessageQueue == nil {
		return nil, fmt.Errorf("message queue config not loaded")
	}
	return r.MessageQueue, nil
}

// MustGetMessageQueue 获取消息队列配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetMessageQueue() *messagequeue.Config {
	cfg, err := r.GetMessageQueue()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasMessageQueue 检查是否已加载消息队列配置
func (r *DefaultConfigRegistry) HasMessageQueue() bool {
	return r.MessageQueue != nil
}

// GetConfigCenter 获取配置中心配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetConfigCenter() (*configcenter.Config, error) {
	if r.ConfigCenter == nil {
		return nil, fmt.Errorf("config center config not loaded")
	}
	return r.ConfigCenter, nil
}

// MustGetConfigCenter 获取配置中心配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetConfigCenter() *configcenter.Config {
	cfg, err := r.GetConfigCenter()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasConfigCenter 检查是否已加载配置中心配置
func (r *DefaultConfigRegistry) HasConfigCenter() bool {
	return r.ConfigCenter != nil
}

// GetServices 获取 RPC 服务地址配置（实现 ConfigRegistry 接口）
func (r *DefaultConfigRegistry) GetServices() (map[string]ServiceConfig, error) {
	if r.Services == nil {
		return nil, fmt.Errorf("services config not loaded")
	}
	return r.Services, nil
}

// MustGetServices 获取 RPC 服务地址配置，如果不存在则 panic
func (r *DefaultConfigRegistry) MustGetServices() map[string]ServiceConfig {
	cfg, err := r.GetServices()
	if err != nil {
		panic(err)
	}
	return cfg
}

// HasServices 检查是否已加载服务地址配置
func (r *DefaultConfigRegistry) HasServices() bool {
	return len(r.Services) > 0
}

// SetFieldByPath 根据路径设置字段值（通用实现）
// 这是通用的字段设置逻辑，可以被所有服务复用
func SetFieldByPath(registry *DefaultConfigRegistry, module, fieldPath, value string) error {
	var cfg interface{}
	switch module {
	case "mysql":
		if registry.Mysql == nil {
			registry.Mysql = &db.MySQLConfig{}
		}
		cfg = registry.Mysql
	case "redis":
		if registry.Redis == nil {
			registry.Redis = &db.RedisConfig{}
		}
		cfg = registry.Redis
	case "server":
		if registry.Server == nil {
			registry.Server = &ServerConfig{}
		}
		cfg = registry.Server
	case "log":
		if registry.Log == nil {
			registry.Log = &log.Config{}
		}
		cfg = registry.Log
	case "xxl_job":
		if registry.XxlJob == nil {
			registry.XxlJob = &xxljob.Config{}
		}
		cfg = registry.XxlJob
	case "clickhouse":
		if registry.ClickHouse == nil {
			registry.ClickHouse = &clickhouse.Config{}
		}
		cfg = registry.ClickHouse
	case "email":
		if registry.Email == nil {
			registry.Email = &email.Config{}
		}
		cfg = registry.Email
	case "elasticsearch":
		if registry.Elasticsearch == nil {
			registry.Elasticsearch = &elasticsearch.Config{}
		}
		cfg = registry.Elasticsearch
	case "mongodb":
		if registry.MongoDB == nil {
			registry.MongoDB = &mongodb.Config{}
		}
		cfg = registry.MongoDB
	case "postgresql":
		if registry.PostgreSQL == nil {
			registry.PostgreSQL = &db.PostgreSQLConfig{}
		}
		cfg = registry.PostgreSQL
	case "oss":
		if registry.OSS == nil {
			registry.OSS = &oss.Config{}
		}
		cfg = registry.OSS
	case "websocket":
		if registry.WebSocket == nil {
			registry.WebSocket = &websocket.Config{}
		}
		cfg = registry.WebSocket
	case "message_queue":
		if registry.MessageQueue == nil {
			registry.MessageQueue = &messagequeue.Config{}
		}
		cfg = registry.MessageQueue
	case "features":
		// 处理 server.features 配置
		if registry.Server == nil {
			registry.Server = &ServerConfig{}
		}
		// features 是 Server 的嵌套字段，需要特殊处理
		// 这里直接使用 Server 作为目标，fieldPath 已经去掉了 "features." 前缀
		cfg = registry.Server
	case "config_center":
		// 处理配置中心配置
		if registry.ConfigCenter == nil {
			registry.ConfigCenter = &configcenter.Config{}
		}
		cfg = registry.ConfigCenter
	case "services":
		// 处理服务地址配置
		if registry.Services == nil {
			registry.Services = make(map[string]ServiceConfig)
		}
		// services 是 map 类型，需要特殊处理
		// 这里直接使用 registry 作为目标，fieldPath 格式为 "service_name.address"
		return pkgConfig.SetFieldByPath(registry.Services, fieldPath, value)
	default:
		return fmt.Errorf("unknown module: %s", module)
	}

	return pkgConfig.SetFieldByPath(cfg, fieldPath, value)
}
