package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FeishuNotifier 飞书机器人通知
// 文档：https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot
//
// 使用示例：
//
//	n := message.NewFeishuNotifier("https://open.feishu.cn/open-apis/bot/v2/hook/xxx")
//	n.Send(ctx, message.TextMsg("服务器告警"))
//	n.Send(ctx, message.MarkdownMsg("告警", "**CPU** 使用率超过 90%"))
type FeishuNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewFeishuNotifier 创建飞书机器人通知器
func NewFeishuNotifier(webhookURL string) *FeishuNotifier {
	return &FeishuNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *FeishuNotifier) Name() string { return "feishu" }

// Send 发送消息到飞书
func (f *FeishuNotifier) Send(ctx context.Context, msg Message) error {
	var payload map[string]interface{}

	switch msg.Type {
	case MsgTypeMarkdown, MsgTypeCard:
		// 飞书卡片消息（支持富文本）
		payload = map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"tag": "div",
						"text": map[string]interface{}{
							"content": msg.Content,
							"tag":     "lark_md",
						},
					},
				},
				"header": map[string]interface{}{
					"title": map[string]interface{}{
						"content": msg.Title,
						"tag":     "plain_text",
					},
				},
			},
		}
	default:
		// 纯文本消息
		payload = map[string]interface{}{
			"msg_type": "text",
			"content": map[string]interface{}{
				"text": msg.Content,
			},
		}
	}

	return f.post(ctx, payload)
}

func (f *FeishuNotifier) post(ctx context.Context, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("feishu: marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("feishu: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("feishu: send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu: unexpected status code: %d", resp.StatusCode)
	}

	// 解析飞书响应
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil // 解析失败不视为错误
	}
	if result.Code != 0 {
		return fmt.Errorf("feishu: api error code=%d msg=%s", result.Code, result.Msg)
	}
	return nil
}
