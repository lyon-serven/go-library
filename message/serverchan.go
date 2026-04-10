package message

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ServerChanNotifier 基于 Server酱 的微信消息推送
// 官网：https://sct.ftqq.com
//
// 快速开始：
//  1. 访问 https://sct.ftqq.com，微信扫码登录
//  2. 获取 SendKey（格式：SCT_xxxxxx）
//  3. 关注绑定的服务号，即可在微信中收到消息
//
// 免费版：每天 5 条，足够告警使用
// 付费版：无限制
//
// 使用示例：
//
//	n := message.NewServerChanNotifier("SCT_xxxxxx")
//	n.Send(ctx, message.TextMsg("Redis 连接断开"))
//	n.Send(ctx, message.MarkdownMsg("🔴 Redis 告警", "**CPU** 超过 90%"))
type ServerChanNotifier struct {
	sendKey string
	client  *http.Client
}

// NewServerChanNotifier 创建 Server酱 通知器
// sendKey：在 https://sct.ftqq.com 获取，格式 SCT_xxxxxx
func NewServerChanNotifier(sendKey string) *ServerChanNotifier {
	return &ServerChanNotifier{
		sendKey: sendKey,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ServerChanNotifier) Name() string { return "serverchan" }

// Send 发送消息到微信
// Title 对应消息标题，Content 对应消息正文（支持 Markdown）
func (s *ServerChanNotifier) Send(ctx context.Context, msg Message) error {
	title := msg.Title
	content := msg.Content

	// 纯文本时 title 用 content 前 32 个字符
	if title == "" {
		runes := []rune(content)
		if len(runes) > 32 {
			title = string(runes[:32])
		} else {
			title = content
		}
	}

	apiURL := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", s.sendKey)

	form := url.Values{}
	form.Set("title", title)
	form.Set("desp", content)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("serverchan: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("serverchan: send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("serverchan: unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
