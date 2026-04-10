package main

import (
	"context"
	"fmt"
	"time"

	"gitee.com/wangsoft/go-library/message"
)

func main() {
	DemoServerChan()
}

func DemoServerChan() {
	fmt.Println("=== Server酱 消息通知示例 ===")
	fmt.Println("注册地址：https://sct.ftqq.com，微信扫码登录获取 SendKey")

	// 替换为你自己的 SendKey（格式：SCT_xxxxxx）
	sendKey := "ST23424234"
	n := message.NewServerChanNotifier(sendKey)
	ctx := context.Background()

	// 1. 发送纯文本
	fmt.Println("\n1. 发送纯文本消息...")
	err := n.Send(ctx, message.TextMsg("懂王又发新消息了：特朗普在伊朗和委内瑞拉相继干预后，彻底瓦解了中国通过制裁获取石油进口的暴利模式：justthenews.com/government/dip"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	//// 2. 发送 Markdown
	//fmt.Println("\n2. 发送 Markdown 消息...")
	//content := "## 告警详情\n\n" +
	//	"| 项目 | 内容 |\n" +
	//	"|--|--|\n" +
	//	"| 服务 | Redis |\n" +
	//	"| 状态 | 🔴 连接断开 |\n" +
	//	"| 时间 | " + time.Now().Format("2006-01-02 15:04:05") + " |\n\n" +
	//	"> 请立即检查 Redis 服务状态"
	//err = n.Send(ctx, message.MarkdownMsg("🔴 Redis 告警", content))
	//if err != nil {
	//	fmt.Println("   ❌ 失败:", err)
	//} else {
	//	fmt.Println("   ✅ 发送成功")
	//}
	//
	//// 3. 模拟心跳告警回调
	//fmt.Println("\n3. 模拟心跳告警回调...")
	//onAlert := func(level, msg string) {
	//	alertMsg := message.MarkdownMsg(
	//		"⚠️ 系统告警",
	//		fmt.Sprintf("**级别**：%s\n\n**内容**：%s\n\n**时间**：%s",
	//			level, msg, time.Now().Format("15:04:05")),
	//	)
	//	if err := n.Send(ctx, alertMsg); err != nil {
	//		fmt.Println("   ❌ 告警发送失败:", err)
	//	} else {
	//		fmt.Println("   ✅ 告警发送成功:", msg)
	//	}
	//}
	//
	//onAlert("ERROR", "Redis 连接断开")
	//onAlert("RECOVER", "Redis 连接已恢复")
	//onAlert("WARN", "Redis 延迟过高：350ms")
}
func DemoWxPusher() {
	ctx := context.Background()
	// 创建微信推送通知器
	wxNotifier := message.NewWxPusherNotifier("*****", "*****")
	// 发文本
	err := wxNotifier.Send(ctx, message.TextMsg("测试-Redis 连接断开！"))
	if err != nil {
		fmt.Println("报错" + err.Error())
		return
	}

	// 发 Markdown（微信里渲染）
	err = wxNotifier.Send(ctx, message.MarkdownMsg("🔴 Redis 告警", "**连接断开**\n\n时间："+time.Now().String()))
	if err != nil {
		fmt.Println("报错" + err.Error())
		return
	}
}
