package main

import (
	"fmt"
	"mylib/config"
)

// 简单的测试结构
type TestConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func main() {
	// 创建一个简单的YAML文件进行测试
	yamlContent := `app:
  name: "TestApp"
  version: "1.0.0"

database:
  host: "localhost"  
  port: 5432
  username: "test"

server:
  address: "0.0.0.0"
  port: 8080
`

	// 写入文件
	if err := config.SaveYAMLConfig("test.yaml", map[string]interface{}{
		"app": map[string]interface{}{
			"name":    "TestApp",
			"version": "1.0.0",
		},
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "test",
		},
		"server": map[string]interface{}{
			"address": "0.0.0.0",
			"port":    8080,
		},
	}); err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}

	// 测试配置分段
	var testConfig TestConfig
	fmt.Printf("测试配置类型: %T\n", &testConfig)

	section := config.NewConfigSection("database")
	if err := section.LoadSection("test.yaml", &testConfig); err != nil {
		fmt.Printf("配置分段失败: %v\n", err)
	} else {
		fmt.Printf("配置分段成功: %+v\n", testConfig)
	}
}
