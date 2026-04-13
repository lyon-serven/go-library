package main

import (
	"fmt"

	"github.com/lyon-serven/go-library/config"
)

func main() {
	ManageDemo()
}

func ManageDemo() {
	appConfig := &Config{}

	// 使用Option模式创建ConfigManager - 多种使用方式

	// 方式1: 启用加密功能
	manager := config.NewConfigManager(
		config.WithEnableDecryption(),
		// config.WithPollingWatcher(), // 强制使用轮询监听器
		//config.WithDisabledWatcher(), // 禁用文件监听
	)
	defer manager.Close()

	// 方法1: 相对路径（基于当前工作目录）
	relativePath := "./examples/config/configs/config.local.yaml"
	// 加载配置
	err := manager.LoadConfig(relativePath, appConfig)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}
	// 打印加载的配置信息
	fmt.Printf("配置加载成功！\n")
	fmt.Printf("应用名称: %s\n", appConfig.System.App)
	fmt.Printf("数据库DSN: %s\n", appConfig.Data.Database.Movie.DSNDes)
	fmt.Printf("Redis主机: %s\n", appConfig.Data.Redis.Host)
	fmt.Printf("日志级别: %s\n", appConfig.Log.Level)

	// 监听配置变化
	err = manager.WatchConfig(relativePath, appConfig, func() {
		fmt.Println("配置已更新")
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 保存配置
	err = manager.SaveConfig("./examples/config/configs/new_config.yaml", appConfig)
	if err != nil {
		fmt.Println("保存配置失败:", err)
	}
}
