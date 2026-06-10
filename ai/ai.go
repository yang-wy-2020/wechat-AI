package ai

import (
	"fmt"

	"github.com/yang-wy-2020/wechat-AI/ai/baidu"
	"github.com/yang-wy-2020/wechat-AI/ai/claude"
	"github.com/yang-wy-2020/wechat-AI/ai/moonshot"
	"github.com/yang-wy-2020/wechat-AI/ai/openai"
	"github.com/yang-wy-2020/wechat-AI/ai/qwen"
	"github.com/yang-wy-2020/wechat-AI/config"
	"github.com/yang-wy-2020/wechat-AI/pkg/llm"
	"github.com/yang-wy-2020/wechat-AI/pkg/logger"
)

var providers = map[string]llm.Provider{
	"moonshot": &moonshot.Provider{},
	"openai":   &openai.Provider{},
	"claude":   &claude.Provider{},
	"baidu":    &baidu.Provider{},
	"qwen":     &qwen.Provider{},
}

func Completions(msg string) (string, error) {
	cfg := config.LoadConfig()

	p, ok := providers[cfg.Provider]
	if !ok {
		return "", fmt.Errorf("unknown ai provider: %s, available: moonshot, openai, claude, baidu, qwen", cfg.Provider)
	}

	logger.Info(fmt.Sprintf("using provider: %s, model: %s", p.Name(), cfg.Model))

	messages := make([]llm.ChatMessage, 0, 2)
	if cfg.SystemPrompt != "" {
		messages = append(messages, llm.ChatMessage{Role: "system", Content: cfg.SystemPrompt})
	}
	messages = append(messages, llm.ChatMessage{Role: "user", Content: msg})

	return p.Chat(messages)
}
