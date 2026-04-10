package main

import (
	"time"

	"gitee.com/wangsoft/go-library/log"
	"go.uber.org/zap"
)

func main2() {
	// ========================================
	// 示例 1: 开发环境 - 用户平台
	// ========================================
	println("========================================")
	println("示例 1: 开发环境 - 用户平台")
	println("========================================")

	config1 := &log.LogConfig{
		LogToStdout:  true,
		LogToFile:    false,
		Level:        "info",
		Format:       "json",
		Environment:  "dev",
		SystemName:   "user-platform",
		EnableCaller: true,
	}

	logger1, err := log.NewLogger(config1)
	if err != nil {
		panic(err)
	}
	defer logger1.Sync()

	// 所有日志都会自动包含 env 和 service 字段
	logger1.Info("用户登录",
		zap.String("user_id", "12345"),
		zap.String("ip", "192.168.1.100"),
	)

	logger1.Warn("登录失败次数过多",
		zap.String("user_id", "12345"),
		zap.Int("fail_count", 5),
	)

	// ========================================
	// 示例 2: 生产环境 - 订单系统
	// ========================================
	println("\n========================================")
	println("示例 2: 生产环境 - 订单系统")
	println("========================================")

	config2 := &log.LogConfig{
		LogToStdout:  true,
		LogToFile:    false,
		Level:        "info",
		Format:       "json",
		Environment:  "prod",
		SystemName:   "order-system",
		EnableCaller: true,
	}

	logger2, err := log.NewLogger(config2)
	if err != nil {
		panic(err)
	}
	defer logger2.Sync()

	logger2.Info("创建订单",
		zap.String("order_id", "ORD-2024-001"),
		zap.String("user_id", "12345"),
		zap.Float64("amount", 299.99),
	)

	logger2.Error("支付失败",
		zap.String("order_id", "ORD-2024-001"),
		zap.String("error", "insufficient balance"),
	)

	// ========================================
	// 示例 3: 测试环境 - 支付平台
	// ========================================
	println("\n========================================")
	println("示例 3: 测试环境 - 支付平台")
	println("========================================")

	config3 := &log.LogConfig{
		LogToStdout:  true,
		LogToFile:    false,
		Level:        "debug",
		Format:       "json",
		Environment:  "test",
		SystemName:   "payment-platform",
		EnableCaller: true,
	}

	logger3, err := log.NewLogger(config3)
	if err != nil {
		panic(err)
	}
	defer logger3.Sync()

	logger3.Debug("调试支付流程",
		zap.String("payment_id", "PAY-001"),
		zap.String("method", "alipay"),
	)

	logger3.Info("支付成功",
		zap.String("payment_id", "PAY-001"),
		zap.Float64("amount", 299.99),
	)

	// ========================================
	// 示例 4: 使用 YAML 配置文件
	// ========================================
	println("\n========================================")
	println("示例 4: 从配置文件加载")
	println("========================================")

	// 实际使用中，可以从 YAML 文件加载配置
	// 例如：
	// config:
	//   log:
	//     environment: prod
	//     system_name: api-gateway
	//     level: info
	//     log_to_stdout: true
	//     log_to_file: true
	//     path: ./logs
	//     file_name: api-gateway.log

	config4 := &log.LogConfig{
		LogToStdout:  true,
		LogToFile:    true,
		Level:        "info",
		Format:       "json",
		Environment:  "prod",
		SystemName:   "api-gateway",
		Path:         "./logs",
		FileName:     "api-gateway.log",
		RotationTime: time.Hour * 24,
		EnableCaller: true,
	}

	logger4, err := log.NewLogger(config4)
	if err != nil {
		panic(err)
	}
	defer logger4.Sync()

	logger4.Info("API 请求",
		zap.String("method", "POST"),
		zap.String("path", "/api/v1/users"),
		zap.Int("status", 200),
		zap.Duration("latency", 45*time.Millisecond),
	)

	// ========================================
	// 示例 5: 使用默认日志器
	// ========================================
	println("\n========================================")
	println("示例 5: 使用默认日志器")
	println("========================================")

	// 初始化默认日志器
	defaultConfig := &log.LogConfig{
		LogToStdout:  true,
		LogToFile:    false,
		Level:        "info",
		Format:       "json",
		Environment:  "dev",
		SystemName:   "my-app",
		EnableCaller: true,
	}

	err = log.InitDefault(defaultConfig)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// 使用全局便捷方法
	log.Info("应用启动",
		zap.String("version", "1.0.0"),
		zap.Int("port", 8080),
	)

	log.Infof("服务监听在端口 %d", 8080)

	// ========================================
	// 总结
	// ========================================
	println("\n========================================")
	println("总结")
	println("========================================")
	println()
	println("✓ 每条日志都自动包含 env 和 system 字段")
	println("✓ 便于在日志聚合系统中过滤和查询")
	println("✓ 快速定位问题所属的环境和系统")
	println()
	println("日志格式示例:")
	println(`{
  "level": "info",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "caller": "main.go:25",
  "message": "用户登录",
  "env": "prod",
  "system": "user-platform",
  "user_id": "12345",
  "ip": "192.168.1.100"
}`)
	println()
	println("在 ELK 中查询示例:")
	println(`  env:"prod" AND system:"user-platform" AND level:"error"`)
	println(`  env:"dev" AND message:"登录"`)
	println("========================================")
}
