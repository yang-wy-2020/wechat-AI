package config

import (
	"encoding/json"
	"fmt"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// Configuration 项目配置
type Configuration struct {
	// gpt apikey
	ApiKey string `json:"api_key"`
	// 自动通过好友
	AutoPass bool `json:"auto_pass"`
	// 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout"`
	// GPT请求最大字符数
	MaxTokens uint `json:"max_tokens"`
	// GPT模型
	Model string `json:"model"`
	// 热度
	Temperature float64 `json:"temperature"`
	// 回复前缀
	ReplyPrefix string `json:"reply_prefix"`
	// 清空会话口令
	SessionClearToken string `json:"session_clear_token"`
	// 系统提示（发送给模型的 system 消息，用于设定回复风格，例如：简洁回复）
	SystemPrompt string `json:"system_prompt"`
	// 白名单：仅列出的好友昵称可以触发私聊自动回复（为空则表示全部允许）
	WhitelistUsers []string `json:"whitelist_users"`
	// 白名单：仅列出的群名称可以触发群聊自动回复（为空则表示全部允许）
	WhitelistGroups []string `json:"whitelist_groups"`
	// 回复间隔，支持字符串（"1s"）或数值（纳秒），例如 "1s" 或 1000000000
	ReplyInterval Duration `json:"reply_interval"`
}

// Duration wraps time.Duration to support flexible JSON unmarshalling from
// either a string like "1s" or a numeric value (nanoseconds).
type Duration time.Duration

// UnmarshalJSON accepts both string (e.g. "1s") and number (nanoseconds).
func (d *Duration) UnmarshalJSON(b []byte) error {
	// try string
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		dur, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*d = Duration(dur)
		return nil
	}
	// try number (nanoseconds)
	var n int64
	if err := json.Unmarshal(b, &n); err == nil {
		*d = Duration(time.Duration(n))
		return nil
	}
	return fmt.Errorf("invalid duration: %s", string(b))
}

// MarshalJSON writes duration as string
func (d Duration) MarshalJSON() ([]byte, error) {
	s := time.Duration(d).String()
	return json.Marshal(s)
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 给配置赋默认值
		config = &Configuration{
			AutoPass:          false,
			SessionTimeout:    60,
			MaxTokens:         512,
			Model:             "moonshot-v1-8k",
			Temperature:       0.9,
			SessionClearToken: "下一个问题",
			SystemPrompt:      "",
			WhitelistUsers:    []string{},
			WhitelistGroups:   []string{},
        		ReplyInterval:     Duration(time.Second),
		}

		// 判断配置文件是否存在，存在直接JSON读取
		_, err := os.Stat("config.json")
		if err == nil {
			f, err := os.Open("config.json")
			if err != nil {
				log.Fatalf("open config err: %v", err)
				return
			}
			defer f.Close()
			encoder := json.NewDecoder(f)
			err = encoder.Decode(config)
			if err != nil {
				log.Fatalf("decode config err: %v", err)
				return
			}
		}
		// 有环境变量使用环境变量
		ApiKey := os.Getenv("APIKEY")
		AutoPass := os.Getenv("AUTO_PASS")
		SessionTimeout := os.Getenv("SESSION_TIMEOUT")
		Model := os.Getenv("MODEL")
		MaxTokens := os.Getenv("MAX_TOKENS")
		Temperature := os.Getenv("TEMPREATURE")
		ReplyPrefix := os.Getenv("REPLY_PREFIX")
		SessionClearToken := os.Getenv("SESSION_CLEAR_TOKEN")
		ReplyIntervalEnv := os.Getenv("REPLY_INTERVAL")
		if ApiKey != "" {
			config.ApiKey = ApiKey
		}
		if AutoPass == "true" {
			config.AutoPass = true
		}
		if SessionTimeout != "" {
			duration, err := time.ParseDuration(SessionTimeout)
			if err != nil {
				logger.Danger(fmt.Sprintf("config session timeout err: %v ,get is %v", err, SessionTimeout))
				return
			}
			config.SessionTimeout = duration
		}
		if Model != "" {
			config.Model = Model
		}
		if MaxTokens != "" {
			max, err := strconv.Atoi(MaxTokens)
			if err != nil {
				logger.Danger(fmt.Sprintf("config MaxTokens err: %v ,get is %v", err, MaxTokens))
				return
			}
			config.MaxTokens = uint(max)
		}
		if Temperature != "" {
			temp, err := strconv.ParseFloat(Temperature, 64)
			if err != nil {
				logger.Danger(fmt.Sprintf("config Temperature err: %v ,get is %v", err, Temperature))
				return
			}
			config.Temperature = temp
		}
		if ReplyPrefix != "" {
			config.ReplyPrefix = ReplyPrefix
		}
		if SessionClearToken != "" {
			config.SessionClearToken = SessionClearToken
		}
		if ReplyIntervalEnv != "" {
			duration, err := time.ParseDuration(ReplyIntervalEnv)
			if err != nil {
				logger.Danger(fmt.Sprintf("config reply interval err: %v ,get is %v", err, ReplyIntervalEnv))
			} else {
				config.ReplyInterval = Duration(duration)
			}
		}

	})
	if config.ApiKey == "" {
		logger.Danger("config err: api key required")
	}

	return config
}
