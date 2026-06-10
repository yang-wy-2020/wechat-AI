package baidu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/yang-wy-2020/wechat-AI/config"
	"github.com/yang-wy-2020/wechat-AI/pkg/llm"
	"github.com/yang-wy-2020/wechat-AI/pkg/logger"
)

type Provider struct {
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func (p *Provider) Name() string { return "baidu" }

const tokenURL = "https://aip.baidubce.com/oauth/2.0/token"
const chatURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/"

type tokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type requestBody struct {
	Messages    []llm.ChatMessage `json:"messages"`
	System      string            `json:"system,omitempty"`
	MaxTokens   uint              `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
}

type responseBody struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

func (p *Provider) getAccessToken() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.expiresAt) {
		return p.accessToken, nil
	}

	cfg := config.LoadConfig()
	if cfg.SecretKey == "" {
		return "", errors.New("baidu: secret_key required for client_secret")
	}

	reqURL := fmt.Sprintf("%s?grant_type=client_credentials&client_id=%s&client_secret=%s",
		tokenURL, url.QueryEscape(cfg.ApiKey), url.QueryEscape(cfg.SecretKey))

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("baidu get token error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result tokenResp
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	if result.AccessToken == "" {
		return "", errors.New("baidu: failed to get access_token, check api_key and secret_key")
	}

	p.accessToken = result.AccessToken
	p.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	return p.accessToken, nil
}

func (p *Provider) Chat(messages []llm.ChatMessage) (string, error) {
	cfg := config.LoadConfig()

	token, err := p.getAccessToken()
	if err != nil {
		return "", err
	}

	systemText := ""
	chatMessages := make([]llm.ChatMessage, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			systemText = m.Content
		} else {
			chatMessages = append(chatMessages, m)
		}
	}

	body := requestBody{
		Messages:    chatMessages,
		System:      systemText,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	modelName := cfg.Model
	if modelName == "" {
		modelName = "ernie-4.0-8k-latest"
	}
	reqURL := chatURL + modelName + "?access_token=" + token

	logger.Info(fmt.Sprintf("baidu request: %s", string(data)))

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("baidu api error, status: %d, body: %s", resp.StatusCode, string(respData)))
	}

	logger.Info(fmt.Sprintf("baidu response: %s", string(respData)))

	var result responseBody
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", errors.New(fmt.Sprintf("baidu api error: %s", result.Error))
	}

	if result.Result != "" {
		logger.Info(fmt.Sprintf("baidu reply: %s", result.Result))
		return result.Result, nil
	}

	return "", errors.New("baidu: empty response")
}
