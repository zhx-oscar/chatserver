package util

type CtoSMsgType int8

const (
	CtoSMsgType_GetUserList CtoSMsgType = iota
	CtoSMsgType_Chat
)

type StoCMsgType int8

const (
	StoCMsgType_UserList StoCMsgType = iota
	StoCMsgType_Chat
)

type ChatType int

const (
	ChatType_Normal ChatType = iota
	ChatType_Test
)

type ChatMode int

const (
	ChatMode_Broadcast ChatMode = iota
	ChatMode_Private
)

const (
	ChatLengthSize    uint16 = 2 // 聊天消息长度字段字节数
	ChatModeSize      uint16 = 1 // 聊天类型字段字节数
	PrAddrPortLenSize uint16 = 2 // ip端口长度字段字节数
	PackageLengthSize uint16 = 2 // 消息长度字段字节数
	MsgTypeSize       uint16 = 1 // 消息类型字段字节数
)
