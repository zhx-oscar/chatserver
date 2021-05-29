package util

import (
	"encoding/binary"
	"io"
)

type Chat struct {
	Length        uint16 // 消息长度
	ChatMode      int8   //聊天类型
	PrAddrPortLen uint16 // ip端口长度
	PrAddrPort    []byte // ip端口
	Content       []byte // 聊天内容
}

func (ch *Chat) Pack() []byte {
	bs := make([]byte, ch.Length)
	binary.BigEndian.PutUint16(bs[:ChatLengthSize], ch.Length)
	bs[ChatLengthSize] = byte(ch.ChatMode)
	binary.BigEndian.PutUint16(bs[ChatLengthSize+ChatModeSize:ChatLengthSize+ChatModeSize+PrAddrPortLenSize], ch.PrAddrPortLen)
	copy(bs[ChatLengthSize+ChatModeSize+PrAddrPortLenSize:ChatLengthSize+ChatModeSize+PrAddrPortLenSize+ch.PrAddrPortLen], ch.PrAddrPort[:])
	copy(bs[ChatLengthSize+ChatModeSize+PrAddrPortLenSize+ch.PrAddrPortLen:], ch.Content[:])

	return bs
}

// 聊天消息解包
func (ch *Chat) Unpack(msg []byte) error {
	ch.Length = binary.BigEndian.Uint16(msg[:ChatLengthSize])
	ch.ChatMode = int8(msg[ChatLengthSize])
	ch.PrAddrPortLen = binary.BigEndian.Uint16(msg[ChatLengthSize+ChatModeSize : ChatLengthSize+ChatModeSize+PrAddrPortLenSize])
	if ch.PrAddrPortLen > 0 {
		ch.PrAddrPort = make([]byte, ch.PrAddrPortLen)
		copy(ch.PrAddrPort, msg[ChatLengthSize+ChatModeSize+PrAddrPortLenSize:ChatLengthSize+ChatModeSize+PrAddrPortLenSize+ch.PrAddrPortLen])
	}

	ch.Content = make([]byte, ch.Length-(ChatLengthSize+ChatModeSize+PrAddrPortLenSize)-ch.PrAddrPortLen)
	copy(ch.Content, msg[(ChatLengthSize+ChatModeSize+PrAddrPortLenSize)+ch.PrAddrPortLen:])

	return nil
}

type Package struct {
	Length  uint16 // 数据部分长度
	MsgType int8   // 消息类型
	Msg     []byte // 数据部分
}

// 消息打包
func (p *Package) Pack(writer io.Writer) error {
	bs := make([]byte, p.Length)
	binary.BigEndian.PutUint16(bs[:PackageLengthSize], p.Length)
	bs[PackageLengthSize] = byte(p.MsgType)
	copy(bs[PackageLengthSize+MsgTypeSize:], p.Msg[:])

	if _, err := writer.Write(bs); err != nil {
		return err
	}

	return nil
}

// 消息解包
func (p *Package) Unpack(reader io.Reader) error {
	bs := make([]byte, PackageLengthSize+MsgTypeSize)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	p.Length = binary.BigEndian.Uint16(bs[:PackageLengthSize])
	p.MsgType = int8(bs[PackageLengthSize])

	p.Msg = make([]byte, p.Length-(PackageLengthSize+MsgTypeSize))
	if _, err := io.ReadFull(reader, p.Msg); err != nil {
		return err
	}
	return nil
}
