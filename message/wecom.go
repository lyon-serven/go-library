package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WecomNotifier 企业微信机器人通知
// 文档：https://developer.work.weixin.qq.com/document/path/91770
//
// 使用示例：
//
//	n := message.NewWecomNotifier("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx")
//	n.Send(ctx, message.TextMsg("服务器告警"))
//	n.Send(ctx, message.MarkdownMsg("", "## 告警\n> CPU 超过 90%"))
type WecomNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewWecomNotifier 创建企业微信机器人通知器
func NewWecomNotifier(webhookURL string) *WecomNotifier {
	return &WecomNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WecomNotifier) Name() string { return "wecom" }

// Send 发送消息到企业微信
func (w *WecomNotifier) Send(ctx context.Context, msg Message) error {
	var payload map[string]interface{}

	switch msg.Type {
	case MsgTypeMarkdown:
		payload = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"content": msg.Content,
			},
		}
	case MsgTypeCard:
		// 企业微信模板卡片（news 类型近似实现）
		payload = map[string]interface{}{
			"msgtype": "news",
			"news": map[string]interface{}{
				"articles": []map[string]interface{}{
					{
						"title":       msg.Title,
						"description": msg.Content,
					},
				},
			},
		}
	default:
		// 纯文本，支持 @ 成员
		content := msg.Content
		for _, mobile := range msg.AtMobiles {
			content += fmt.Sprintf("\n<@%s>", mobile)
		}
		payload = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content":               content,
				"mentioned_mobile_list": msg.AtMobiles,
			},
		}
		if msg.AtAll {
			payload["text"].(map[string]interface{})["mentioned_mobile_list"] = []string{"@all"}
		}
	}

	return w.post(ctx, payload)
}

func (w *WecomNotifier) post(ctx context.Context, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wecom: marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("wecom: send request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wecom: api error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
