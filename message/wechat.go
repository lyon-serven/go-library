package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WechatNotifier 微信公众号模板消息通知
// 文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
//
// 注意：微信公众号模板消息需要：
//  1. 已认证的服务号（订阅号不支持）
//  2. 在公众平台申请模板
//  3. 用户已关注公众号
//
// 使用示例：
//
//	n := message.NewWechatNotifier("AppID", "AppSecret")
//	err := n.Send(ctx, message.Message{
//	    Type:    message.MsgTypeCard,
//	    Title:   "服务告警",
//	    Content: "CPU 使用率超过 90%",
//	})
type WechatNotifier struct {
	appID     string
	appSecret string
	// TemplateID 模板消息 ID（在公众平台配置）
	TemplateID string
	// ToUser 接收者 OpenID 列表
	ToUsers []string
	// RedirectURL 点击消息跳转的链接（可选）
	RedirectURL string

	client *http.Client

	// 内部 token 缓存
	accessToken   string
	tokenExpireAt time.Time
}

// NewWechatNotifier 创建微信公众号通知器
func NewWechatNotifier(appID, appSecret string) *WechatNotifier {
	return &WechatNotifier{
		appID:     appID,
		appSecret: appSecret,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WechatNotifier) Name() string { return "wechat" }

// Send 向所有 ToUsers 发送模板消息
func (w *WechatNotifier) Send(ctx context.Context, msg Message) error {
	token, err := w.getAccessToken(ctx)
	if err != nil {
		return err
	}

	for _, openID := range w.ToUsers {
		if err := w.sendTemplate(ctx, token, openID, msg); err != nil {
			return fmt.Errorf("wechat: send to %s failed: %w", openID, err)
		}
	}
	return nil
}

// sendTemplate 向单个用户发送模板消息
func (w *WechatNotifier) sendTemplate(ctx context.Context, token, openID string, msg Message) error {
	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", token)

	payload := map[string]interface{}{
		"touser":      openID,
		"template_id": w.TemplateID,
		"url":         w.RedirectURL,
		"data": map[string]interface{}{
			"first": map[string]interface{}{
				"value": msg.Title,
				"color": "#173177",
			},
			"content": map[string]interface{}{
				"value": msg.Content,
				"color": "#173177",
			},
			"remark": map[string]interface{}{
				"value": time.Now().Format("2006-01-02 15:04:05"),
				"color": "#999999",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("wechat: marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("wechat: send request failed: %w", err)
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
		return fmt.Errorf("wechat: api error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// getAccessToken 获取微信 access_token（带缓存，有效期 2 小时）
func (w *WechatNotifier) getAccessToken(ctx context.Context) (string, error) {
	// token 未过期直接返回
	if w.accessToken != "" && time.Now().Before(w.tokenExpireAt) {
		return w.accessToken, nil
	}

	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(w.appID),
		url.QueryEscape(w.appSecret),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("wechat: get access_token failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"` // 有效期（秒），通常 7200
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("wechat: decode token response failed: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat: get token error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}

	// 缓存 token，提前 5 分钟过期以防边界问题
	w.accessToken = result.AccessToken
	w.tokenExpireAt = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)

	return w.accessToken, nil
}
