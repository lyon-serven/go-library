// Package message 提供飞书、钉钉、企业微信、微信公众号的消息通知封装
package message

import "context"

// MsgType 消息类型
type MsgType string

const (
	MsgTypeText     MsgType = "text"     // 纯文本
	MsgTypeMarkdown MsgType = "markdown" // Markdown
	MsgTypeCard     MsgType = "card"     // 卡片（富文本）
)

// Message 统一消息结构
type Message struct {
	// Type 消息类型
	Type MsgType
	// Title 标题（Markdown / Card 类型使用）
	Title string
	// Content 消息正文
	Content string
	// AtMobiles 需要 @ 的手机号列表（钉钉 / 企业微信支持）
	AtMobiles []string
	// AtAll 是否 @ 所有人
	AtAll bool
}

// Notifier 消息通知接口
type Notifier interface {
	// Send 发送消息
	Send(ctx context.Context, msg Message) error
	// Name 返回通知渠道名称
	Name() string
}

// TextMsg 快速构建纯文本消息
func TextMsg(content string) Message {
	return Message{Type: MsgTypeText, Content: content}
}

// MarkdownMsg 快速构建 Markdown 消息
func MarkdownMsg(title, content string) Message {
	return Message{Type: MsgTypeMarkdown, Title: title, Content: content}
}

// CardMsg 快速构建卡片消息
func CardMsg(title, content string) Message {
	return Message{Type: MsgTypeCard, Title: title, Content: content}
}
