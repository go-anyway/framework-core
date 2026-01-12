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

// DependencyType 依赖类型
type DependencyType string

const (
	// DependencyMySQL MySQL 数据库依赖
	DependencyMySQL DependencyType = "mysql"
	// DependencyRedis Redis 缓存依赖
	DependencyRedis DependencyType = "redis"
	// DependencyKafka Kafka 消息队列依赖
	DependencyKafka DependencyType = "kafka"
	// DependencyMQ MQ 消息队列依赖
	DependencyMQ DependencyType = "mq"
	// DependencyClickHouse ClickHouse 数据库依赖
	DependencyClickHouse DependencyType = "clickhouse"
	// DependencyEmail 邮件服务依赖
	DependencyEmail DependencyType = "email"
	// DependencyElasticsearch Elasticsearch 搜索依赖
	DependencyElasticsearch DependencyType = "elasticsearch"
	// DependencyLog 日志依赖
	DependencyLog DependencyType = "log"
	// DependencyXxlJob XXL-JOB 任务调度依赖
	DependencyXxlJob DependencyType = "xxl_job"
	// DependencyMongoDB MongoDB 数据库依赖
	DependencyMongoDB DependencyType = "mongodb"
	// DependencyPostgreSQL PostgreSQL 数据库依赖
	DependencyPostgreSQL DependencyType = "postgresql"
	// DependencyOSS 对象存储依赖（支持 AWS S3、阿里云 OSS、腾讯云 COS、MinIO）
	DependencyOSS DependencyType = "oss"
	// DependencyConfigCenter 配置中心依赖（支持 Apollo、Nacos、Consul）
	DependencyConfigCenter DependencyType = "config_center"
	// DependencyWebSocket WebSocket 服务器依赖
	DependencyWebSocket DependencyType = "websocket"
	// DependencyDistributedLock 分布式锁依赖
	DependencyDistributedLock DependencyType = "distributed_lock"
	// DependencyLocalCache 本地缓存依赖
	DependencyLocalCache DependencyType = "local_cache"
	// DependencyCron Cron 定时任务依赖
	DependencyCron DependencyType = "cron"
	// DependencyRateLimit 分布式限流依赖
	DependencyRateLimit DependencyType = "rate_limit"
)

// Dependency 依赖信息
type Dependency struct {
	Type        DependencyType `json:"type"`
	Name        string         `json:"name"`
	Initialized bool           `json:"initialized"`
	Error       error          `json:"error,omitempty"`
}

// DependencyList 依赖清单
type DependencyList []Dependency

// HasDependency 检查是否包含指定类型的依赖
func (dl DependencyList) HasDependency(depType DependencyType) bool {
	for _, dep := range dl {
		if dep.Type == depType && dep.Initialized {
			return true
		}
	}
	return false
}

// GetDependency 获取指定类型的依赖
func (dl DependencyList) GetDependency(depType DependencyType) *Dependency {
	for i := range dl {
		if dl[i].Type == depType {
			return &dl[i]
		}
	}
	return nil
}

// String 返回依赖清单的字符串表示
func (dl DependencyList) String() string {
	var result string
	for i, dep := range dl {
		if i > 0 {
			result += ", "
		}
		status := "✓"
		if !dep.Initialized {
			status = "✗"
		}
		result += string(dep.Type) + ":" + dep.Name + "[" + status + "]"
	}
	return result
}
