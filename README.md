mirai-framework
=====
一个基于MariGo的QQ机器人框架, 完全插件化的设计, 帮您轻而易举的建立属于自己的机器人, 对其增改插件, 同时保持更为清晰的代码结构。

# 设计思路

所有的功能都是由插件完成, 事件发生时, 调度器对插件循环调用, 插件响应是否处理该事件, 直至有插件响应事件, 插件发生异常, 或插件轮训结束, 最后日志结果被记录, 事件响应周期结束。
![img.png](images/invoke.png)

## 插件

- Id 插件的ID
- Name 插件的名称
- OnPrivateMessage 收到私聊消息时
- OnGroupMessage 收到组群消息时
- OnTempMessage 收到临时消息时
- OnMessage 收到消息时, 优先级低于明确类型的Message
- OnNewFriendRequest 收到好友请求时
- OnNewFriendAdded 添加了好友时
- OnGroupInvited 收到组群邀请时
- OnJoinGroup 加入组群时
- OnLeaveGroup 离开组群时

## 动作监听器

- Id 监听器的ID
- Name 监听器的名称
- OnSendPrivateMessage 发送了私聊消息将会执行回调
- OnSendGroupMessage 发送了组群消息将会执行回调
- OnSendTempMessage 发送了私聊消息将会执行回调

## 额外的api支持

- func (c *Client) MessageSenderUin 获得消息的发送者, 支持所有类型的消息
- func (c *Client) MessageElements 获得消息的组成, 支持所有类型的消息
- func (c *Client) MessageContent 获得消息的内容, 支持所有类型的消息
- func (c *Client) MessageFirstAt 获得消息中第一个AT的人
- func (c *Client) CardNameInGroup 获取群名片
- func (c *Client) MakeReplySendingMessage 创建一个回复消息, 如果是群员则自动带上@
- func (c *Client) ReplyRawMessage 快捷回复 将消息按照原来的路径发回, 群员将自动带上@
- func (c *Client) UploadReplyImage 上传图片, 接受人为消息源, 回复图片消息使用
- func (c *Client) UploadReplyVideo 上传视频, 接受人为消息源, 回复视频消息使用
- func (c *Client) AtElement 创建一个at
- func (c *Client) ReplyText 快速回复一个文本消息

## 插件拦截器

**client.SetPluginBlocker()** 可以实现插件拦截, 实现个别群个人启用禁用插件

# 如何使用

## 实现一个插件超级简单

```text
package hello

import "github.com/niuhuan/mirai-framework"

func PluginInstance() *mirai.Plugin {
	return &mirai.Plugin{
		Id: func() string {
			return "HELLO_WORLD"
		},
		Name: func() string {
			return "你好世界"
		},
		OnMessage: func(client *mirai.Client, messageInterface interface{}) bool {
			content := client.MessageContent(messageInterface)
			if content == "你好" {
				client.ReplyText(messageInterface, "世界")
				return true
			}
			return false
		},
	}
}
```

为什么用 struct 而不是 interface

- interface只需要选择其中几个func实现, 这种场景还是比较少见的
- 用interface会强制实现所有方法, 你需要实现太多方法了, 如果用embedded-struct将会失去IDE智能的提示

## 启动机器人

```text
  func main() {
      // 初始化手机机型等信息
      config.InitDeviceInfo()
      // 创建机器人
      client := mirai.NewClientMd5(Account.Uin, Account.PasswordBytes)
      // 注册插件
      client.SetActionListenersAndPlugins(
          nil,
          []*mirai.Plugin{
              hello.PluginInstance(),
          },
      )
      // 登录
      cmdLogin(client)
      // 等待退出信号
	  ch := make(chan os.Signal)
	  signal.Notify(ch, os.Interrupt, os.Kill)
	  <-ch
  }
```

- [InitDeviceInfo](https://github.com/niuhuan/mirai-bot/blob/master/config/device.go) 从设备读取机型等信息
- [cmdLogin](https://github.com/niuhuan/mirai-bot/blob/master/login/login.go) 处理登录验证码, 设备锁等功能

# 机器人模版

- [mirai-bot](https://github.com/niuhuan/mirai-bot)
