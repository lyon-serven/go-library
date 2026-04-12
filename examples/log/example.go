package main

import (
	"fmt"
	"time"

	"github.com/lyon-serven/go-library/log"
)

func main_log() {
	logTask()
	logTask2()
}
func logDemo1() {
	// 测试默认日志器
	log.Info("测试信息日志")
	log.Errorf("测试错误日志: %s", "这是一个测试错误")

	// 测试带字段的日志
	//log.InfoK("用户操作", "user_id", 123)

	// 测试自定义配置
	config := log.DefaultConfig()
	config.LogToFile = false // 只输出到控制台
	config.Level = "debug"

	logger, err := log.NewLogger(config)
	if err != nil {
		log.Errorf("创建日志器失败: %v", err)
		return
	}

	logger.Debug("这是调试信息")
	logger.Info("自定义日志器测试成功")

	// 同步日志
	log.Sync()
}
func logDemo2() {
	// 简单使用
	log.Info("应用启动")
	log.Errorf("发生错误: %v", fmt.Errorf("111"))

	// 带字段
	//log.InfoK("用户登录", "user_id", 123)

	// 自定义配置
	config := log.DefaultConfig()
	config.Level = "debug"
	logger, _ := log.NewLogger(config)
	logger.Info("自定义配置")
}

func logTask() {
	config := &log.LogConfig{
		LogToFile:    true,
		Path:         "./logs/job1",           // 独立目录
		FileName:     "task.txt",              // 支持.txt格式
		FilePattern:  "/%Y_%m/%d/task_%H.log", // 去掉job1前缀
		RotationTime: time.Hour,               // 每小时轮转
		FileAge:      time.Hour * 24 * 7,      // 保留7天
	}

	logger, err := log.NewLogger(config)
	if err != nil {
		panic(err)
	}
	logger.Info("任务开始111111111")
}

func logTask2() {
	config := &log.LogConfig{
		LogToFile:    true,
		Path:         "./logs/job2",           // 独立目录
		FileName:     "task.txt",              // 支持.txt格式
		FilePattern:  "/%Y_%m/%d/task_%H.log", // 去掉job2前缀
		RotationTime: time.Hour,               // 每小时轮转
		FileAge:      time.Hour * 24 * 7,      // 保留7天
	}

	logger, err := log.NewLogger(config)
	if err != nil {
		panic(err)
	}
	logger.Info("任务开始22222222")
}
