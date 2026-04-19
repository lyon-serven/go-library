package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lyon-serven/go-library/message"
)

func main() {
	DemoWecomExternal()
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

func DemoWecom() {
	fmt.Println("=== 企业微信机器人消息通知示例 ===")
	fmt.Println("前往企业微信群 -> 添加机器人 -> 复制 Webhook 地址")

	// 替换为你自己的 Webhook Key
	webhookURL := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR-KEY-HERE"
	n := message.NewWecomNotifier(webhookURL)
	ctx := context.Background()

	// 1. 发送纯文本
	fmt.Println("\n1. 发送纯文本消息...")
	err := n.Send(ctx, message.TextMsg("服务器巡检完成，一切正常"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	// 2. 发送 Markdown
	fmt.Println("\n2. 发送 Markdown 消息...")
	content := "## 🔴 Redis 告警\n\n" +
		"| 项目 | 内容 |\n" +
		"|--|--|\n" +
		"| 服务 | Redis |\n" +
		"| 状态 | 连接断开 |\n" +
		"| 时间 | " + time.Now().Format("2006-01-02 15:04:05") + " |\n\n" +
		"> 请立即检查 Redis 服务状态"
	err = n.Send(ctx, message.MarkdownMsg("Redis 告警", content))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	// 3. 发送文本并 @ 指定成员
	fmt.Println("\n3. 发送文本并 @ 指定成员...")
	atMsg := message.Message{
		Type:      message.MsgTypeText,
		Content:   "请相关同学处理线上告警",
		AtMobiles: []string{"13800138000"},
	}
	err = n.Send(ctx, atMsg)
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	// 4. 发送文本并 @ 所有人
	fmt.Println("\n4. 发送文本并 @ 所有人...")
	atAllMsg := message.Message{
		Type:    message.MsgTypeText,
		Content: "紧急：生产环境数据库宕机，请所有人立即响应！",
		AtAll:   true,
	}
	err = n.Send(ctx, atAllMsg)
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	// 5. 发送卡片消息（news 类型）
	fmt.Println("\n5. 发送卡片消息...")
	err = n.Send(ctx, message.CardMsg("部署通知", "v1.2.3 已成功部署到生产环境"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}
}

func DemoWecomApp() {
	fmt.Println("=== 企业微信应用消息示例 ===")

	// 三个参数从企业微信管理后台获取：
	//   CorpID:  我的企业 -> 企业信息 -> 企业ID
	//   Secret:  应用管理 -> 自建 -> 选择应用 -> Secret
	//   AgentID: 应用管理 -> 自建 -> 选择应用 -> AgentId
	n := message.NewWecomAppNotifier("YOUR_CORP_ID", "YOUR_APP_SECRET", 1000001)
	ctx := context.Background()

	// 1. 发给应用可见范围内所有人
	fmt.Println("\n1. 发送文本给所有人...")
	err := n.Send(ctx, message.TextMsg("服务器巡检完成，一切正常"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	// 2. 发给指定用户（UserID 在后台「通讯录」里查看）
	fmt.Println("\n2. 发送 Markdown 给指定用户...")
	content := "## 部署通知\n\n" +
		"> **v1.2.3** 已成功部署到生产环境\n\n" +
		"时间：" + time.Now().Format("2006-01-02 15:04:05")
	err = n.WithToUser("zhangsan", "lisi").Send(ctx, message.MarkdownMsg("部署通知", content))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}

	// 3. 发给指定部门
	fmt.Println("\n3. 发送卡片消息给指定部门...")
	err = n.WithToParty("2").Send(ctx, message.CardMsg("告警通知", "Redis 连接断开，请立即处理"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 发送成功")
	}
}

func DemoWecomExternal() {
	fmt.Println("=== 企业微信客户群发消息示例 ===")

	// secret 用「客户联系」应用的 Secret，不是普通自建应用的
	n := message.NewWecomExternalNotifier("ww701dc44d44493e95", "o5GTOAn4ycWnvhmfhkJP85R0j_qUSZJMgNeU3CuPbjI")
	ctx := context.Background()

	// 1. 先拉取所有客户群列表
	fmt.Println("\n1. 获取客户群列表...")
	groups, err := n.ListGroupChats(ctx)
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
		return
	}
	fmt.Printf("   共 %d 个客户群\n", len(groups))
	for _, g := range groups {
		fmt.Printf("   - chat_id: %s\n", g.ChatID)
	}

	if len(groups) == 0 {
		fmt.Println("   没有可用的客户群")
		return
	}

	// 2. 发给所有客户群（sender 填群主的 UserID，必填）
	fmt.Println("\n2. 创建群发任务...")
	chatIDs := make([]string, 0, len(groups))
	for _, g := range groups {
		chatIDs = append(chatIDs, g.ChatID)
	}

	err = n.SendToGroups(ctx, "YOUR_SENDER_USERID", chatIDs, message.TextMsg("您好，感谢您的支持，有任何问题欢迎随时联系我们。"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 群发任务创建成功，请在企业微信 App 中确认发送")
	}

	// 3. 发给指定的某个群
	fmt.Println("\n3. 发给指定群...")
	err = n.SendToGroups(ctx, "YOUR_SENDER_USERID", []string{groups[0].ChatID}, message.TextMsg("这条消息只发给第一个群"))
	if err != nil {
		fmt.Println("   ❌ 失败:", err)
	} else {
		fmt.Println("   ✅ 群发任务创建成功")
	}
}
