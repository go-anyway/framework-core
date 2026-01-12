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

package app

import (
	"testing"
	"time"

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

type mockConfigRegistry struct {
	mysqlCfg      *db.MySQLConfig
	redisCfg      *db.RedisConfig
	serverCfg     *ServerConfig
	kafkaCfg      *messagequeue.KafkaConfig
	mqCfg         *messagequeue.RabbitMQConfig
	clickhouseCfg *clickhouse.Config
	emailCfg      *email.Config
	mongodbCfg    *mongodb.Config
	postgresqlCfg *db.PostgreSQLConfig
	ossCfg        *oss.Config
	websocketCfg  *websocket.Config
	esCfg         *elasticsearch.Config
	xxljobCfg     *xxljob.Config
}

func (m *mockConfigRegistry) GetMySQL() (*db.MySQLConfig, error) {
	return m.mysqlCfg, nil
}

func (m *mockConfigRegistry) GetRedis() (*db.RedisConfig, error) {
	return m.redisCfg, nil
}

func (m *mockConfigRegistry) GetClickHouse() (*clickhouse.Config, error) {
	return m.clickhouseCfg, nil
}

func (m *mockConfigRegistry) GetEmail() (*email.Config, error) {
	return m.emailCfg, nil
}

func (m *mockConfigRegistry) GetServer() (*ServerConfig, error) {
	return m.serverCfg, nil
}

func (m *mockConfigRegistry) GetLog() (*log.Config, error) {
	return &log.Config{}, nil
}

func (m *mockConfigRegistry) GetXxlJob() (*xxljob.Config, error) {
	return m.xxljobCfg, nil
}

func (m *mockConfigRegistry) GetElasticsearch() (*elasticsearch.Config, error) {
	return m.esCfg, nil
}

func (m *mockConfigRegistry) GetMongoDB() (*mongodb.Config, error) {
	return m.mongodbCfg, nil
}

func (m *mockConfigRegistry) GetPostgreSQL() (*db.PostgreSQLConfig, error) {
	return m.postgresqlCfg, nil
}

func (m *mockConfigRegistry) GetOSS() (*oss.Config, error) {
	return m.ossCfg, nil
}

func (m *mockConfigRegistry) GetWebSocket() (*websocket.Config, error) {
	return m.websocketCfg, nil
}

func (m *mockConfigRegistry) GetMessageQueue() (*messagequeue.Config, error) {
	return nil, nil
}

func (m *mockConfigRegistry) GetConfigCenter() (*configcenter.Config, error) {
	return nil, nil
}

func (m *mockConfigRegistry) GetServices() (map[string]ServiceConfig, error) {
	return make(map[string]ServiceConfig), nil
}

func TestMySQLConfig_DSN(t *testing.T) {
	cfg := &db.MySQLConfig{
		Host:      "localhost",
		Port:      3306,
		Database:  "testdb",
		Username:  "user",
		Password:  "pass",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}

	expected := "user:pass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=true&loc=Local"
	if dsn := cfg.DSN(); dsn != expected {
		t.Errorf("DSN() = %s, want %s", dsn, expected)
	}
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := &db.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	expected := "localhost:6379"
	if addr := cfg.Addr(); addr != expected {
		t.Errorf("Addr() = %s, want %s", addr, expected)
	}
}

func TestRedisConfig_DialTimeoutDuration(t *testing.T) {
	cfg := &db.RedisConfig{
		DialTimeout: pkgConfig.Duration(5 * time.Second),
	}

	if d := cfg.DialTimeoutDuration(); d != 5*time.Second {
		t.Errorf("DialTimeoutDuration() = %v, want %v", d, 5*time.Second)
	}
}

func TestKafkaConfig_DialTimeoutDuration(t *testing.T) {
	// KafkaConfig 已迁移到 messagequeue 包，此测试已废弃
	t.Skip("KafkaConfig has been moved to messagequeue package")
}

func TestMQConfig_DialTimeoutDuration(t *testing.T) {
	// MQConfig 已迁移到 messagequeue 包，此测试已废弃
	t.Skip("MQConfig has been moved to messagequeue package")
}

func TestClickHouseConfig_DialTimeoutDuration(t *testing.T) {
	cfg := &clickhouse.Config{
		DialTimeout: pkgConfig.Duration(30 * time.Second),
	}

	if d := cfg.DialTimeoutDuration(); d != 30*time.Second {
		t.Errorf("DialTimeoutDuration() = %v, want %v", d, 30*time.Second)
	}
}

func TestEmailConfig_TimeoutDuration(t *testing.T) {
	cfg := &email.Config{
		Timeout: pkgConfig.Duration(30 * time.Second),
	}

	if d := cfg.TimeoutDuration(); d != 30*time.Second {
		t.Errorf("TimeoutDuration() = %v, want %v", d, 30*time.Second)
	}
}

func TestGRPCConfig_Address(t *testing.T) {
	cfg := &GRPCConfig{
		Port: "50051",
	}

	expected := ":50051"
	if addr := cfg.Address(); addr != expected {
		t.Errorf("Address() = %s, want %s", addr, expected)
	}
}

func TestGRPCConfig_FullAddress(t *testing.T) {
	cfg := &GRPCConfig{
		Host: "0.0.0.0",
		Port: "50051",
	}

	expected := "0.0.0.0:50051"
	if addr := cfg.FullAddress(); addr != expected {
		t.Errorf("FullAddress() = %s, want %s", addr, expected)
	}
}

func TestHTTPConfig_Address(t *testing.T) {
	cfg := &HTTPConfig{
		Port: "8080",
	}

	expected := ":8080"
	if addr := cfg.Address(); addr != expected {
		t.Errorf("Address() = %s, want %s", addr, expected)
	}
}

func TestLogConfig_ToLogOptions(t *testing.T) {
	cfg := &log.Config{
		Level:             "debug",
		Format:            "json",
		OutputPaths:       []string{"/var/log/app.log"},
		ErrorOutputPaths:  []string{"/var/log/app-error.log"},
		DisableCaller:     true,
		DisableStacktrace: true,
		Filename:          "/var/log/app.log",
		MaxSize:           100,
		MaxAge:            7,
		MaxBackups:        3,
		Compress:          true,
		Development:       false,
	}

	opts := cfg.ToOptions()

	if opts.Level != "debug" {
		t.Errorf("Level = %s, want debug", opts.Level)
	}
	if opts.Format != "json" {
		t.Errorf("Format = %s, want json", opts.Format)
	}
	if len(opts.OutputPaths) != 1 || opts.OutputPaths[0] != "/var/log/app.log" {
		t.Errorf("OutputPaths = %v, want [/var/log/app.log]", opts.OutputPaths)
	}
	if opts.MaxSize != 100 {
		t.Errorf("MaxSize = %d, want 100", opts.MaxSize)
	}
}

func TestConfigValidator_Validate(t *testing.T) {
	validator := NewConfigValidator()

	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "test-service",
		},
		mysqlCfg: &db.MySQLConfig{
			Host:           "localhost",
			Port:           3306,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		redisCfg: &db.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			PoolSize:     20,
			MinIdleConns: 5,
			DB:           0,
		},
		kafkaCfg: &messagequeue.KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
			Enabled:  true,
		},
		mqCfg: &messagequeue.RabbitMQConfig{
			URL:      "amqp://localhost:5672",
			Username: "guest",
			Password: "guest",
		},
		clickhouseCfg: &clickhouse.Config{
			Host:         "localhost",
			Port:         9000,
			Database:     "testdb",
			Username:     "user",
			Password:     "pass",
			MaxOpenConns: 100,
		},
		emailCfg: &email.Config{
			Host:     "localhost",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			From:     "test@example.com",
		},
		mongodbCfg: &mongodb.Config{
			Host:        "localhost",
			Port:        27017,
			Database:    "testdb",
			MaxPoolSize: 100,
		},
		postgresqlCfg: &db.PostgreSQLConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		ossCfg: &oss.Config{
			Default: "aws",
			Stores: map[string]oss.StoreConfig{
				"aws": {
					Type: "aws",
					AWS: &oss.AWSConfig{
						Region:          "us-east-1",
						AccessKeyID:     "test-key",
						SecretAccessKey: "test-secret",
						BucketName:      "test-bucket",
					},
				},
			},
		},
		esCfg: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		},
		xxljobCfg: &xxljob.Config{
			Enabled:     false,
			ServerAddr:  "http://localhost:8080/xxl-job-admin",
			RegistryKey: "test-key",
		},
	}

	err := validator.Validate(cfg)
	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestConfigValidator_ValidateServerConfig(t *testing.T) {
	validator := NewConfigValidator()

	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "",
			GRPC: GRPCConfig{
				Port: "invalid",
			},
		},
		mysqlCfg: &db.MySQLConfig{
			Host:           "localhost",
			Port:           3306,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		redisCfg: &db.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			PoolSize:     20,
			MinIdleConns: 5,
			DB:           0,
		},
		kafkaCfg: &messagequeue.KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
			Enabled:  true,
		},
		mqCfg: &messagequeue.RabbitMQConfig{
			URL:      "amqp://localhost:5672",
			Username: "guest",
			Password: "guest",
		},
		clickhouseCfg: &clickhouse.Config{
			Host:         "localhost",
			Port:         9000,
			Database:     "testdb",
			Username:     "user",
			Password:     "pass",
			MaxOpenConns: 100,
		},
		emailCfg: &email.Config{
			Host:     "localhost",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			From:     "test@example.com",
		},
		mongodbCfg: &mongodb.Config{
			Host:        "localhost",
			Port:        27017,
			Database:    "testdb",
			MaxPoolSize: 100,
		},
		postgresqlCfg: &db.PostgreSQLConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		ossCfg: &oss.Config{
			Default: "aws",
			Stores: map[string]oss.StoreConfig{
				"aws": {
					Type: "aws",
					AWS: &oss.AWSConfig{
						Region:          "us-east-1",
						AccessKeyID:     "test-key",
						SecretAccessKey: "test-secret",
						BucketName:      "test-bucket",
					},
				},
			},
		},
		esCfg: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		},
		xxljobCfg: &xxljob.Config{
			Enabled:     false,
			ServerAddr:  "http://localhost:8080/xxl-job-admin",
			RegistryKey: "test-key",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() should return error for invalid server config")
	}
}

func TestConfigValidator_ValidateMySQLConfig(t *testing.T) {
	validator := NewConfigValidator()

	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "test",
		},
		mysqlCfg: &db.MySQLConfig{
			Host:           "localhost",
			Port:           99999,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 0,
		},
		redisCfg: &db.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			PoolSize:     20,
			MinIdleConns: 5,
			DB:           0,
		},
		kafkaCfg: &messagequeue.KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
			Enabled:  true,
		},
		mqCfg: &messagequeue.RabbitMQConfig{
			URL:      "amqp://localhost:5672",
			Username: "guest",
			Password: "guest",
		},
		clickhouseCfg: &clickhouse.Config{
			Host:         "localhost",
			Port:         9000,
			Database:     "testdb",
			Username:     "user",
			Password:     "pass",
			MaxOpenConns: 100,
		},
		emailCfg: &email.Config{
			Host:     "localhost",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			From:     "test@example.com",
		},
		mongodbCfg: &mongodb.Config{
			Host:        "localhost",
			Port:        27017,
			Database:    "testdb",
			MaxPoolSize: 100,
		},
		postgresqlCfg: &db.PostgreSQLConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		ossCfg: &oss.Config{
			Default: "aws",
			Stores: map[string]oss.StoreConfig{
				"aws": {
					Type: "aws",
					AWS: &oss.AWSConfig{
						Region:          "us-east-1",
						AccessKeyID:     "test-key",
						SecretAccessKey: "test-secret",
						BucketName:      "test-bucket",
					},
				},
			},
		},
		esCfg: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		},
		xxljobCfg: &xxljob.Config{
			Enabled:     false,
			ServerAddr:  "http://localhost:8080/xxl-job-admin",
			RegistryKey: "test-key",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() should return error for invalid mysql config")
	}
}

func TestConfigValidator_ValidateRedisConfig(t *testing.T) {
	validator := NewConfigValidator()

	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "test",
		},
		mysqlCfg: &db.MySQLConfig{
			Host:           "localhost",
			Port:           3306,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		redisCfg: &db.RedisConfig{
			Host:         "localhost",
			Port:         99999,
			PoolSize:     0,
			MinIdleConns: -1,
			DB:           16,
		},
		kafkaCfg: &messagequeue.KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
			Enabled:  true,
		},
		mqCfg: &messagequeue.RabbitMQConfig{
			URL:      "amqp://localhost:5672",
			Username: "guest",
			Password: "guest",
		},
		clickhouseCfg: &clickhouse.Config{
			Host:         "localhost",
			Port:         9000,
			Database:     "testdb",
			Username:     "user",
			Password:     "pass",
			MaxOpenConns: 100,
		},
		emailCfg: &email.Config{
			Host:     "localhost",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			From:     "test@example.com",
		},
		mongodbCfg: &mongodb.Config{
			Host:        "localhost",
			Port:        27017,
			Database:    "testdb",
			MaxPoolSize: 100,
		},
		postgresqlCfg: &db.PostgreSQLConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		ossCfg: &oss.Config{
			Default: "aws",
			Stores: map[string]oss.StoreConfig{
				"aws": {
					Type: "aws",
					AWS: &oss.AWSConfig{
						Region:          "us-east-1",
						AccessKeyID:     "test-key",
						SecretAccessKey: "test-secret",
						BucketName:      "test-bucket",
					},
				},
			},
		},
		esCfg: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		},
		xxljobCfg: &xxljob.Config{
			Enabled:     false,
			ServerAddr:  "http://localhost:8080/xxl-job-admin",
			RegistryKey: "test-key",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() should return error for invalid redis config")
	}
}

func TestConfigValidator_ValidateKafkaConfig(t *testing.T) {
	validator := NewConfigValidator()

	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "test",
		},
		mysqlCfg: &db.MySQLConfig{
			Host:           "localhost",
			Port:           3306,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		redisCfg: &db.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			PoolSize:     20,
			MinIdleConns: 5,
			DB:           0,
		},
		kafkaCfg: &messagequeue.KafkaConfig{
			Brokers:      []string{},
			ClientID:     "",
			EnableSASL:   true,
			SASLUsername: "",
			SASLPassword: "",
			Enabled:      true,
		},
		mqCfg: &messagequeue.RabbitMQConfig{
			URL:      "amqp://localhost:5672",
			Username: "guest",
			Password: "guest",
		},
		clickhouseCfg: &clickhouse.Config{
			Host:         "localhost",
			Port:         9000,
			Database:     "testdb",
			Username:     "user",
			Password:     "pass",
			MaxOpenConns: 100,
		},
		emailCfg: &email.Config{
			Host:     "localhost",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			From:     "test@example.com",
		},
		mongodbCfg: &mongodb.Config{
			Host:        "localhost",
			Port:        27017,
			Database:    "testdb",
			MaxPoolSize: 100,
		},
		postgresqlCfg: &db.PostgreSQLConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		ossCfg: &oss.Config{
			Default: "aws",
			Stores: map[string]oss.StoreConfig{
				"aws": {
					Type: "aws",
					AWS: &oss.AWSConfig{
						Region:          "us-east-1",
						AccessKeyID:     "test-key",
						SecretAccessKey: "test-secret",
						BucketName:      "test-bucket",
					},
				},
			},
		},
		esCfg: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		},
		xxljobCfg: &xxljob.Config{
			Enabled:     false,
			ServerAddr:  "http://localhost:8080/xxl-job-admin",
			RegistryKey: "test-key",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() should return error for invalid kafka config")
	}
}

func TestValidateConfig_ConvenienceFunction(t *testing.T) {
	cfg := &mockConfigRegistry{
		serverCfg: &ServerConfig{
			ServiceName: "test-service",
		},
		mysqlCfg: &db.MySQLConfig{
			Host:           "localhost",
			Port:           3306,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		redisCfg: &db.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			PoolSize:     20,
			MinIdleConns: 5,
			DB:           0,
		},
		kafkaCfg: &messagequeue.KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
			Enabled:  true,
		},
		mqCfg: &messagequeue.RabbitMQConfig{
			URL:      "amqp://localhost:5672",
			Username: "guest",
			Password: "guest",
		},
		clickhouseCfg: &clickhouse.Config{
			Host:         "localhost",
			Port:         9000,
			Database:     "testdb",
			Username:     "user",
			Password:     "pass",
			MaxOpenConns: 100,
		},
		emailCfg: &email.Config{
			Host:     "localhost",
			Port:     587,
			Username: "test@example.com",
			Password: "password",
			From:     "test@example.com",
		},
		mongodbCfg: &mongodb.Config{
			Host:        "localhost",
			Port:        27017,
			Database:    "testdb",
			MaxPoolSize: 100,
		},
		postgresqlCfg: &db.PostgreSQLConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "testdb",
			Username:       "user",
			Password:       "pass",
			MaxConnections: 100,
		},
		ossCfg: &oss.Config{
			Default: "aws",
			Stores: map[string]oss.StoreConfig{
				"aws": {
					Type: "aws",
					AWS: &oss.AWSConfig{
						Region:          "us-east-1",
						AccessKeyID:     "test-key",
						SecretAccessKey: "test-secret",
						BucketName:      "test-bucket",
					},
				},
			},
		},
		esCfg: &elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		},
		xxljobCfg: &xxljob.Config{
			Enabled:     false,
			ServerAddr:  "http://localhost:8080/xxl-job-admin",
			RegistryKey: "test-key",
		},
	}

	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v", err)
	}
}

func TestDependencyList_HasDependency(t *testing.T) {
	list := DependencyList{
		{Type: DependencyMySQL, Name: "MySQL", Initialized: true},
		{Type: DependencyRedis, Name: "Redis", Initialized: false},
	}

	if !list.HasDependency(DependencyMySQL) {
		t.Error("HasDependency(DependencyMySQL) should return true")
	}
	if list.HasDependency(DependencyRedis) {
		t.Error("HasDependency(DependencyRedis) should return false (not initialized)")
	}
	if list.HasDependency(DependencyKafka) {
		t.Error("HasDependency(DependencyKafka) should return false (not present)")
	}
}

func TestDependencyList_GetDependency(t *testing.T) {
	list := DependencyList{
		{Type: DependencyMySQL, Name: "MySQL", Initialized: true},
		{Type: DependencyRedis, Name: "Redis", Initialized: false},
	}

	dep := list.GetDependency(DependencyMySQL)
	if dep == nil || dep.Name != "MySQL" {
		t.Errorf("GetDependency(DependencyMySQL) = %v, want MySQL", dep)
	}

	dep = list.GetDependency(DependencyKafka)
	if dep != nil {
		t.Error("GetDependency(DependencyKafka) should return nil")
	}
}

func TestDependencyList_String(t *testing.T) {
	list := DependencyList{
		{Type: DependencyMySQL, Name: "MySQL", Initialized: true},
		{Type: DependencyRedis, Name: "Redis", Initialized: false},
	}

	str := list.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}

func TestDependencyContainer_RegisterAndGet(t *testing.T) {
	container := NewDependencyContainer()

	service := &testService{Name: "test"}

	Register(container, service)

	retrieved, ok := Get[*testService](container)
	if !ok {
		t.Fatal("Get() returned false")
	}
	if retrieved.Name != service.Name {
		t.Errorf("Get() = %v, want %v", retrieved, service)
	}
}

func TestDependencyContainer_Has(t *testing.T) {
	container := NewDependencyContainer()

	service := &testService{}

	if Has[*testService](container) {
		t.Error("Has() should return false before registration")
	}

	Register(container, service)

	if !Has[*testService](container) {
		t.Error("Has() should return true after registration")
	}
}

func TestDependencyContainer_MustGet(t *testing.T) {
	container := NewDependencyContainer()

	service := &testService{Value: 42}
	Register(container, service)

	retrieved := MustGet[*testService](container)
	if retrieved.Value != 42 {
		t.Errorf("MustGet() = %v, want 42", retrieved.Value)
	}
}

func TestDependencyContainer_MustGetPanic(t *testing.T) {
	container := NewDependencyContainer()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet() should panic for missing dependency")
		}
	}()

	type TestService struct{}
	_ = MustGet[TestService](container)
}

func TestDependencyContainer_GetAll(t *testing.T) {
	container := NewDependencyContainer()

	type ServiceA struct{}
	type ServiceB struct{}

	Register(container, &ServiceA{})
	Register(container, &ServiceB{})

	types := container.GetAll()
	if len(types) != 2 {
		t.Errorf("GetAll() returned %d types, want 2", len(types))
	}
}

func TestLifecycle_RegisterAndShutdown(t *testing.T) {
	lifecycle := NewLifecycle()
	executed := false

	lifecycle.RegisterShutdown(func() error {
		executed = true
		return nil
	})

	if lifecycle.IsStarted() {
		t.Error("IsStarted() should return false before starting")
	}

	lifecycle.SetStarted()

	if !lifecycle.IsStarted() {
		t.Error("IsStarted() should return true after starting")
	}

	err := lifecycle.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
	if !executed {
		t.Error("shutdown function should have been executed")
	}
}

func TestLifecycle_ShutdownOrder(t *testing.T) {
	lifecycle := NewLifecycle()
	order := make([]int, 0)

	lifecycle.RegisterShutdown(func() error {
		order = append(order, 1)
		return nil
	})
	lifecycle.RegisterShutdown(func() error {
		order = append(order, 2)
		return nil
	})
	lifecycle.RegisterShutdown(func() error {
		order = append(order, 3)
		return nil
	})

	lifecycle.Shutdown()

	if len(order) != 3 {
		t.Fatalf("order length = %d, want 3", len(order))
	}
	if order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Errorf("order = %v, want [3, 2, 1] (reverse order)", order)
	}
}

func TestLifecycle_ShutdownErrors(t *testing.T) {
	lifecycle := NewLifecycle()

	lifecycle.RegisterShutdown(func() error {
		return nil
	})
	lifecycle.RegisterShutdown(func() error {
		return &testError{"error1"}
	})
	lifecycle.RegisterShutdown(func() error {
		return &testError{"error2"}
	})

	err := lifecycle.Shutdown()
	if err == nil {
		t.Error("Shutdown() should return error")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

type mockModule struct{}

func (m *mockModule) Name() string {
	return "mock-module"
}

func (m *mockModule) Initialize(app *App) ([]ServiceRegistration, error) {
	return nil, nil
}

func (m *mockModule) Shutdown() error {
	return nil
}

type testService struct {
	Name  string
	Value int
}

func TestModuleRegistry_RegisterAndGet(t *testing.T) {
	registry := NewModuleRegistry()

	mod := &mockModule{}

	registry.Register(mod)

	modules := registry.GetModules()
	if len(modules) != 1 {
		t.Errorf("GetModules() returned %d modules, want 1", len(modules))
	}
}

func TestModuleRegistry_RegisterAll(t *testing.T) {
	registry := NewModuleRegistry()

	mod1 := &mockModule{}
	mod2 := &mockModule{}

	registry.RegisterAll(mod1, mod2)

	modules := registry.GetModules()
	if len(modules) != 2 {
		t.Errorf("GetModules() returned %d modules, want 2", len(modules))
	}
}

func TestServiceRegistry_RegisterAndGet(t *testing.T) {
	registry := NewServiceRegistry()

	reg := ServiceRegistration{
		Name: "test-service",
	}

	registry.Register(reg)

	services := registry.GetServices()
	if len(services) != 1 {
		t.Errorf("GetServices() returned %d services, want 1", len(services))
	}
	if services[0].Name != "test-service" {
		t.Errorf("GetServices()[0].Name = %s, want test-service", services[0].Name)
	}
}

func TestServiceRegistry_RegisterAll(t *testing.T) {
	registry := NewServiceRegistry()

	regs := []ServiceRegistration{
		{Name: "service1"},
		{Name: "service2"},
	}

	registry.RegisterAll(regs...)

	services := registry.GetServices()
	if len(services) != 2 {
		t.Errorf("GetServices() returned %d services, want 2", len(services))
	}
}
