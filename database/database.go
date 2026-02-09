package database

import (
	"autoSSL/config"
	"autoSSL/ent"
	"autoSSL/logger"
	"errors"
	"fmt"
	"log"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	err    error
	client *sql.Driver
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.AppConfig) (*ent.Client, error) {
	log.Printf("Initializing database connection...")

	dbType := cfg.Database.Driver
	if dbType == "" {
		dbType = "sqlite"
	}

	switch dbType {
	case "postgres":
		log.Printf("连接到Postgres数据库 %q", cfg.Database.Host)
		client, err = sql.Open(string(dbType), fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port))
		if err != nil {
			return nil, fmt.Errorf("连接Postgres失败: %v", err)
		}

	case "mysql":
		log.Printf("连接到MySQL数据库 %q", cfg.Database.Host)
		var host string
		if cfg.Database.UnixSocket {
			host = fmt.Sprintf("unix(%s)", cfg.Database.Host)
		} else {
			host = fmt.Sprintf("(%s:%d)", cfg.Database.Host, cfg.Database.Port)
		}
		logger.Debug("MySQL连接主机", zap.String("host", host))
		client, err = sql.Open(string(dbType), fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.User,
			cfg.Database.Password,
			host,
			cfg.Database.Name))
		if err != nil {
			return nil, fmt.Errorf("连接MySQL失败: %v", err)
		}

	// case "sqlite":
	// 	dsn = cfg.Database.DSN

	default:
		return nil, fmt.Errorf("不支持的数据库类型: %q", dbType)
	}

	// 设置连接池
	db := client.DB()
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Second * 30)
	driverOpt := ent.Driver(client)

	// 测试数据库连接
	if err = db.Ping(); err != nil {
		return nil, errors.New("数据库连接测试失败: " + err.Error())
	}

	// 初始化数据库客户端
	// 当日志级别为debug时，开启ent的调试模式打印数据库查询日志
	if cfg.Logger.Level == "debug" {
		driverOpt = ent.Driver(dialect.Debug(client, func(v ...interface{}) {
			logger.Debug("数据库查询", zap.Any("query", v))
		}))
	}

	return ent.NewClient(driverOpt), nil
}
