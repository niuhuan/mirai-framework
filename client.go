package mirai

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

// 初始化Logger

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(os.Stdout)
}

// 客户端对象

type Client struct {
	*client.QQClient
	Logger           *logrus.Logger
	actionsListeners []*ActionListener
	plugins          []*Plugin
}

func NewClient(uin int64, password string) *Client {
	return NewClientMd5(uin, md5.Sum([]byte(password)))
}

func NewClientMd5(uin int64, password [16]byte) *Client {
	c := &Client{
		Logger:   logrus.New(),
		QQClient: client.NewClientMd5(uin, password),
	}
	c.OnLog(func(qqClient *client.QQClient, event *client.LogEvent) {
		c.Logger.Debugf("%s : %s", event.Type, event.Message)
	})
	c.OnPrivateMessage(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		c.logMessage(privateMessage, logFlagReceiving)
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return ((*mPoint).OnPrivateMessage != nil && (*mPoint).OnPrivateMessage(c, privateMessage)) ||
					((*mPoint).OnMessage != nil && (*mPoint).OnMessage(c, privateMessage))
			})
	})
	c.OnGroupMessage(func(client *client.QQClient, groupMessage *message.GroupMessage) {
		c.logMessage(groupMessage, logFlagReceiving)
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return ((*mPoint).OnGroupMessage != nil && (*mPoint).OnGroupMessage(c, groupMessage)) ||
					((*mPoint).OnMessage != nil && (*mPoint).OnMessage(c, groupMessage))
			},
		)
	})
	c.OnTempMessage(func(qqClient *client.QQClient, tempMessage *client.TempMessageEvent) {
		c.logMessage(tempMessage, logFlagReceiving)
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return (*mPoint).OnTempMessage(c, tempMessage.Message) || (*mPoint).OnMessage(c, tempMessage.Message)
			},
		)
	})
	c.OnNewFriendRequest(func(qqClient *client.QQClient, request *client.NewFriendRequest) {
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return (*mPoint).OnNewFriendRequest(c, request)
			},
		)
	})
	c.OnNewFriendAdded(func(qqClient *client.QQClient, event *client.NewFriendEvent) {
		qqClient.ReloadFriendList()
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return (*mPoint).OnNewFriendAdded(c, event)
			},
		)
	})
	c.OnGroupInvited(func(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return (*mPoint).OnGroupInvited(c, request)
			},
		)
	})
	c.OnJoinGroup(func(qqClient *client.QQClient, info *client.GroupInfo) {
		qqClient.ReloadGroupList()
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return (*mPoint).OnJoinGroup(c, info)
			},
		)
	})
	c.OnLeaveGroup(func(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
		qqClient.ReloadGroupList()
		c.steamPlugins(
			func(mPoint *Plugin) bool {
				return (*mPoint).OnLeaveGroup(c, event)
			},
		)
	})
	return c
}

// SetActionListenersAndPlugins 设置监听器以及插件
func (c *Client) SetActionListenersAndPlugins(actionListeners []*ActionListener, plugins []*Plugin) {
	c.actionsListeners = actionListeners
	c.plugins = plugins
	var idList string
	var nameList string
	for i := 0; i < len(c.actionsListeners); i++ {
		l := c.actionsListeners[i]
		// ID校验
		if l.Id == nil || strings.TrimSpace(l.Id()) == "" {
			panic("actionsListeners的ID不可为空")
		}
		id := strings.TrimSpace(l.Id()) + ","
		if strings.Contains(idList, id) {
			panic("actionsListeners的ID不可重复")
		}
		idList += id
		// 名称校验
		if l.Name == nil || strings.TrimSpace(l.Name()) == "" {
			panic("actionsListeners的Name不可为空")
		}
		name := strings.TrimSpace(l.Name()) + ","
		if strings.Contains(nameList, name) {
			panic("actionsListeners的name不可重复")
		}
		nameList += name
	}
	idList = ""
	nameList = ""
	for i := 0; i < len(c.plugins); i++ {
		l := c.plugins[i]
		// ID校验
		if l.Id == nil || strings.TrimSpace(l.Id()) == "" {
			panic("plugins的ID不可为空")
		}
		id := strings.TrimSpace(l.Id()) + ","
		if strings.Contains(idList, id) {
			panic("plugins的ID不可重复 : " + id)
		}
		idList += id
		// 名称校验
		if l.Name == nil || strings.TrimSpace(l.Name()) == "" {
			panic("plugins的Name不可为空")
		}
		name := strings.TrimSpace(l.Name()) + ","
		if strings.Contains(nameList, name) {
			panic("plugins的name不可重复")
		}
		nameList += name
	}
}

// 遍历所有的插件, 插件调用失败的时候会recovery

func (c *Client) steamPlugins(fun func(plugin *Plugin) bool) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.Logger.Error(fmt.Sprintf("event error: %v\n%s", err, debug.Stack()))
			}
		}()
		for i := 0; i < len(c.plugins); i++ {
			if fun(c.plugins[i]) {
				c.Logger.Info(fmt.Sprintf("<<< PROCESS BY MODULE(%s)", (c.plugins[i]).Id()))
				return
			}
		}
		c.Logger.Info(fmt.Sprintf("<<< NOT PROCESS"))
	}()
}

func (c *Client) steamActionListeners(fun func(actionListener *ActionListener) bool) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.Logger.Error(fmt.Sprintf("action listener error: %v\n%s", err, debug.Stack()))
			}
		}()
		for i := 0; i < len(c.actionsListeners); i++ {
			if fun(c.actionsListeners[i]) {
				c.Logger.Info(fmt.Sprintf("<<< PROCESS BY MODULE(%s)", (c.actionsListeners[i]).Id()))
				return
			}
		}
		c.Logger.Info(fmt.Sprintf("<<< NOT PROCESS"))
	}()
}

// 覆盖

func (c *Client) SendPrivateMessage(target int64, m *message.SendingMessage) *message.PrivateMessage {
	message := c.QQClient.SendPrivateMessage(target, m)
	c.logMessage(message, logFlagReceiving)
	c.steamActionListeners(func(actionListener *ActionListener) bool {
		return actionListener.OnSendPrivateMessage != nil && actionListener.OnSendPrivateMessage(c, message)
	})
	return message
}

func (c *Client) SendGroupMessage(groupCode int64, m *message.SendingMessage, f ...bool) *message.GroupMessage {
	message := c.QQClient.SendGroupMessage(groupCode, m)
	c.logMessage(message, logFlagReceiving)
	c.steamActionListeners(func(actionListener *ActionListener) bool {
		return actionListener.OnSendGroupMessage != nil && actionListener.OnSendGroupMessage(c, message)
	})
	return message
}

func (c *Client) SendGroupTempMessage(groupCode, target int64, m *message.SendingMessage) *message.TempMessage {
	message := c.QQClient.SendGroupTempMessage(groupCode, target, m)
	c.logMessage(message, logFlagReceiving, target)
	c.steamActionListeners(func(actionListener *ActionListener) bool {
		return actionListener.OnSendTempMessage != nil && actionListener.OnSendTempMessage(c, message, target)
	})
	return message
}

// 工具方法

// MessageElements 获取消息的组成
func (c *Client) MessageElements(messageInterface interface{}) []message.IMessageElement {
	return MessageElements(messageInterface)
}

// MessageElements 获取消息的组成
func MessageElements(messageInterface interface{}) []message.IMessageElement {
	in := reflect.ValueOf(messageInterface).Elem().FieldByName("Elements").Interface()
	if array, ok := in.([]message.IMessageElement); ok {
		return array
	}
	return nil
}

// MessageContent 获取消息的文本内容
func (c *Client) MessageContent(messageInterface interface{}) string {
	return reflect.ValueOf(messageInterface).MethodByName("ToString").Call([]reflect.Value{})[0].String()
}

// MessageFirstAt 第一个At的用户
func (c *Client) MessageFirstAt(groupMessage *message.GroupMessage) int64 {
	return MessageFirstAt(groupMessage)
}

// MessageFirstAt 第一个At的用户
func MessageFirstAt(groupMessage *message.GroupMessage) int64 {
	for _, element := range groupMessage.Elements {
		if element.Type() == message.At {
			if at, ok := element.(*message.AtElement); ok {
				return at.Target
			}
		}
	}
	return 0
}

// CardNameInGroup 获取成员名称
func (c *Client) CardNameInGroup(groupCode int64, uin int64) string {
	for _, group := range c.GroupList {
		if group.Code == groupCode {
			for _, member := range group.Members {
				if member.Uin == uin {
					name := member.CardName
					if len(name) == 0 {
						name = member.Nickname
					}
					return name
				}
			}
			break
		}
	}
	return fmt.Sprintf("%d", uin)
}

// MessageSenderUin 获取消息的发送者
func (c *Client) MessageSenderUin(source interface{}) int64 {
	return MessageSenderUin(source)
}

// MessageSenderUin 获取消息的发送者
func  MessageSenderUin(source interface{}) int64 {
	if privateMessage, b := (source).(*message.PrivateMessage); b {
		return privateMessage.Sender.Uin
	} else if groupMessage, b := (source).(*message.GroupMessage); b {
		return groupMessage.Sender.Uin
	} else if tempMessage, b := (source).(*message.TempMessage); b {
		return tempMessage.Sender.Uin
	}
	return 0
}

// MakeReplySendingMessage 创建一个SendingMessage, 将会用于回复
func (c *Client) MakeReplySendingMessage(source interface{}) *message.SendingMessage {
	sending := message.NewSendingMessage()
	if groupMessage, b := (source).(*message.GroupMessage); b {
		sendGroupCode := groupMessage.GroupCode
		atUin := groupMessage.Sender.Uin
		return sending.Append(c.AtElement(sendGroupCode, atUin)).Append(message.NewText("\n\n"))
	}
	return sending
}

// ReplyRawMessage 回复一个消息到源消息, 对消息内容不做处理
func (c *Client) ReplyRawMessage(source interface{}, sendingMessage *message.SendingMessage) {
	if privateMessage, b := (source).(*message.PrivateMessage); b {
		c.SendPrivateMessage(privateMessage.Sender.Uin, sendingMessage)
	} else if groupMessage, b := (source).(*message.GroupMessage); b {
		c.SendGroupMessage(groupMessage.GroupCode, sendingMessage)
	} else if tempMessage, b := (source).(*message.TempMessage); b {
		c.SendGroupTempMessage(tempMessage.GroupCode, tempMessage.Sender.Uin, sendingMessage)
	}
}

// UploadReplyImage 上传文件用作回复
func (c *Client) UploadReplyImage(source interface{}, buffer []byte) (message.IMessageElement, error) {
	if privateMessage, b := (source).(*message.PrivateMessage); b {
		return c.UploadPrivateImage(privateMessage.Sender.Uin, bytes.NewReader(buffer))
	} else if groupMessage, b := (source).(*message.GroupMessage); b {
		return c.UploadGroupImage(groupMessage.GroupCode, bytes.NewReader(buffer))
	} else if tempMessage, b := (source).(*message.TempMessage); b {
		return c.UploadPrivateImage(tempMessage.Sender.Uin, bytes.NewReader(buffer))
	}
	return nil, errors.New("!")
}

// UploadReplyVideo 上传视频文件
func (c *Client) UploadReplyVideo(source interface{}, video []byte, thumb []byte) (*message.ShortVideoElement, error) {
	groupCode := int64(0)
	if groupMessage, b := (source).(*message.GroupMessage); b {
		groupCode = groupMessage.GroupCode
	}
	return c.UploadGroupShortVideo(groupCode, bytes.NewReader(video), bytes.NewReader(thumb))
}

// AtElement 创建一个At
func (c *Client) AtElement(groupCode int64, uin int64) *message.AtElement {
	return message.NewAt(uin, fmt.Sprintf("@%s", c.CardNameInGroup(groupCode, uin)))
}

// ReplyText 快捷回复消息
func (c *Client) ReplyText(source interface{}, content string) {
	c.ReplyRawMessage(source, c.MakeReplySendingMessage(source).Append(message.NewText(content)))
}

// 插件

type Plugin struct {
	Id                 func() string
	Name               func() string
	OnPrivateMessage   func(client *Client, privateMessage *message.PrivateMessage) bool
	OnGroupMessage     func(client *Client, groupMessage *message.GroupMessage) bool
	OnTempMessage      func(client *Client, tempMessage *message.TempMessage) bool
	OnMessage          func(client *Client, messageInterface interface{}) bool
	OnNewFriendRequest func(client *Client, request *client.NewFriendRequest) bool
	OnNewFriendAdded   func(client *Client, event *client.NewFriendEvent) bool
	OnGroupInvited     func(client *Client, info *client.GroupInvitedRequest) bool
	OnJoinGroup        func(client *Client, info *client.GroupInfo) bool
	OnLeaveGroup       func(client *Client, event *client.GroupLeaveEvent) bool
}

// 监听器

type ActionListener struct {
	Id                   func() string
	Name                 func() string
	OnSendPrivateMessage func(c *Client, message *message.PrivateMessage) bool
	OnSendGroupMessage   func(c *Client, message *message.GroupMessage) bool
	OnSendTempMessage    func(c *Client, message *message.TempMessage, target int64) bool
}

// 显示用的日志

const logFlagReceiving = "RECEIVING"
const logFlagSending = "SENDING"

func (c *Client) logMessage(m interface{}, logFlag string, ext ...interface{}) {

	var flag string
	var entries []message.IMessageElement

	if logFlag == logFlagSending {
		flag = "<<< Sending <<<"
	}
	if logFlag == logFlagReceiving {
		flag = ">>> Receiving >>>"
	}

	if privateMessage, ok := m.(*message.PrivateMessage); ok {
		entries = privateMessage.Elements
		flag += " PRIVATE :"
		if logFlag == logFlagSending {
			flag += fmt.Sprintf(" UID(%d) ", privateMessage.Target)
		}
		if logFlag == logFlagReceiving {
			flag += fmt.Sprintf(" UID(%d) ", privateMessage.Sender.Uin)
		}
	}
	if groupMessage, ok := m.(*message.GroupMessage); ok {
		entries = groupMessage.Elements
		flag += " GROUP :"
		if logFlag == logFlagSending {
			flag += fmt.Sprintf(" GID(%d) ", groupMessage.GroupCode)
		}
		if logFlag == logFlagReceiving {
			flag += fmt.Sprintf(" GID(%d) UID(%d) ", groupMessage.GroupCode, groupMessage.Sender.Uin)
		}
	}
	if tempMessage, ok := m.(*message.TempMessage); ok {
		entries = tempMessage.Elements
		flag += " TEMP :"
		if logFlag == logFlagSending {
			flag += fmt.Sprintf(" GID(%d) ", tempMessage.GroupCode)
			if len(ext) > 0 {
				if id, ok := ext[0].(int64); ok {
					flag += fmt.Sprintf("UID(%d) ", id)
				}
			}
		}
		if logFlag == logFlagReceiving {
			flag += fmt.Sprintf(" GID(%d) UID(%d) ", tempMessage.GroupCode, tempMessage.Sender.Uin)
		}
	}

	contentBuff, e := c.FormatMessageElements(entries)

	if e != nil {
		logger.Error("LOG ERROR : ", flag, " : ", e.Error())
	}
	content := string(contentBuff)

	builder := strings.Builder{}
	builder.WriteString(flag)
	builder.WriteString("\n")
	builder.WriteString(content)
	logger.Info(builder.String())
}

func (c *Client) FormatMessageElements(entries []message.IMessageElement) ([]byte, error) {
	var fEntries []interface{}

	for i := range entries {
		if app, b := (entries[i]).(*message.LightAppElement); b {
			fEntries = append(fEntries, map[string]string{
				"Type":    "LightAPP",
				"Content": app.Content,
			})
		} else if text, b := (entries[i]).(*message.TextElement); b {
			fEntries = append(fEntries, map[string]string{
				"Type":    "Text",
				"Content": text.Content,
			})
		} else if img, b := (entries[i]).(*message.GroupImageElement); b {
			fEntries = append(fEntries, map[string]string{
				"ImageId": img.ImageId,
				"Type":    "Image",
				"Url":     img.Url,
			})
		} else if img, b := (entries[i]).(*message.FriendImageElement); b {
			fEntries = append(fEntries, map[string]string{
				"ImageId": img.ImageId,
				"Type":    "Image",
				"Url":     img.Url,
			})
		} else if at, b := (entries[i]).(*message.AtElement); b {
			fEntries = append(fEntries, map[string]interface{}{
				"Type":    "At",
				"Target":  at.Target,
				"Display": at.Display,
			})
		} else if voice, b := (entries[i]).(*message.VoiceElement); b {
			fEntries = append(fEntries, map[string]string{
				"Type":    "Voice",
				"Name":    voice.Name,
				"Display": voice.Url,
			})
		} else if redBag, b := (entries[i]).(*message.RedBagElement); b {
			fEntries = append(fEntries, map[string]interface{}{
				"Type":   "RegBag",
				"Title":  redBag.Title,
				"RbType": int(redBag.MsgType),
			})
		} else {
			fEntries = append(fEntries, map[string]interface{}{
				"Type":    "Other",
				"SubType": int(entries[i].Type()),
			})
		}
	}
	return json.Marshal(&fEntries)
}
