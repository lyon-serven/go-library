package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// 已关闭微信通知接口，改为推荐使用 WxPusher 公众号消息推送，支持个人微信接收，且无需认证服务号。(废物)

// WxPusherNotifier 基于 WxPusher 的微信消息推送
// 官网：https://wxpusher.zjiecode.com
//
// 快速开始：
//  1. 访问 https://wxpusher.zjiecode.com，微信扫码登录
//  2. 创建应用，获取 AppToken
//  3. 关注 "WxPusher" 公众号，获取你的 UID
//  4. 即可接收消息到微信
//
// 使用示例：
//
//	n := message.NewWxPusherNotifier("AT_xxxxxx", "UID_xxxxxx")
//	n.Send(ctx, message.TextMsg("Redis 连接断开"))
//	n.Send(ctx, message.MarkdownMsg("告警", "**CPU** 超过 90%"))
type WxPusherNotifier struct {
	appToken string
	uids     []string // 接收消息的用户 UID 列表
	topicIDs []int    // 接收消息的主题 ID 列表（可选，用于群发）
	url      string   // 点击消息跳转的链接（可选）
	client   *http.Client
}

// NewWxPusherNotifier 创建 WxPusher 通知器
// appToken：应用 Token（AT_xxx）
// uids：接收消息的用户 UID 列表（UID_xxx），至少填一个
func NewWxPusherNotifier(appToken string, uids ...string) *WxPusherNotifier {
	return &WxPusherNotifier{
		appToken: appToken,
		uids:     uids,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// WithTopics 设置接收消息的主题 ID（群发场景）
func (w *WxPusherNotifier) WithTopics(topicIDs ...int) *WxPusherNotifier {
	w.topicIDs = topicIDs
	return w
}

// WithURL 设置点击消息后的跳转链接
func (w *WxPusherNotifier) WithURL(url string) *WxPusherNotifier {
	w.url = url
	return w
}

func (w *WxPusherNotifier) Name() string { return "wxpusher" }

// Send 发送消息到微信
func (w *WxPusherNotifier) Send(ctx context.Context, msg Message) error {
	// WxPusher contentType：1=文本 2=HTML 3=Markdown
	contentType := 1
	content := msg.Content

	switch msg.Type {
	case MsgTypeMarkdown:
		contentType = 3
		if msg.Title != "" {
			content = "## " + msg.Title + "\n\n" + msg.Content
		}
	case MsgTypeCard:
		contentType = 2 // HTML
		content = fmt.Sprintf("<h3>%s</h3><p>%s</p>", msg.Title, msg.Content)
	}

	payload := map[string]interface{}{
		"appToken":    w.appToken,
		"content":     content,
		"contentType": contentType,
		"uids":        w.uids,
		"url":         w.url,
	}
	if len(w.topicIDs) > 0 {
		payload["topicIds"] = w.topicIDs
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wxpusher: marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://wxpusher.zjiecode.com/api/send/message", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wxpusher: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("wxpusher: send request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if !result.Success {
		return fmt.Errorf("wxpusher: api error code=%d msg=%s", result.Code, result.Message)
	}
	return nil
}
