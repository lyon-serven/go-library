package main

import (
	"fmt"

	"gitee.com/wangsoft/go-library/config"
)

func main() {
	ManageDemo()
}

func ManageDemo() {
	appConfig := &Config{}
	manager := config.NewConfigManager()
	defer manager.Close()

	// 加载配置
	err := manager.LoadConfig("./configs/config.local.yaml", &appConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 监听配置变化
	err = manager.WatchConfig("./configs/config.local.yaml", &appConfig, func() {
		fmt.Println("配置已更新")
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 保存配置
	err = manager.SaveConfig("./configs/new_config.yaml", &appConfig)
	if err != nil {
		fmt.Println("保存配置失败:", err)
	}
}
