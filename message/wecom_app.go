package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WecomAppNotifier 企业微信应用消息通知（支持外部群/个人）
// 文档：https://developer.work.weixin.qq.com/document/path/90236
//
// 使用示例：
//
//	n := message.NewWecomAppNotifier("ww企业ID", "应用Secret", 1000001)
//	n.Send(ctx, message.TextMsg("你好"))                          // 发给应用可见范围内所有人
//	n.WithToUser("zhangsan", "lisi").Send(ctx, msg)              // 发给指定用户
//	n.WithToParty("1").Send(ctx, msg)                            // 发给指定部门
type WecomAppNotifier struct {
	corpID  string
	secret  string
	agentID int64
	client  *http.Client

	// 收件人，不设置则发给应用可见范围内全部成员（touser="@all"）
	toUser  string // 多个用 | 分隔，如 "zhangsan|lisi"
	toParty string // 部门 ID，多个用 | 分隔
	toTag   string // 标签 ID，多个用 | 分隔

	// access_token 缓存
	mu       sync.Mutex
	token    string
	tokenExp time.Time
}

// NewWecomAppNotifier 创建企业微信应用通知器
// corpID: 企业 ID（管理后台 -> 我的企业 -> 企业信息）
// secret: 应用 Secret（应用管理 -> 自建 -> 对应应用）
// agentID: 应用 AgentID
func NewWecomAppNotifier(corpID, secret string, agentID int64) *WecomAppNotifier {
	return &WecomAppNotifier{
		corpID:  corpID,
		secret:  secret,
		agentID: agentID,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// WithToUser 指定接收消息的用户（UserID），多个用 | 分隔，"@all" 表示全部
func (w *WecomAppNotifier) WithToUser(users ...string) *WecomAppNotifier {
	result := ""
	for i, u := range users {
		if i > 0 {
			result += "|"
		}
		result += u
	}
	w.toUser = result
	return w
}

// WithToParty 指定接收消息的部门 ID，多个用 | 分隔
func (w *WecomAppNotifier) WithToParty(parties ...string) *WecomAppNotifier {
	result := ""
	for i, p := range parties {
		if i > 0 {
			result += "|"
		}
		result += p
	}
	w.toParty = result
	return w
}

// WithToTag 指定接收消息的标签 ID，多个用 | 分隔
func (w *WecomAppNotifier) WithToTag(tags ...string) *WecomAppNotifier {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += "|"
		}
		result += t
	}
	w.toTag = result
	return w
}

func (w *WecomAppNotifier) Name() string { return "wecom_app" }

// Send 发送消息
func (w *WecomAppNotifier) Send(ctx context.Context, msg Message) error {
	token, err := w.getToken(ctx)
	if err != nil {
		return err
	}

	toUser := w.toUser
	if toUser == "" && w.toParty == "" && w.toTag == "" {
		toUser = "@all"
	}

	base := map[string]any{
		"touser":  toUser,
		"toparty": w.toParty,
		"totag":   w.toTag,
		"agentid": w.agentID,
	}

	switch msg.Type {
	case MsgTypeMarkdown:
		base["msgtype"] = "markdown"
		base["markdown"] = map[string]any{"content": msg.Content}
	case MsgTypeCard:
		base["msgtype"] = "textcard"
		base["textcard"] = map[string]any{
			"title":       msg.Title,
			"description": msg.Content,
			"url":         "https://work.weixin.qq.com",
		}
	default:
		base["msgtype"] = "text"
		base["text"] = map[string]any{"content": msg.Content}
	}

	body, _ := json.Marshal(base)
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom_app: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("wecom_app: send failed: %w", err)
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
		return fmt.Errorf("wecom_app: api error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// getToken 获取 access_token，带本地缓存（提前 5 分钟刷新）
func (w *WecomAppNotifier) getToken(ctx context.Context) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.token != "" && time.Now().Before(w.tokenExp) {
		return w.token, nil
	}

	url := fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		w.corpID, w.secret,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("wecom_app: get token request failed: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("wecom_app: get token failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("wecom_app: decode token response failed: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wecom_app: get token error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}

	w.token = result.AccessToken
	// 提前 5 分钟过期，避免边界问题
	w.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return w.token, nil
}
