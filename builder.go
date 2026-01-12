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
	"fmt"
	"io"
	"strings"

	"github.com/go-anyway/framework-cache"
	"github.com/go-anyway/framework-clickhouse"
	"github.com/go-anyway/framework-configcenter"
	"github.com/go-anyway/framework-cron"
	"github.com/go-anyway/framework-db"
	"github.com/go-anyway/framework-elasticsearch"
	"github.com/go-anyway/framework-email"
	"github.com/go-anyway/framework-gateway"
	"github.com/go-anyway/framework-hotreload"
	"github.com/go-anyway/framework-log"
	"github.com/go-anyway/framework-messagequeue"
	"github.com/go-anyway/framework-metrics"
	"github.com/go-anyway/framework-mongodb"
	"github.com/go-anyway/framework-oss"
	"github.com/go-anyway/framework-ratelimit"
	"github.com/go-anyway/framework-trace"
	"github.com/go-anyway/framework-websocket"
	"github.com/go-anyway/framework-xxljob"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// DependencyInitializer 依赖初始化器函数类型
// 用于插件化依赖初始化
type DependencyInitializer func(builder *AppBuilder) error

// RouteSetupFunc 路由和服务注册回调函数类型
// router: 已配置好中间件的 Gin 路由引擎
// handler: 网关处理器，用于注册 gRPC 服务
type RouteSetupFunc func(router *gin.Engine, handler interface{}) error

// redisClientWrapper 包装 redis.Client 以实现 ratelimit.RedisClient 接口
type redisClientWrapper struct {
	client *redis.Client
}

// Eval 执行 Lua 脚本
func (w *redisClientWrapper) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	cmd := w.client.Eval(ctx, script, keys, args...)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

// AppBuilder App 构建器，提供链式初始化方法
type AppBuilder struct {
	app                *App
	customDependencies []Dependency
	httpInit           bool
	routeSetupFunc     RouteSetupFunc
	configDir          string
}

// registerDependency 注册依赖到依赖清单（公共方法）
func (b *AppBuilder) registerDependency(depType DependencyType, name string, initialized bool, err error) {
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        depType,
		Name:        name,
		Initialized: initialized,
		Error:       err,
	})
}

// handleConfigError 处理配置获取错误（公共方法）
func (b *AppBuilder) handleConfigError(depType DependencyType, name string, err error) {
	b.registerDependency(depType, name, false, fmt.Errorf("%s config not found: %w", name, err))
	log.Warn("Config not found, skipping initialization", zap.String("name", name), zap.Error(err))
}

// handleDisabledDependency 处理未启用的依赖（公共方法）
func (b *AppBuilder) handleDisabledDependency(depType DependencyType, name string) {
	b.registerDependency(depType, name, false, nil)
	log.Info("Dependency is disabled, skipping initialization", zap.String("name", name))
}

// handleInitError 处理初始化错误（公共方法）
func (b *AppBuilder) handleInitError(depType DependencyType, name string, err error) {
	b.registerDependency(depType, name, false, fmt.Errorf("failed to connect to %s: %w", name, err))
	log.Error("Failed to initialize dependency", zap.String("name", name), zap.Error(err))
}

// handleInitSuccess 处理初始化成功（公共方法）
func (b *AppBuilder) handleInitSuccess(depType DependencyType, name string) {
	b.registerDependency(depType, name, true, nil)
}

// NewAppBuilder 创建新的 App 构建器
func NewAppBuilder(cfg ConfigRegistry) *AppBuilder {
	// 从配置中读取服务名
	serviceName := "service" // 默认值
	if serverCfg, err := cfg.GetServer(); err == nil {
		serviceName = serverCfg.ServiceName
	}

	return &AppBuilder{
		app: &App{
			ServiceName:         serviceName,
			Config:              cfg,
			lifecycle:           NewLifecycle(),
			dependencies:        make(DependencyList, 0),
			serviceRegistry:     NewServiceRegistry(),
			moduleRegistry:      NewModuleRegistry(),
			dependencyContainer: NewDependencyContainer(),
		},
	}
}

// WithMySQL 初始化 MySQL 数据库连接
func (b *AppBuilder) WithMySQL() *AppBuilder {
	mysqlCfg, err := b.app.Config.GetMySQL()
	if err != nil {
		b.handleConfigError(DependencyMySQL, "MySQL", err)
		return b
	}

	// 检查是否启用
	if !mysqlCfg.Enabled {
		b.handleDisabledDependency(DependencyMySQL, "MySQL")
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := mysqlCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyMySQL, "MySQL", err)
		return b
	}

	// 禁用 GORM 的终端日志输出，使用 Silent 模式
	// 因为我们有自己的追踪日志系统
	opts.LogLevel = logger.Silent
	dbInstance, err := db.New(opts)
	if err != nil {
		b.handleInitError(DependencyMySQL, "MySQL", err)
		return b
	}

	b.app.MySQL = dbInstance
	b.handleInitSuccess(DependencyMySQL, "MySQL")

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.MySQL != nil {
			sqlDB, err := b.app.MySQL.DB()
			if err != nil {
				return fmt.Errorf("failed to get underlying sql.DB: %w", err)
			}
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("failed to close MySQL connection: %w", err)
			}
			log.Info("MySQL connection closed")
		}
		return nil
	})

	log.Info("MySQL database initialized successfully",
		zap.String("host", mysqlCfg.Host),
		zap.Int("port", mysqlCfg.Port),
		zap.String("database", mysqlCfg.Database),
	)

	return b
}

// WithRedis 初始化 Redis 连接
func (b *AppBuilder) WithRedis() *AppBuilder {
	redisCfg, err := b.app.Config.GetRedis()
	if err != nil {
		b.handleConfigError(DependencyRedis, "Redis", err)
		return b
	}

	// 检查是否启用
	if !redisCfg.Enabled {
		b.handleDisabledDependency(DependencyRedis, "Redis")
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := redisCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyRedis, "Redis", err)
		return b
	}

	// 启用命令追踪
	redisClient, err := db.NewRedis(opts)
	if err != nil {
		b.handleInitError(DependencyRedis, "Redis", err)
		return b
	}

	b.app.Redis = redisClient
	b.handleInitSuccess(DependencyRedis, "Redis")

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.Redis != nil {
			if err := b.app.Redis.Close(); err != nil {
				return fmt.Errorf("failed to close Redis connection: %w", err)
			}
			log.Info("Redis connection closed")
		}
		return nil
	})

	log.Info("Redis initialized successfully",
		zap.String("host", redisCfg.Host),
		zap.Int("port", redisCfg.Port),
		zap.Int("db", redisCfg.DB),
	)

	return b
}

// WithMessageQueue 统一初始化消息队列（根据配置自动选择类型）
func (b *AppBuilder) WithMessageQueue() *AppBuilder {
	mqConfig, err := b.app.Config.GetMessageQueue()
	if err != nil {
		b.handleConfigError(DependencyKafka, "MessageQueue", err)
		return b
	}

	// 如果类型为 none，跳过初始化
	if mqConfig.Type == "" || mqConfig.Type == "none" {
		log.Info("Message queue type is 'none', skipping initialization")
		return b
	}

	// 构建统一的消息队列选项
	opts, err := b.buildMessageQueueOptions(mqConfig)
	if err != nil {
		depType := getMessageQueueDependencyType(mqConfig.Type)
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        depType,
			Name:        "MessageQueue",
			Initialized: false,
			Error:       fmt.Errorf("failed to build message queue options: %w", err),
		})
		log.Error("Failed to build message queue options", zap.Error(err))
		return b
	}

	// 创建统一的消息队列客户端
	client, err := messagequeue.NewClient(opts)
	if err != nil {
		depType := getMessageQueueDependencyType(mqConfig.Type)
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        depType,
			Name:        "MessageQueue",
			Initialized: false,
			Error:       fmt.Errorf("failed to create message queue client: %w", err),
		})
		log.Error("Failed to initialize message queue", zap.Error(err))
		return b
	}

	// 执行连接自检，确保服务端可用
	if err := client.HealthCheck(); err != nil {
		depType := getMessageQueueDependencyType(mqConfig.Type)
		_ = client.Close() // 关闭已创建的客户端
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        depType,
			Name:        "MessageQueue",
			Initialized: false,
			Error:       fmt.Errorf("message queue health check failed: %w (please ensure the message queue server is running)", err),
		})
		log.Error("Message queue health check failed",
			zap.String("type", mqConfig.Type),
			zap.Error(err),
			zap.String("hint", "Please ensure the message queue server is running and accessible"),
		)
		return b
	}

	b.app.MessageQueue = client
	depType := getMessageQueueDependencyType(mqConfig.Type)
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        depType,
		Name:        "MessageQueue",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.MessageQueue != nil {
			if err := b.app.MessageQueue.Close(); err != nil {
				return fmt.Errorf("failed to close message queue connection: %w", err)
			}
			log.Info("Message queue connection closed")
		}
		return nil
	})

	log.Info("Message queue initialized successfully",
		zap.String("type", mqConfig.Type),
	)

	return b
}

// buildMessageQueueOptions 从配置构建统一的消息队列选项
func (b *AppBuilder) buildMessageQueueOptions(mqConfig *messagequeue.Config) (*messagequeue.Options, error) {
	// 直接使用 messagequeue.Config 的 ToOptions 方法
	return mqConfig.ToOptions()
}

// getMessageQueueDependencyType 根据消息队列类型获取依赖类型
func getMessageQueueDependencyType(mqType string) DependencyType {
	switch messagequeue.Type(mqType) {
	case messagequeue.TypeKafka:
		return DependencyKafka
	case messagequeue.TypeRabbitMQ:
		return DependencyMQ
	default:
		// 未知类型或 none，返回 Kafka 作为默认
		return DependencyKafka
	}
}

// WithClickHouse 初始化 ClickHouse 连接
func (b *AppBuilder) WithClickHouse() *AppBuilder {
	clickhouseCfg, err := b.app.Config.GetClickHouse()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyClickHouse,
			Name:        "ClickHouse",
			Initialized: false,
			Error:       fmt.Errorf("clickhouse config not found: %w", err),
		})
		log.Warn("ClickHouse config not found, skipping ClickHouse initialization", zap.Error(err))
		return b
	}

	// 检查是否启用
	if !clickhouseCfg.Enabled {
		log.Info("ClickHouse is disabled, skipping ClickHouse initialization")
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyClickHouse,
			Name:        "ClickHouse",
			Initialized: false,
			Error:       nil,
		})
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := clickhouseCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyClickHouse, "ClickHouse", err)
		return b
	}

	// 启用查询追踪
	clickhouseClient, err := clickhouse.NewClickHouse(opts)
	if err != nil {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyClickHouse,
			Name:        "ClickHouse",
			Initialized: false,
			Error:       fmt.Errorf("failed to connect to clickhouse: %w", err),
		})
		log.Error("Failed to initialize ClickHouse", zap.Error(err))
		return b
	}

	b.app.ClickHouse = clickhouseClient
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyClickHouse,
		Name:        "ClickHouse",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.ClickHouse != nil {
			if err := b.app.ClickHouse.Close(); err != nil {
				return fmt.Errorf("failed to close ClickHouse connection: %w", err)
			}
			log.Info("ClickHouse connection closed")
		}
		return nil
	})

	log.Info("ClickHouse initialized successfully",
		zap.String("host", clickhouseCfg.Host),
		zap.Int("port", clickhouseCfg.Port),
		zap.String("database", clickhouseCfg.Database),
	)

	return b
}

// WithElasticsearch 初始化 Elasticsearch 连接
func (b *AppBuilder) WithElasticsearch() *AppBuilder {
	esCfg, err := b.app.Config.GetElasticsearch()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyElasticsearch,
			Name:        "Elasticsearch",
			Initialized: false,
			Error:       fmt.Errorf("elasticsearch config not found: %w", err),
		})
		log.Warn("Elasticsearch config not found, skipping Elasticsearch initialization", zap.Error(err))
		return b
	}

	// 检查是否启用
	if !esCfg.Enabled {
		log.Info("Elasticsearch is disabled, skipping Elasticsearch initialization")
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyElasticsearch,
			Name:        "Elasticsearch",
			Initialized: false,
			Error:       nil,
		})
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := esCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyElasticsearch, "Elasticsearch", err)
		return b
	}

	// 启用查询追踪
	esClient, err := elasticsearch.NewElasticsearch(opts)
	if err != nil {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyElasticsearch,
			Name:        "Elasticsearch",
			Initialized: false,
			Error:       fmt.Errorf("failed to connect to elasticsearch: %w", err),
		})
		log.Error("Failed to initialize Elasticsearch", zap.Error(err))
		return b
	}

	b.app.Elasticsearch = esClient
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyElasticsearch,
		Name:        "Elasticsearch",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.Elasticsearch != nil {
			if err := b.app.Elasticsearch.Close(); err != nil {
				return fmt.Errorf("failed to close Elasticsearch connection: %w", err)
			}
			log.Info("Elasticsearch connection closed")
		}
		return nil
	})

	log.Info("Elasticsearch initialized successfully",
		zap.Strings("addresses", esCfg.Addresses),
		zap.Int("max_retries", esCfg.MaxRetries),
	)

	return b
}

// WithEmail 初始化邮件客户端
func (b *AppBuilder) WithEmail() *AppBuilder {
	emailCfg, err := b.app.Config.GetEmail()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyEmail,
			Name:        "Email",
			Initialized: false,
			Error:       fmt.Errorf("email config not found: %w", err),
		})
		log.Warn("Email config not found, skipping Email initialization", zap.Error(err))
		return b
	}

	// 检查是否启用
	if !emailCfg.Enabled {
		log.Info("Email is disabled, skipping Email initialization")
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyEmail,
			Name:        "Email",
			Initialized: false,
			Error:       nil,
		})
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := emailCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyEmail, "Email", err)
		return b
	}

	// 启用邮件追踪
	emailClient, err := email.NewEmail(opts)
	if err != nil {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyEmail,
			Name:        "Email",
			Initialized: false,
			Error:       fmt.Errorf("failed to create email client: %w", err),
		})
		log.Error("Failed to initialize Email", zap.Error(err))
		return b
	}

	b.app.Email = emailClient
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyEmail,
		Name:        "Email",
		Initialized: true,
		Error:       nil,
	})

	// 邮件客户端不需要关闭连接（每次发送都是新连接）
	log.Info("Email initialized successfully",
		zap.String("host", emailCfg.Host),
		zap.Int("port", emailCfg.Port),
		zap.String("from", emailCfg.From),
	)

	return b
}

// WithXxlJob 初始化 XXL-JOB 执行器
// middlewares: 可选的中间件列表，按顺序应用
// 使用示例：
//
//	WithXxlJob()  // 无中间件
//	WithXxlJob(xxljob.RecoveryMiddleware(), xxljob.TimeoutMiddleware(30*time.Second))
func (b *AppBuilder) WithXxlJob(middlewares ...xxljob.Middleware) *AppBuilder {
	xxlJobCfg, err := b.app.Config.GetXxlJob()
	if err != nil {
		b.handleConfigError(DependencyXxlJob, "XXL-JOB", err)
		return b
	}

	if !xxlJobCfg.Enabled {
		log.Info("XXL-JOB is disabled, skipping initialization")
		return b
	}

	// 使用 NewFromConfig 创建执行器
	executor, err := xxljob.NewFromConfig(xxlJobCfg)
	if err != nil {
		b.handleInitError(DependencyXxlJob, "XXL-JOB", err)
		return b
	}

	// 注意：NewFromConfig 不支持中间件，如果需要中间件，需要使用 Builder 模式
	// 这里为了保持一致性，如果提供了中间件，需要重新构建
	if len(middlewares) > 0 {
		// 使用 Builder 模式创建执行器，并添加中间件
		executorBuilder := xxljob.NewExecutorBuilder().
			ServerAddr(xxlJobCfg.ServerAddr).
			RegistryKey(xxlJobCfg.RegistryKey).
			ExecutorPort(xxlJobCfg.ExecutorPort).
			LogPath(xxlJobCfg.LogPath).
			LogRetentionDays(xxlJobCfg.LogRetentionDays).
			Trace(xxlJobCfg.EnableTrace).
			QuietMode(xxlJobCfg.QuietMode)

		// 可选配置
		if xxlJobCfg.AccessToken != "" {
			executorBuilder = executorBuilder.AccessToken(xxlJobCfg.AccessToken)
		}
		if xxlJobCfg.ExecutorIP != "" {
			executorBuilder = executorBuilder.ExecutorIP(xxlJobCfg.ExecutorIP)
		}

		// 添加中间件
		executorBuilder = executorBuilder.Middlewares(middlewares...)

		// 构建执行器
		executor, err = executorBuilder.Build()
		if err != nil {
			b.handleInitError(DependencyXxlJob, "XXL-JOB", err)
			return b
		}
	}

	b.app.XxlJob = executor
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyXxlJob,
		Name:        "XXL-JOB",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.XxlJob != nil {
			if err := b.app.XxlJob.Stop(); err != nil {
				return fmt.Errorf("failed to stop XXL-JOB executor: %w", err)
			}
			log.Info("XXL-JOB executor stopped")
		}
		return nil
	})

	log.Info("XXL-JOB initialized successfully",
		zap.String("server_addr", xxlJobCfg.ServerAddr),
		zap.String("registry_key", xxlJobCfg.RegistryKey),
		zap.String("executor_port", xxlJobCfg.ExecutorPort),
		zap.Int("middleware_count", len(middlewares)),
	)

	return b
}

// WithMongoDB 初始化 MongoDB 连接
func (b *AppBuilder) WithMongoDB() *AppBuilder {
	mongodbCfg, err := b.app.Config.GetMongoDB()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyMongoDB,
			Name:        "MongoDB",
			Initialized: false,
			Error:       fmt.Errorf("mongodb config not found: %w", err),
		})
		log.Warn("MongoDB config not found, skipping MongoDB initialization", zap.Error(err))
		return b
	}

	// 检查是否启用
	if !mongodbCfg.Enabled {
		log.Info("MongoDB is disabled, skipping MongoDB initialization")
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyMongoDB,
			Name:        "MongoDB",
			Initialized: false,
			Error:       nil,
		})
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := mongodbCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyMongoDB, "MongoDB", err)
		return b
	}

	// 启用操作追踪
	mongodbClient, err := mongodb.NewMongoDB(opts)
	if err != nil {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyMongoDB,
			Name:        "MongoDB",
			Initialized: false,
			Error:       fmt.Errorf("failed to connect to mongodb: %w", err),
		})
		log.Error("Failed to initialize MongoDB", zap.Error(err))
		return b
	}

	b.app.MongoDB = mongodbClient
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyMongoDB,
		Name:        "MongoDB",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.MongoDB != nil {
			ctx := context.Background()
			if err := b.app.MongoDB.Close(ctx); err != nil {
				return fmt.Errorf("failed to close MongoDB connection: %w", err)
			}
			log.Info("MongoDB connection closed")
		}
		return nil
	})

	log.Info("MongoDB initialized successfully",
		zap.String("host", mongodbCfg.Host),
		zap.Int("port", mongodbCfg.Port),
		zap.String("database", mongodbCfg.Database),
	)

	return b
}

// WithPostgreSQL 初始化 PostgreSQL 数据库连接
func (b *AppBuilder) WithPostgreSQL() *AppBuilder {
	postgresqlCfg, err := b.app.Config.GetPostgreSQL()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyPostgreSQL,
			Name:        "PostgreSQL",
			Initialized: false,
			Error:       fmt.Errorf("postgresql config not found: %w", err),
		})
		log.Warn("PostgreSQL config not found, skipping PostgreSQL initialization", zap.Error(err))
		return b
	}

	// 检查是否启用
	if !postgresqlCfg.Enabled {
		log.Info("PostgreSQL is disabled, skipping PostgreSQL initialization")
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyPostgreSQL,
			Name:        "PostgreSQL",
			Initialized: false,
			Error:       nil,
		})
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := postgresqlCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyPostgreSQL, "PostgreSQL", err)
		return b
	}

	// 禁用 GORM 的终端日志输出，使用 Silent 模式
	// 因为我们有自己的追踪日志系统
	opts.LogLevel = logger.Silent
	dbInstance, err := db.NewPostgreSQL(opts)
	if err != nil {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyPostgreSQL,
			Name:        "PostgreSQL",
			Initialized: false,
			Error:       fmt.Errorf("failed to connect to postgresql: %w", err),
		})
		log.Error("Failed to initialize PostgreSQL", zap.Error(err))
		return b
	}

	b.app.PostgreSQL = dbInstance
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyPostgreSQL,
		Name:        "PostgreSQL",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.PostgreSQL != nil {
			sqlDB, err := b.app.PostgreSQL.DB()
			if err != nil {
				return fmt.Errorf("failed to get underlying sql.DB: %w", err)
			}
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
			}
			log.Info("PostgreSQL connection closed")
		}
		return nil
	})

	log.Info("PostgreSQL database initialized successfully",
		zap.String("host", postgresqlCfg.Host),
		zap.Int("port", postgresqlCfg.Port),
		zap.String("database", postgresqlCfg.Database),
		zap.String("ssl_mode", postgresqlCfg.SSLMode),
	)

	return b
}

// WithOSS 初始化对象存储连接（支持 AWS S3、阿里云 OSS、腾讯云 COS、MinIO）
func (b *AppBuilder) WithOSS() *AppBuilder {
	ossCfg, err := b.app.Config.GetOSS()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyOSS,
			Name:        "OSS",
			Initialized: false,
			Error:       fmt.Errorf("oss config not found: %w", err),
		})
		log.Warn("OSS config not found, skipping OSS initialization", zap.Error(err))
		return b
	}

	// 如果未启用，跳过初始化
	if !ossCfg.Enabled {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyOSS,
			Name:        "OSS",
			Initialized: false,
			Error:       nil,
		})
		log.Info("OSS is disabled, skipping OSS initialization")
		return b
	}

	// 使用工厂模式创建对象存储工厂
	storeFactory, err := oss.NewStoreFactory(ossCfg)
	if err != nil {
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyOSS,
			Name:        "OSS",
			Initialized: false,
			Error:       fmt.Errorf("failed to create OSS store factory: %w", err),
		})
		log.Error("Failed to initialize OSS store factory", zap.Error(err))
		return b
	}

	b.app.OSS = storeFactory
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyOSS,
		Name:        "OSS",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if b.app.OSS != nil {
			if err := b.app.OSS.Close(); err != nil {
				return fmt.Errorf("failed to close OSS connection: %w", err)
			}
		}
		return nil
	})

	log.Info("OSS store factory initialized successfully",
		zap.String("default", ossCfg.Default),
		zap.Strings("stores", storeFactory.List()),
	)

	return b
}

// WithWebSocket 初始化 WebSocket 服务器
func (b *AppBuilder) WithWebSocket() *AppBuilder {
	wsCfg, err := b.app.Config.GetWebSocket()
	if err != nil {
		// 如果配置不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyWebSocket,
			Name:        "WebSocket",
			Initialized: false,
			Error:       fmt.Errorf("websocket config not found: %w", err),
		})
		log.Warn("WebSocket config not found, skipping WebSocket initialization", zap.Error(err))
		return b
	}

	// 检查是否启用
	if !wsCfg.Enabled {
		log.Info("WebSocket is disabled, skipping WebSocket initialization")
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyWebSocket,
			Name:        "WebSocket",
			Initialized: false,
			Error:       nil,
		})
		return b
	}

	// 使用 Config.ToOptions() 方法转换配置
	opts, err := wsCfg.ToOptions()
	if err != nil {
		b.handleInitError(DependencyWebSocket, "WebSocket", err)
		return b
	}

	// 设置 CheckOrigin（默认允许所有来源，可以通过配置自定义）
	opts.CheckOrigin = func(r interface{}) bool {
		return true
	}

	wsServer := websocket.NewServer(opts)

	// 注册到依赖容器
	b.app.GetDependencyContainer().register(wsServer)

	// 注册依赖
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyWebSocket,
		Name:        "WebSocket",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		// WebSocket 服务器会在 context 取消时自动关闭
		log.Info("WebSocket server closed")
		return nil
	})

	log.Info("WebSocket server initialized successfully",
		zap.Int("read_buffer_size", wsCfg.ReadBufferSize),
		zap.Int("write_buffer_size", wsCfg.WriteBufferSize),
		zap.Int64("max_message_size", wsCfg.MaxMessageSize),
	)

	return b
}

// WithLocalCache 初始化本地缓存
func (b *AppBuilder) WithLocalCache() *AppBuilder {
	// 本地缓存不需要配置，使用默认配置
	localCache := cache.NewLocalCache(cache.DefaultOptions())

	// 注册到依赖容器
	b.app.GetDependencyContainer().register(localCache)

	// 注册依赖
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyLocalCache,
		Name:        "LocalCache",
		Initialized: true,
		Error:       nil,
	})

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		localCache.Clear()
		log.Info("Local cache cleared")
		return nil
	})

	log.Info("Local cache initialized successfully")

	return b
}

// WithCron 初始化 Cron 调度器
func (b *AppBuilder) WithCron() *AppBuilder {
	// Cron 调度器不需要配置，使用默认配置
	scheduler := cron.NewScheduler(cron.DefaultOptions())

	// 注册到依赖容器
	b.app.GetDependencyContainer().register(scheduler)

	// 注册依赖
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyCron,
		Name:        "Cron",
		Initialized: true,
		Error:       nil,
	})

	// 启动调度器
	scheduler.Start()

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		ctx := scheduler.Stop()
		<-ctx.Done()
		log.Info("Cron scheduler stopped")
		return nil
	})

	log.Info("Cron scheduler initialized successfully")

	return b
}

// WithDistributedLock 初始化分布式锁（需要 Redis）
func (b *AppBuilder) WithDistributedLock() *AppBuilder {
	if b.app.Redis == nil {
		// 如果 Redis 不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyDistributedLock,
			Name:        "DistributedLock",
			Initialized: false,
			Error:       fmt.Errorf("redis is required for distributed lock"),
		})
		log.Warn("Redis not available, skipping distributed lock initialization")
		return b
	}

	// 分布式锁通过工厂函数创建，不需要全局实例
	// 这里只注册依赖类型，实际使用时通过 lock.NewRedisLock 创建
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyDistributedLock,
		Name:        "DistributedLock",
		Initialized: true,
		Error:       nil,
	})

	log.Info("Distributed lock support initialized successfully (Redis available)")

	return b
}

// WithRateLimit 初始化分布式限流（需要 Redis）
func (b *AppBuilder) WithRateLimit() *AppBuilder {
	if b.app.Redis == nil {
		// 如果 Redis 不存在，记录但不阻止构建
		b.app.dependencies = append(b.app.dependencies, Dependency{
			Type:        DependencyRateLimit,
			Name:        "RateLimit",
			Initialized: false,
			Error:       fmt.Errorf("redis is required for rate limit"),
		})
		log.Warn("Redis not available, skipping rate limit initialization")
		return b
	}

	// 分布式限流通过工厂函数创建，不需要全局实例
	// 这里只注册依赖类型，实际使用时通过 ratelimit.NewRedisLimiter 创建
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyRateLimit,
		Name:        "RateLimit",
		Initialized: true,
		Error:       nil,
	})

	log.Info("Rate limit support initialized successfully (Redis available)")

	return b
}

// ConfigCenterOption 配置中心选项函数类型
type ConfigCenterOption func(*configCenterOptions)

// configCenterOptions 配置中心选项
type configCenterOptions struct {
	fieldSetter       func(module, fieldPath, value string) error
	allowedPrefixes   []string
	reloaders         []hotreload.Reloader
	reloaderFactories []func(*App) []hotreload.Reloader // 延迟创建重载器的工厂函数
}

// WithFieldSetter 设置字段设置器（用于系统配置热加载）
func WithFieldSetter(setter func(module, fieldPath, value string) error, prefixes []string) ConfigCenterOption {
	return func(opts *configCenterOptions) {
		opts.fieldSetter = setter
		opts.allowedPrefixes = prefixes
	}
}

// WithReloaders 设置配置重载器（用于自定义配置热加载）
func WithReloaders(reloaders ...hotreload.Reloader) ConfigCenterOption {
	return func(opts *configCenterOptions) {
		opts.reloaders = append(opts.reloaders, reloaders...)
	}
}

// WithReloaderFactories 设置重载器工厂函数（用于延迟创建重载器）
// 这些工厂函数会在 App 构建后调用，可以访问 App 中的组件
func WithReloaderFactories(factories ...func(*App) []hotreload.Reloader) ConfigCenterOption {
	return func(opts *configCenterOptions) {
		opts.reloaderFactories = append(opts.reloaderFactories, factories...)
	}
}

// WithConfigCenter 初始化配置中心和热加载系统
// cfg: 配置中心配置（从配置注册表中获取，如果为 nil 或 type 为 "none" 则跳过初始化）
// opts: 配置选项（字段设置器、重载器等）
// 注意：如果需要在重载器中访问 App 中的组件（如 RateLimitManager），
// 可以使用 WithReloaders 并提供获取组件的函数
func (b *AppBuilder) WithConfigCenter(cfg *configcenter.Config, opts ...ConfigCenterOption) *AppBuilder {
	// 解析选项
	options := &configCenterOptions{
		allowedPrefixes: make([]string, 0),
		reloaders:       make([]hotreload.Reloader, 0),
	}
	for _, opt := range opts {
		opt(options)
	}

	// 如果配置为空或类型为 none，跳过初始化
	if cfg == nil || cfg.Type == "none" {
		log.Debug("Config center not configured, skipping initialization")
		return b
	}

	// 创建配置中心客户端
	client, err := configcenter.NewFromConfig(cfg)
	if err != nil {
		b.handleInitError(DependencyConfigCenter, "ConfigCenter", err)
		return b
	}

	// 创建热加载管理器
	hotReloadManager := hotreload.NewManager()

	// 设置字段设置器（如果提供）
	if options.fieldSetter != nil {
		hotReloadManager.SetFieldSetter(options.fieldSetter, options.allowedPrefixes)
	}

	// 注册重载器（如果提供）
	for _, reloader := range options.reloaders {
		if err := hotReloadManager.RegisterReloader(reloader); err != nil {
			log.Warn("Failed to register reloader",
				zap.String("reloader", reloader.Patterns()[0]),
				zap.Error(err))
		}
	}

	// 注册延迟创建的重载器（在 App 构建后创建）
	// 这些重载器可以访问 App 中的组件
	for _, factory := range options.reloaderFactories {
		for _, reloader := range factory(b.app) {
			if err := hotReloadManager.RegisterReloader(reloader); err != nil {
				log.Warn("Failed to register reloader from factory",
					zap.String("reloader", reloader.Patterns()[0]),
					zap.Error(err))
			}
		}
	}

	// 连接配置中心和热加载管理器
	if err := client.ConnectHotReload(hotReloadManager); err != nil {
		b.handleInitError(DependencyConfigCenter, "ConfigCenter", err)
		return b
	}

	// 设置允许的配置前缀（用于过滤配置变更）
	if len(options.allowedPrefixes) > 0 {
		client.SetAllowedPrefixes(options.allowedPrefixes)
	}

	// 异步加载初始配置（不阻塞启动）
	// 配置会通过 watcher.Handle 统一处理，由 hotreload.Manager 管理
	go func() {
		if len(options.allowedPrefixes) > 0 {
			if err := client.LoadToTarget(options.allowedPrefixes); err != nil {
				log.Warn("Failed to load config from config center", zap.Error(err))
			} else {
				log.Info("Initial config loaded from config center",
					zap.String("type", string(client.Type)),
					zap.String("name", client.Name()))
			}
		}
	}()

	// 启动配置监听
	if err := client.WatchChanges(); err != nil {
		b.handleInitError(DependencyConfigCenter, "ConfigCenter", err)
		return b
	}

	// 注册关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		if err := client.Close(); err != nil {
			return err
		}
		log.Info("Config center closed",
			zap.String("type", string(client.Type)),
			zap.String("name", client.Name()))
		return nil
	})

	b.handleInitSuccess(DependencyConfigCenter, "ConfigCenter")
	log.Info("Config center initialized successfully",
		zap.String("type", string(client.Type)),
		zap.String("name", client.Name()),
		zap.Int("reloader_count", len(options.reloaders)),
		zap.Strings("allowed_prefixes", options.allowedPrefixes))

	return b
}

// WithDependency 注册并初始化自定义依赖（类型安全）
// depType: 依赖类型（用于依赖清单）
// init: 依赖初始化函数，返回依赖实例
// 使用示例:
//
//	builder.WithDependency("myService", app.DependencyType("custom"), func(b *app.AppBuilder) error {
//	    service := NewMyService()
//	    b.RegisterDependency(service)  // 类型安全注册
//	    return nil
//	})
func (b *AppBuilder) WithDependency(name string, depType DependencyType, init DependencyInitializer) *AppBuilder {
	// 执行初始化
	if err := init(b); err != nil {
		b.customDependencies = append(b.customDependencies, Dependency{
			Type:        depType,
			Name:        name,
			Initialized: false,
			Error:       fmt.Errorf("failed to initialize dependency %s: %w", name, err),
		})
		log.Error("Failed to initialize custom dependency", zap.String("name", name), zap.Error(err))
	} else {
		b.customDependencies = append(b.customDependencies, Dependency{
			Type:        depType,
			Name:        name,
			Initialized: true,
			Error:       nil,
		})
		log.Info("Custom dependency initialized", zap.String("name", name))
	}

	return b
}

// RegisterDependency 注册依赖实例
// 在 WithDependency 的初始化函数中调用此方法来注册依赖
// 使用示例: builder.RegisterDependency(myService)
// 注意：为了类型安全，推荐使用 app.Register(container, instance) 函数
func (b *AppBuilder) RegisterDependency(instance interface{}) {
	if instance == nil {
		log.Error("Cannot register nil dependency")
		return
	}
	container := b.app.GetDependencyContainer()
	container.register(instance)
}

// WithModules 注册业务模块
// 使用示例: builder.WithModules(apiModule.NewApiModule())
func (b *AppBuilder) WithModules(modules ...Module) *AppBuilder {
	b.app.moduleRegistry.RegisterAll(modules...)
	log.Info("Modules registered", zap.Int("count", len(modules)))
	return b
}

// collectInitializedServices 收集已成功初始化的服务列表
func (b *AppBuilder) collectInitializedServices() []string {
	var services []string

	// 检查显式依赖
	if b.app.MySQL != nil {
		services = append(services, "MySQL")
	}
	if b.app.PostgreSQL != nil {
		services = append(services, "PostgreSQL")
	}
	if b.app.Redis != nil {
		services = append(services, "Redis")
	}
	if b.app.MessageQueue != nil {
		services = append(services, "MessageQueue")
	}
	if b.app.ClickHouse != nil {
		services = append(services, "ClickHouse")
	}
	if b.app.Elasticsearch != nil {
		services = append(services, "Elasticsearch")
	}
	if b.app.Email != nil {
		services = append(services, "Email")
	}
	if b.app.MongoDB != nil {
		services = append(services, "MongoDB")
	}
	if b.app.OSS != nil {
		services = append(services, "OSS")
	}

	// 检查依赖容器中的依赖
	if b.app.dependencies.HasDependency(DependencyWebSocket) {
		services = append(services, "WebSocket")
	}
	if b.app.dependencies.HasDependency(DependencyLocalCache) {
		services = append(services, "LocalCache")
	}
	if b.app.dependencies.HasDependency(DependencyCron) {
		services = append(services, "Cron")
	}
	if b.app.dependencies.HasDependency(DependencyDistributedLock) {
		services = append(services, "DistributedLock")
	}
	if b.app.dependencies.HasDependency(DependencyRateLimit) {
		services = append(services, "RateLimit")
	}

	return services
}

// WithServices 初始化所有模块和服务
// 这会初始化所有已注册的模块，并将它们的服务注册到服务注册表
func (b *AppBuilder) WithServices() *AppBuilder {
	// 初始化所有模块，直接传递 app 对象
	services, err := b.app.moduleRegistry.InitializeAll(b.app)
	if err != nil {
		log.Error("Failed to initialize modules", zap.Error(err))
		return b
	}

	// 将所有模块的服务注册到服务注册表
	b.app.serviceRegistry.RegisterAll(services...)

	// 注册模块关闭函数
	b.app.lifecycle.RegisterShutdown(func() error {
		return b.app.moduleRegistry.ShutdownAll()
	})

	// 收集已成功初始化的服务列表
	initializedServices := b.collectInitializedServices()

	log.Info("Services initialized successfully",
		zap.Strings("initialized_services", initializedServices),
		zap.Int("module_count", len(b.app.moduleRegistry.GetModules())),
		zap.Int("service_count", len(b.app.serviceRegistry.GetServices())),
	)
	return b
}

// WithHTTP 初始化 HTTP 服务器相关组件（限流器、中间件、Gin 路由等）
// configDir: 配置目录路径（用于加载限流规则配置文件）
// routeSetupFunc: 路由设置回调函数（可选）
func (b *AppBuilder) WithHTTP(configDir string, routeSetupFunc RouteSetupFunc) *AppBuilder {
	if b.httpInit {
		return b
	}

	serverCfg, err := b.app.Config.GetServer()
	if err != nil {
		log.Warn("Server config not found, skipping HTTP initialization", zap.Error(err))
		return b
	}

	// 检查是否有 HTTP 配置
	if serverCfg.HTTP == nil {
		log.Warn("HTTP config not found, skipping HTTP initialization")
		return b
	}

	b.configDir = configDir
	b.routeSetupFunc = routeSetupFunc

	// 初始化限流器
	b.initHTTPRateLimiter(serverCfg)

	b.initHTTPMiddleware(serverCfg)

	if routeSetupFunc != nil {
		if b.app.Router == nil {
			return b
		}
		if err := routeSetupFunc(b.app.Router, b.app.Handler); err != nil {
			log.Error("Failed to setup routes", zap.Error(err))
			return b
		}
	}

	b.httpInit = true
	return b
}

// initHTTPRateLimiter 初始化 HTTP 限流器
func (b *AppBuilder) initHTTPRateLimiter(serverCfg *ServerConfig) {
	cfg := serverCfg.Features
	storage := cfg.RateLimit.Storage
	if storage == "" {
		storage = "local" // 默认使用本地限流器
	}

	log.Info("Rate limit configuration",
		zap.Bool("enabled", cfg.RateLimit.Enabled),
		zap.Float64("rate", cfg.RateLimit.Rate),
		zap.Int("burst", cfg.RateLimit.Burst),
		zap.String("storage", storage))

	if !cfg.RateLimit.Enabled {
		log.Info("Rate limit is disabled, skipping rate limit manager creation")
		return
	}

	// 创建业务限流管理器选项
	options := ratelimit.DefaultManagerOptions()

	// 设置默认限流器选项
	if cfg.RateLimit.Rate > 0 {
		options.DefaultLimiterOptions.Rate = cfg.RateLimit.Rate
	}
	if cfg.RateLimit.Burst > 0 {
		options.DefaultLimiterOptions.Burst = cfg.RateLimit.Burst
	}
	// 从追踪配置中获取是否启用追踪
	options.DefaultLimiterOptions.EnableTrace = cfg.Tracing.Enabled

	// 根据配置选择限流器工厂
	if storage == "redis" {
		// 如果 Redis 可用，使用 Redis 限流器
		if b.app.Redis != nil {
			options.LimiterFactory = func(opts *ratelimit.LimiterOptions) (ratelimit.Limiter, error) {
				// 创建 Redis 客户端适配器
				redisClientAdapter := &redisClientWrapper{client: b.app.Redis}
				// 创建 Redis 限流器
				return ratelimit.NewRedisLimiter(redisClientAdapter, opts)
			}
			log.Info("Using Redis distributed rate limiter")
		} else {
			// Redis 不可用，回退到本地限流器
			log.Warn("Redis not available, falling back to local rate limiter")
			options.LimiterFactory = ratelimit.DefaultLimiterFactory
		}
	} else {
		// 使用本地限流器（默认）
		options.LimiterFactory = ratelimit.DefaultLimiterFactory
		log.Info("Using local rate limiter")
	}

	// 创建业务限流管理器
	manager, err := ratelimit.NewManager(options)
	if err != nil {
		log.Error("Failed to create rate limit manager", zap.Error(err))
		return
	}
	if manager == nil {
		log.Error("Failed to create rate limit manager, rate limiting will not work")
		return
	}

	// 直接存储在 App 中
	b.app.RateLimitManager = manager
	log.Info("Rate limit manager created and stored in App",
		zap.String("service_name", b.app.ServiceName))

	// 1. 首先尝试从配置文件加载业务规则
	var loadedRules []*ratelimit.Rule
	if b.configDir != "" {
		rules, err := ratelimit.LoadRulesFromConfigDir(b.configDir)
		if err != nil {
			log.Warn("Failed to load rate limit rules from config file",
				zap.String("config_dir", b.configDir),
				zap.Error(err))
		} else if len(rules) > 0 {
			loadedRules = rules
			log.Info("Rate limit rules loaded from config file",
				zap.String("config_dir", b.configDir),
				zap.Int("rule_count", len(rules)))
		}
	}

	// 2. 如果配置文件中有规则，使用配置文件中的规则
	// 否则注册默认的全局限流规则
	if len(loadedRules) > 0 {
		// 注册从配置文件加载的规则
		for _, rule := range loadedRules {
			if err := manager.RegisterRule(rule); err != nil {
				log.Warn("Failed to register rate limit rule from config",
					zap.String("rule", rule.Name),
					zap.Error(err))
			} else {
				log.Info("Rate limit rule registered from config",
					zap.String("rule", rule.Name),
					zap.Int("priority", rule.Priority))
			}
		}
	} else {
		// 注册默认的全局限流规则
		defaultRule := &ratelimit.Rule{
			Name:        "global_default",
			Enabled:     true,
			Priority:    100,
			Dimensions:  []ratelimit.RateLimitDimension{ratelimit.DimensionGlobal},
			Rate:        cfg.RateLimit.Rate,
			Burst:       cfg.RateLimit.Burst,
			Algorithm:   ratelimit.AlgorithmTokenBucket,
			Fallback:    ratelimit.StrategyReject,
			Description: "全局默认限流规则",
		}

		if err := manager.RegisterRule(defaultRule); err != nil {
			log.Warn("Failed to register default rate limit rule", zap.Error(err))
		} else {
			log.Info("Default rate limit rule registered",
				zap.String("rule", defaultRule.Name),
				zap.Float64("rate", defaultRule.Rate),
				zap.Int("burst", defaultRule.Burst))
		}
	}

	log.Info("Rate limit manager created successfully",
		zap.String("name", b.app.ServiceName),
		zap.Int("total_rules", len(manager.GetRules())))
}

// initHTTPMiddleware 初始化 HTTP 中间件和路由
func (b *AppBuilder) initHTTPMiddleware(serverCfg *ServerConfig) {
	cfg := serverCfg.Features

	// 创建网关处理器
	gwHandler := gateway.NewGatewayHandler()
	b.app.Handler = gwHandler

	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)

	// 禁用 Gin 默认日志输出（我们使用自己的日志系统）
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	router := gin.New()
	// 只添加 Recovery 中间件，不添加 Logger 中间件
	router.Use(gin.Recovery())

	// 添加 CORS 中间件（最外层，根据配置）
	if cfg.CORS.Enabled {
		corsOrigins := cfg.CORS.AllowedOrigins.Strings()
		// 如果配置为空，表示允许所有
		if len(corsOrigins) == 0 {
			corsOrigins = nil // nil 表示允许所有
		}
		router.Use(gateway.CORSMiddleware(corsOrigins))
	}

	// 添加追踪中间件（始终启用，用于 traceID/requestID）
	router.Use(gateway.TraceMiddleware())

	// 添加请求超时中间件（根据配置）
	if cfg.Timeout.Enabled {
		timeout := cfg.Timeout.Timeout.Duration()
		router.Use(gateway.TimeoutMiddleware(timeout))
	}

	// 添加请求体大小限制中间件（根据配置）
	if cfg.SizeLimit.Enabled {
		maxSize := cfg.SizeLimit.MaxSize
		if maxSize <= 0 {
			maxSize = 10 * 1024 * 1024 // 默认 10MB
		}
		router.Use(gateway.SizeLimitMiddleware(maxSize))
	}

	// 添加 Metrics 中间件（根据配置）
	if cfg.Metrics.Enabled {
		router.Use(gateway.MetricsMiddleware())
	}

	// 添加业务限流中间件（根据配置）
	if cfg.RateLimit.Enabled {
		if b.app.RateLimitManager == nil {
			log.Error("Rate limit manager is nil, rate limiting middleware will not be registered")
		} else {
			log.Info("Business rate limit middleware registered successfully")
			router.Use(gateway.BusinessRateLimitMiddleware(b.app.RateLimitManager))
		}
	} else {
		log.Info("Rate limit is disabled, skipping rate limit middleware registration")
	}

	// 注册健康检查路由（必须在其他路由之前注册，以便快速响应）
	gateway.RegisterHealthCheckRoutes(router)

	// 注册 Prometheus metrics 端点（根据配置）
	if cfg.Metrics.Enabled {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	b.app.Router = router
}

// Build 构建 App 实例
func (b *AppBuilder) Build() (*App, error) {
	// 验证配置
	validator := NewConfigValidator()
	if err := validator.Validate(b.app.Config); err != nil {
		log.Error("Config validation failed", zap.Error(err))
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 根据配置设置指标收集开关
	metricsEnabled := true // 默认启用
	if serverCfg, err := b.app.Config.GetServer(); err == nil {
		metricsEnabled = serverCfg.Features.Metrics.Enabled
	}
	metrics.SetEnabled(metricsEnabled)
	if metricsEnabled {
		log.Info("Metrics collection enabled")
	} else {
		log.Info("Metrics collection disabled")
	}

	// 初始化 OpenTelemetry 追踪（根据配置）
	if serverCfg, err := b.app.Config.GetServer(); err == nil {
		if serverCfg.Features.Tracing.Enabled {
			traceCfg := trace.Config{
				ServiceName:    serverCfg.Features.Tracing.ServiceName,
				ServiceVersion: serverCfg.Features.Tracing.ServiceVersion,
				Environment:    serverCfg.Features.Tracing.Environment,
				OTLEndpoint:    serverCfg.Features.Tracing.OTLEndpoint,
				Enabled:        true,
				SampleRate:     serverCfg.Features.Tracing.SampleRate,
			}
			// 如果 service_name 为空，使用 server.service_name
			if traceCfg.ServiceName == "" {
				traceCfg.ServiceName = serverCfg.ServiceName
			}
			if err := trace.Init(traceCfg); err != nil {
				log.Warn("Failed to initialize tracing", zap.Error(err))
			} else {
				log.Info("Tracing initialized", zap.String("endpoint", traceCfg.OTLEndpoint))
				// 注册 trace 关闭函数到 lifecycle
				b.app.lifecycle.RegisterShutdown(func() error {
					if err := trace.Shutdown(context.Background()); err != nil {
						log.Error("Failed to shutdown tracing", zap.Error(err))
						return err
					}
					log.Info("Tracing shutdown completed")
					return nil
				})
			}
		}
	}

	// 添加自定义依赖到依赖清单
	b.app.dependencies = append(b.app.dependencies, b.customDependencies...)

	// 确保日志依赖已初始化（日志在配置加载时已初始化）
	b.app.dependencies = append(b.app.dependencies, Dependency{
		Type:        DependencyLog,
		Name:        "Log",
		Initialized: true,
		Error:       nil,
	})

	// 检查依赖初始化错误：如果依赖 enabled=true 但初始化失败，应该阻止启动
	var failedDependencies []string
	for _, dep := range b.app.dependencies {
		if dep.Error != nil && !dep.Initialized {
			// 检查该依赖是否被启用
			isEnabled := b.isDependencyEnabled(dep.Type)
			if isEnabled {
				failedDependencies = append(failedDependencies, fmt.Sprintf("%s: %v", dep.Name, dep.Error))
				log.Error("Required dependency initialization failed",
					zap.String("dependency", dep.Name),
					zap.String("type", string(dep.Type)),
					zap.Error(dep.Error),
				)
			} else {
				// enabled=false 的依赖失败是正常的，只记录警告
				log.Warn("Optional dependency initialization failed (disabled)",
					zap.String("dependency", dep.Name),
					zap.String("type", string(dep.Type)),
					zap.Error(dep.Error),
				)
			}
		}
	}

	// 如果有启用的依赖初始化失败，返回错误阻止启动
	if len(failedDependencies) > 0 {
		errMsg := fmt.Sprintf("failed to initialize required dependencies: %s", strings.Join(failedDependencies, "; "))
		log.Error("Application build failed due to dependency initialization errors",
			zap.Strings("failed_dependencies", failedDependencies),
			zap.String("error", errMsg),
		)
		return nil, errors.New(errMsg)
	}

	// 打印依赖清单
	log.Info("Application dependencies initialized", zap.String("dependencies", b.app.dependencies.String()))

	return b.app, nil
}

// isDependencyEnabled 检查依赖是否在配置中启用
func (b *AppBuilder) isDependencyEnabled(depType DependencyType) bool {
	switch depType {
	case DependencyMySQL:
		if cfg, err := b.app.Config.GetMySQL(); err == nil {
			return cfg.Enabled
		}
	case DependencyRedis:
		if cfg, err := b.app.Config.GetRedis(); err == nil {
			return cfg.Enabled
		}
	case DependencyClickHouse:
		if cfg, err := b.app.Config.GetClickHouse(); err == nil {
			return cfg.Enabled
		}
	case DependencyElasticsearch:
		if cfg, err := b.app.Config.GetElasticsearch(); err == nil {
			return cfg.Enabled
		}
	case DependencyEmail:
		if cfg, err := b.app.Config.GetEmail(); err == nil {
			return cfg.Enabled
		}
	case DependencyMongoDB:
		if cfg, err := b.app.Config.GetMongoDB(); err == nil {
			return cfg.Enabled
		}
	case DependencyPostgreSQL:
		if cfg, err := b.app.Config.GetPostgreSQL(); err == nil {
			return cfg.Enabled
		}
	case DependencyOSS:
		if cfg, err := b.app.Config.GetOSS(); err == nil {
			return cfg != nil && cfg.Default != "" && cfg.Default != "none"
		}
	case DependencyWebSocket:
		if cfg, err := b.app.Config.GetWebSocket(); err == nil {
			return cfg.Enabled
		}
	case DependencyXxlJob:
		if cfg, err := b.app.Config.GetXxlJob(); err == nil {
			return cfg.Enabled
		}
	}
	// 如果配置不存在或类型不匹配，默认返回 false（认为未启用）
	// 这样可以避免因为配置缺失而误判为启用
	return false
}
