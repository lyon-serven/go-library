package message

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DingtalkNotifier 钉钉机器人通知
// 文档：https://open.dingtalk.com/document/robots/custom-robot-access
//
// 使用示例：
//
//	n := message.NewDingtalkNotifier("https://oapi.dingtalk.com/robot/send?access_token=xxx", "签名密钥")
//	n.Send(ctx, message.TextMsg("服务器告警"))
//	n.Send(ctx, message.MarkdownMsg("告警", "**CPU** 使用率超过 90%"))
type DingtalkNotifier struct {
	webhookURL string
	secret     string // 签名密钥（安全设置中选"加签"时填写，否则留空）
	client     *http.Client
}

// NewDingtalkNotifier 创建钉钉机器人通知器
// secret：钉钉机器人安全设置中"加签"的密钥，不使用加签时传空字符串
func NewDingtalkNotifier(webhookURL, secret string) *DingtalkNotifier {
	return &DingtalkNotifier{
		webhookURL: webhookURL,
		secret:     secret,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DingtalkNotifier) Name() string { return "dingtalk" }

// Send 发送消息到钉钉
func (d *DingtalkNotifier) Send(ctx context.Context, msg Message) error {
	var payload map[string]interface{}

	switch msg.Type {
	case MsgTypeMarkdown:
		payload = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"title": msg.Title,
				"text":  msg.Content,
			},
			"at": map[string]interface{}{
				"atMobiles": msg.AtMobiles,
				"isAtAll":   msg.AtAll,
			},
		}
	case MsgTypeCard:
		// 钉钉无原生卡片，用 ActionCard 近似实现
		payload = map[string]interface{}{
			"msgtype": "actionCard",
			"actionCard": map[string]interface{}{
				"title":          msg.Title,
				"text":           msg.Content,
				"btnOrientation": "0",
			},
		}
	default:
		// 纯文本
		payload = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": msg.Content,
			},
			"at": map[string]interface{}{
				"atMobiles": msg.AtMobiles,
				"isAtAll":   msg.AtAll,
			},
		}
	}

	return d.post(ctx, payload)
}

func (d *DingtalkNotifier) post(ctx context.Context, payload interface{}) error {
	url := d.webhookURL

	// 如果配置了签名密钥，添加时间戳和签名参数
	if d.secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := d.sign(timestamp)
		url = fmt.Sprintf("%s&timestamp=%d&sign=%s", url, timestamp, sign)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("dingtalk: marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("dingtalk: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk: send request failed: %w", err)
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
		return fmt.Errorf("dingtalk: api error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// sign 生成钉钉加签签名
// 算法：HMAC-SHA256(timestamp + "\n" + secret)，再 Base64
func (d *DingtalkNotifier) sign(timestamp int64) string {
	strToSign := fmt.Sprintf("%d\n%s", timestamp, d.secret)
	mac := hmac.New(sha256.New, []byte(d.secret))
	mac.Write([]byte(strToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
