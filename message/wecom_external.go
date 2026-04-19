package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WecomExternalNotifier 企业微信客户群发通知
// 文档：https://developer.work.weixin.qq.com/document/path/92135
//
// 注意：此接口创建群发任务后，需要对应员工在企业微信 App 中确认才会真正发送。
//
// 使用示例：
//
//	n := message.NewWecomExternalNotifier("CORP_ID", "SECRET")
//	chatIDs, _ := n.ListGroupChats(ctx)
//	n.SendToGroups(ctx, "zhangsan", chatIDs, message.TextMsg("你好"))
type WecomExternalNotifier struct {
	corpID string
	secret string
	client *http.Client

	// access_token 缓存（复用 wecom_app 的逻辑）
	token    string
	tokenExp time.Time
}

// NewWecomExternalNotifier 创建客户群发通知器
// corpID: 企业 ID，secret: 「客户联系」功能对应应用的 Secret
func NewWecomExternalNotifier(corpID, secret string) *WecomExternalNotifier {
	return &WecomExternalNotifier{
		corpID: corpID,
		secret: secret,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WecomExternalNotifier) Name() string { return "wecom_external" }

// Send 实现 Notifier 接口，发给所有可见客户群（sender 留空由员工自选）
func (w *WecomExternalNotifier) Send(ctx context.Context, msg Message) error {
	return w.SendToGroups(ctx, "", nil, msg)
}

// SendToGroups 发送群发任务到指定客户群
// sender: 发送人 UserID（群发给客户群时必填，填群主的 UserID）
// chatIDs: 客户群 ID 列表，传 nil 则由员工自行选择
func (w *WecomExternalNotifier) SendToGroups(ctx context.Context, sender string, chatIDs []string, msg Message) error {
	token, err := w.getToken(ctx)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"chat_type": "group",
		"sender":    sender,
		"text": map[string]any{
			"content": msg.Content,
		},
	}
	if len(chatIDs) > 0 {
		payload["chat_id_list"] = chatIDs
	}

	// Markdown / Card 降级为文本（客户群发不支持 markdown）
	if msg.Type == MsgTypeCard && msg.Title != "" {
		payload["text"] = map[string]any{
			"content": fmt.Sprintf("%s\n\n%s", msg.Title, msg.Content),
		}
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/externalcontact/add_msg_template?access_token=%s", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom_external: create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("wecom_external: send failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode  int      `json:"errcode"`
		ErrMsg   string   `json:"errmsg"`
		MsgID    string   `json:"msgid"`
		FailList []string `json:"fail_list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wecom_external: api error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	if len(result.FailList) > 0 {
		return fmt.Errorf("wecom_external: some chats failed: %v", result.FailList)
	}
	return nil
}

// GroupChat 客户群信息
type GroupChat struct {
	ChatID  string `json:"chat_id"`
	Name    string `json:"name"`
	OwnerID string `json:"owner"`
}

// ListGroupChats 获取所有客户群列表（自动翻页）
func (w *WecomExternalNotifier) ListGroupChats(ctx context.Context) ([]GroupChat, error) {
	token, err := w.getToken(ctx)
	if err != nil {
		return nil, err
	}

	var all []GroupChat
	cursor := ""
	for {
		payload := map[string]any{
			"limit":         100,
			"status_filter": 0, // 0=所有群
		}
		if cursor != "" {
			payload["cursor"] = cursor
		}

		body, _ := json.Marshal(payload)
		url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/externalcontact/groupchat/list?access_token=%s", token)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("wecom_external: list groupchat request failed: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("wecom_external: list groupchat failed: %w", err)
		}

		var result struct {
			ErrCode   int    `json:"errcode"`
			ErrMsg    string `json:"errmsg"`
			GroupList []struct {
				ChatID string `json:"chat_id"`
				Status int    `json:"status"`
			} `json:"group_chat_list"`
			NextCursor string `json:"next_cursor"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("wecom_external: decode groupchat list failed: %w", err)
		}
		resp.Body.Close()

		if result.ErrCode != 0 {
			return nil, fmt.Errorf("wecom_external: list groupchat error code=%d msg=%s", result.ErrCode, result.ErrMsg)
		}

		for _, g := range result.GroupList {
			all = append(all, GroupChat{ChatID: g.ChatID})
		}

		if result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}
	return all, nil
}

func (w *WecomExternalNotifier) getToken(ctx context.Context) (string, error) {
	if w.token != "" && time.Now().Before(w.tokenExp) {
		return w.token, nil
	}
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", w.corpID, w.secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wecom_external: get token error code=%d msg=%s", result.ErrCode, result.ErrMsg)
	}
	w.token = result.AccessToken
	w.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return w.token, nil
}
