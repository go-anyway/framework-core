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
	"strconv"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	errors []error
}

// NewConfigValidator 创建新的配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make([]error, 0),
	}
}

// Validate 验证配置注册表
func (v *ConfigValidator) Validate(cfg ConfigRegistry) error {
	v.errors = make([]error, 0)

	// 验证服务器配置
	if serverCfg, err := cfg.GetServer(); err == nil {
		v.validateServerConfig(serverCfg)
	}

	// 验证 MySQL 配置（如果存在）
	if mysqlCfg, err := cfg.GetMySQL(); err == nil {
		if err := mysqlCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("mysql validation failed: %w", err))
		}
	}

	// 验证 Redis 配置（如果存在）
	if redisCfg, err := cfg.GetRedis(); err == nil {
		if err := redisCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("redis validation failed: %w", err))
		}
	}

	// 验证消息队列配置（如果存在）
	if mqConfig, err := cfg.GetMessageQueue(); err == nil {
		if err := mqConfig.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("message_queue validation failed: %w", err))
		}
	}

	// 验证 ClickHouse 配置（如果存在）
	if clickhouseCfg, err := cfg.GetClickHouse(); err == nil {
		if err := clickhouseCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("clickhouse validation failed: %w", err))
		}
	}

	// 验证 Email 配置（如果存在）
	if emailCfg, err := cfg.GetEmail(); err == nil {
		if err := emailCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("email validation failed: %w", err))
		}
	}

	// 验证 MongoDB 配置（如果存在）
	if mongodbCfg, err := cfg.GetMongoDB(); err == nil {
		if err := mongodbCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("mongodb validation failed: %w", err))
		}
	}

	// 验证 PostgreSQL 配置（如果存在）
	if postgresqlCfg, err := cfg.GetPostgreSQL(); err == nil {
		if err := postgresqlCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("postgresql validation failed: %w", err))
		}
	}

	// 验证 OSS 配置（如果存在）
	if ossCfg, err := cfg.GetOSS(); err == nil {
		if err := ossCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("oss validation failed: %w", err))
		}
	}

	// 验证 WebSocket 配置（如果存在）
	if wsCfg, err := cfg.GetWebSocket(); err == nil {
		if err := wsCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("websocket validation failed: %w", err))
		}
	}

	// 验证 Elasticsearch 配置（如果存在）
	if esCfg, err := cfg.GetElasticsearch(); err == nil {
		if err := esCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("elasticsearch validation failed: %w", err))
		}
	}

	// 验证 XxlJob 配置（如果存在）
	if xxlJobCfg, err := cfg.GetXxlJob(); err == nil {
		if err := xxlJobCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("xxl_job validation failed: %w", err))
		}
	}

	// 验证配置中心配置（如果存在）
	if configCenterCfg, err := cfg.GetConfigCenter(); err == nil {
		if err := configCenterCfg.Validate(); err != nil {
			v.errors = append(v.errors, fmt.Errorf("config_center validation failed: %w", err))
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("config validation failed: %v", v.errors)
	}

	return nil
}

// validateServerConfig 验证服务器配置
func (v *ConfigValidator) validateServerConfig(cfg *ServerConfig) {
	if cfg.ServiceName == "" {
		v.errors = append(v.errors, fmt.Errorf("server.service_name is required"))
	}

	// 验证 gRPC 端口
	if cfg.GRPC.Port != "" {
		if port, err := strconv.Atoi(cfg.GRPC.Port); err != nil {
			v.errors = append(v.errors, fmt.Errorf("server.grpc.port must be a valid integer: %w", err))
		} else if port < 1 || port > 65535 {
			v.errors = append(v.errors, fmt.Errorf("server.grpc.port must be between 1 and 65535, got %d", port))
		}
	}
}

// ValidateConfig 验证配置的便捷函数
func ValidateConfig(cfg ConfigRegistry) error {
	validator := NewConfigValidator()
	return validator.Validate(cfg)
}
