package mirai

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

type Handler struct {
	c *Client
}

func (h *Handler) PrivateMessage(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
	h.c.logMessage(privateMessage, logFlagReceiving)
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypePrivate, privateMessage.Sender.Uin) {
				return false
			}
			return (mPoint.OnPrivateMessage != nil && mPoint.OnPrivateMessage(h.c, privateMessage)) ||
				(mPoint.OnMessage != nil && mPoint.OnMessage(h.c, privateMessage))
		})
}

func (h *Handler) GroupMessage(client *client.QQClient, groupMessage *message.GroupMessage) {
	h.c.logMessage(groupMessage, logFlagReceiving)
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypeGroup, groupMessage.GroupCode) {
				return false
			}
			return (mPoint.OnGroupMessage != nil && mPoint.OnGroupMessage(h.c, groupMessage)) ||
				(mPoint.OnMessage != nil && mPoint.OnMessage(h.c, groupMessage))
		},
	)
}

func (h *Handler) TempMessageEvent(qqClient *client.QQClient, tempMessage *client.TempMessageEvent) {
	h.c.logMessage(tempMessage, logFlagReceiving)
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypePrivate, tempMessage.Message.Sender.Uin) {
				return false
			}
			return (mPoint.OnTempMessage != nil && mPoint.OnTempMessage(h.c, tempMessage.Message)) ||
				(mPoint.OnMessage != nil && mPoint.OnMessage(h.c, tempMessage.Message))
		},
	)
}

func (h *Handler) NewFriendRequest(qqClient *client.QQClient, request *client.NewFriendRequest) {
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypePrivate, request.RequesterUin) {
				return false
			}
			return mPoint.OnNewFriendRequest != nil && mPoint.OnNewFriendRequest(h.c, request)
		},
	)
}

func (h *Handler) NewFriendEvent(qqClient *client.QQClient, event *client.NewFriendEvent) {
	qqClient.ReloadFriendList()
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypePrivate, event.Friend.Uin) {
				return false
			}
			return mPoint.OnNewFriendAdded != nil && mPoint.OnNewFriendAdded(h.c, event)
		},
	)
}

func (h *Handler) GroupInvitedRequest(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypeGroup, request.GroupCode) {
				return false
			}
			return mPoint.OnGroupInvited != nil && mPoint.OnGroupInvited(h.c, request)
		},
	)
}

func (h *Handler) MemberJoinGroupEvent(qqClient *client.QQClient, info *client.MemberJoinGroupEvent) {
	qqClient.ReloadGroupList()
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypeGroup, info.Group.Code) {
				return false
			}
			return mPoint.OnJoinGroup != nil && mPoint.OnJoinGroup(h.c, info)
		},
	)
}

func (h *Handler) GroupLeaveEvent(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
	qqClient.ReloadGroupList()
	h.c.steamPlugins(
		func(mPoint *Plugin) bool {
			if h.c.pluginBlocker != nil && h.c.pluginBlocker(mPoint, ContactTypeGroup, event.Group.Code) {
				return false
			}
			return mPoint.OnLeaveGroup != nil && mPoint.OnLeaveGroup(h.c, event)
		},
	)
}
