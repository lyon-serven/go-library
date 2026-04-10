package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		showMenu()
		fmt.Print("\n请选择要运行的示例 (输入序号，输入 q 退出): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" || input == "Q" {
			fmt.Println("退出示例程序")
			break
		}

		option, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ 无效输入，请输入数字序号")
			continue
		}

		fmt.Println()
		switch option {
		case 1:
			logDemo1()
			logDemo2()
		case 2:
			main2()
		default:
			fmt.Println("❌ 无效选项，请重新选择")
		}

		fmt.Println("\n" + "========================================")
		fmt.Print("按回车键继续...")
		reader.ReadString('\n')
	}
}

func showMenu() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║           Log 示例程序选择菜单                            ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("║  1. 基础日志示例 (example.go)                              ║")
	fmt.Println("║     - 默认日志器使用                                      ║")
	fmt.Println("║     - 带字段的日志                                        ║")
	fmt.Println("║     - 自定义配置                                          ║")
	fmt.Println("║     - 多日志任务                                          ║")
	fmt.Println("║                                                           ║")
	fmt.Println("║  2. 环境服务示例 (env_service_example.go)                  ║")
	fmt.Println("║     - 开发环境 - 用户平台                                 ║")
	fmt.Println("║     - 生产环境 - 订单系统                                 ║")
	fmt.Println("║     - 测试环境 - 支付平台                                 ║")
	fmt.Println("║     - YAML 配置文件加载                                   ║")
	fmt.Println("║     - 默认日志器初始化                                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}
