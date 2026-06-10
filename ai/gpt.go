package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yang-wy-2020/wechat-AI/config"
	"github.com/yang-wy-2020/wechat-AI/pkg/logger"
	"io/ioutil"
	"net/http"
	"time"
)

const BASEURL = "https://api.moonshot.cn/v1/"

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequestBody struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   uint          `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ChatResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage map[string]interface{} `json:"usage"`
}

// Completions 使用 chat/completions (messages 格式)
func Completions(msg string) (string, error) {
	cfg := config.LoadConfig()
	// 构造消息数组，只有一条 user 消息
	// 构建 messages：先插入 system（若配置），再插入 user
	messages := make([]ChatMessage, 0, 2)
	if cfg.SystemPrompt != "" {
		messages = append(messages, ChatMessage{Role: "system", Content: cfg.SystemPrompt})
	}
	messages = append(messages, ChatMessage{Role: "user", Content: msg})

	requestBody := ChatRequestBody{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	logger.Info(fmt.Sprintf("request gpt json string : %v", string(requestData)))
	req, err := http.NewRequest("POST", BASEURL+"chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := cfg.ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("请求 GTP 出错，状态码 %d，详情: %s", response.StatusCode, string(body)))
	}

	logger.Info(fmt.Sprintf("response gpt json string : %v", string(body)))

	gptResponseBody := &ChatResponseBody{}
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Message.Content
	}
	logger.Info(fmt.Sprintf("gpt response text: %s ", reply))
	return reply, nil
}
