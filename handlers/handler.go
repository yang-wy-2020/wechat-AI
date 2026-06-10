package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/yang-wy-2020/wechat-AI/config"
	"github.com/yang-wy-2020/wechat-AI/pkg/logger"
	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
	"github.com/skip2/go-qrcode"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var c = cache.New(config.LoadConfig().SessionTimeout, time.Minute*5)

var (
	processedMsgIDs   map[string]struct{}
	processedMsgMu    sync.Mutex
	processedMsgFile  = "processed_msg_ids.json"
	processedMsgOnce  sync.Once
)

func initProcessedMsgIDs() {
	processedMsgOnce.Do(func() {
		processedMsgIDs = make(map[string]struct{})
		data, err := os.ReadFile(processedMsgFile)
		if err != nil {
			return
		}
		var ids []string
		if err := json.Unmarshal(data, &ids); err != nil {
			return
		}
		for _, id := range ids {
			processedMsgIDs[id] = struct{}{}
		}
		logger.Info(fmt.Sprintf("loaded %d processed message IDs", len(processedMsgIDs)))
	})
}

func isMsgProcessed(id string) bool {
	processedMsgMu.Lock()
	defer processedMsgMu.Unlock()
	_, ok := processedMsgIDs[id]
	return ok
}

func markMsgProcessed(id string) {
	processedMsgMu.Lock()
	defer processedMsgMu.Unlock()
	processedMsgIDs[id] = struct{}{}
	if len(processedMsgIDs) > 1000 {
		newIDs := make(map[string]struct{}, 500)
		for k := range processedMsgIDs {
			if len(newIDs) >= 500 {
				break
			}
			newIDs[k] = struct{}{}
		}
		processedMsgIDs = newIDs
	}
	ids := make([]string, 0, len(processedMsgIDs))
	for k := range processedMsgIDs {
		ids = append(ids, k)
	}
	data, _ := json.Marshal(ids)
	os.WriteFile(processedMsgFile, data, 0644)
}

// MessageHandlerInterface 消息处理接口
type MessageHandlerInterface interface {
	handle() error
	ReplyText() error
}

// QrCodeCallBack 登录扫码回调，
func QrCodeCallBack(uuid string) {
	if runtime.GOOS == "windows" {
		// 运行在Windows系统上
		openwechat.PrintlnQrcodeUrl(uuid)
	} else {
		log.Println("login in linux")
		url := "https://login.weixin.qq.com/l/" + uuid
		log.Printf("如果二维码无法扫描，请缩小控制台尺寸，或更换命令行工具，缩小二维码像素")
		q, _ := qrcode.New(url, qrcode.High)
		fmt.Println(q.ToSmallString(true))
	}
}

func NewHandler() (msgFunc func(msg *openwechat.Message), err error) {
	dispatcher := openwechat.NewMessageMatchDispatcher()

	// 清空会话
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return strings.Contains(message.Content, config.LoadConfig().SessionClearToken)
	}, TokenMessageContextHandler())

	// 处理群消息
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsSendByGroup()
	}, GroupMessageContextHandler())

	// 好友申请
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsFriendAdd()
	}, func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		if config.LoadConfig().AutoPass {
			_, err := msg.Agree("")
			if err != nil {
				logger.Warning(fmt.Sprintf("add friend agree error : %v", err))
				return
			}
		}
	})

	// 私聊
	// 获取用户消息处理器
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return !(strings.Contains(message.Content, config.LoadConfig().SessionClearToken) || message.IsSendByGroup() || message.IsFriendAdd())
	}, UserMessageContextHandler())

	dispatch := openwechat.DispatchMessage(dispatcher)

	return func(msg *openwechat.Message) {
		initProcessedMsgIDs()
		if msg.MsgId != "" && isMsgProcessed(msg.MsgId) {
			logger.Info(fmt.Sprintf("skip replayed message: %s", msg.MsgId))
			return
		}
		if msg.MsgId != "" {
			markMsgProcessed(msg.MsgId)
		}
		dispatch(msg)
	}, nil
}
