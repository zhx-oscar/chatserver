package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

var wg sync.WaitGroup

type Chat struct {
	Length        int16  // 数据部分长度
	ChatMode      int8   //聊天类型
	PrAddrPortLen int16  // ip端口长度
	PrAddrPort    []byte // ip端口
	Content       []byte // 聊天内容
}

const (
	ChatLengthsz    = 2
	ChatModesz      = 1
	PrAddrPortLensz = 2
	PackageLengthsz = 2
	MsgTypesz       = 1
)

func (ch *Chat) Pack() []byte {
	bs := make([]byte, ch.Length)
	binary.BigEndian.PutUint16(bs[:ChatLengthsz], uint16(ch.Length))
	bs[ChatLengthsz] = byte(ch.ChatMode)
	binary.BigEndian.PutUint16(bs[ChatLengthsz+ChatModesz:ChatLengthsz+ChatModesz+PrAddrPortLensz], uint16(ch.PrAddrPortLen))
	copy(bs[ChatLengthsz+ChatModesz+PrAddrPortLensz:ChatLengthsz+ChatModesz+PrAddrPortLensz+ch.PrAddrPortLen], ch.PrAddrPort[:])
	copy(bs[ChatLengthsz+ChatModesz+PrAddrPortLensz+ch.PrAddrPortLen:], ch.Content[:])

	return bs
}

type Package struct {
	Length  int16  // 数据部分长度
	MsgType int8   // 消息类型
	Msg     []byte // 数据部分
}

func (p *Package) Pack(writer io.Writer) error {
	bs := make([]byte, p.Length)
	binary.BigEndian.PutUint16(bs[:PackageLengthsz], uint16(p.Length))
	bs[PackageLengthsz] = byte(p.MsgType)
	copy(bs[PackageLengthsz+MsgTypesz:], p.Msg[:])
	if _, err := writer.Write(bs); err != nil {
		return err
	}

	return nil
}

func (p *Package) Unpack(reader io.Reader) error {
	bs := make([]byte, PackageLengthsz+MsgTypesz)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}
	p.Length = int16(binary.BigEndian.Uint16(bs[:PackageLengthsz]))
	p.MsgType = int8(bs[PackageLengthsz])

	p.Msg = make([]byte, p.Length-(PackageLengthsz+MsgTypesz))
	if _, err := io.ReadFull(reader, p.Msg); err != nil {
		return err
	}

	return nil
}

var (
	chatType  = flag.Int("type", int(ChatType_Normal), "聊天类型:0.普通1.测试")
	sAddrPort = flag.String("saddrport", "localhost:8000", "服务器ip和端口")

	Error *log.Logger
)

func init() {
	f, err := os.OpenFile("../log/errors.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// 设置日志输出到文件
	// 定义多个写入器
	writers := []io.Writer{
		f,
		//     os.Stdout,
	}

	fileAndStdoutWriter := io.MultiWriter(writers...)
	// 创建新的log对象
	Error = log.New(fileAndStdoutWriter, "", log.Ldate|log.Ltime|log.Lshortfile)

	//syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())) // 将 stderr 重定向到 f
}

func main() {
	flag.Parse()

	//	sigs := make(chan os.Signal, 1)
	//	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1,
	//		syscall.SIGUSR2, syscall.SIGTSTP)

	if *chatType == int(ChatType_Test) {
		fmt.Println("请输入开启聊天协程数:")
		input := bufio.NewScanner(os.Stdin)
		var gCount, waitTime int
		var err error
		for input.Scan() {
			gCount, err = strconv.Atoi(input.Text())
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println("请输入开启聊天协程数:")
				continue
			}

			if gCount < 1 {
				fmt.Println("测试协程数必须大于等于1")
				continue
			}

			break
		}

		fmt.Println("请输入发送聊天消息的间隔:(单位毫秒)")
		for input.Scan() {
			waitTime, err = strconv.Atoi(input.Text())
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println("请输入发送聊天消息的间隔:(单位毫秒)")
				continue
			}

			if waitTime < 1 {
				fmt.Println("时间间隔必须大于等于1")
			}

			break
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
		}()

		for i := 0; i < gCount; i++ {
			wg.Add(1)
			go func(ctx context.Context) {
				newConn, err := net.Dial("tcp", *sAddrPort)
				if err != nil {
					Error.Println(err.Error())
					return
				}

				wg.Add(1)

				go handleConn(newConn, ctx, &wg, ChatType_Test)
			LOOP:
				for {
					select {
					case <-ctx.Done():
						break LOOP
					default:
					}

					time.Sleep(time.Duration(waitTime) * time.Millisecond)
					sendMsg(newConn, int8(CtoSMsgType_Chat), marshalChat(int8(ChatMode_Broadcast), []byte(""), []byte("This is a test")))
				}

				newConn.Close()
				wg.Done()
			}(ctx)
		}

		//		for {
		//			select {
		//			case <-sigs:
		//				fmt.Println("exitapp,sigs:", sigs)
		//				cancel()
		//				break
		//			default:
		//			}
		//		}

		wg.Wait()
	} else {
		var chatMode ChatMode
		fmt.Println("请输入聊天模式:\n1:广播模式\n2:私聊模式")
		input := bufio.NewScanner(os.Stdin)
	inputmode:
		for input.Scan() {
			switch string(input.Bytes()) {
			case "1":
				chatMode = ChatMode_Broadcast
				break inputmode
			case "2":
				chatMode = ChatMode_Private
				break inputmode
			default:
				fmt.Println("请输入聊天模式:\n1:广播模式\n2:私聊模式")
			}
		}

		conn, err := net.Dial("tcp", *sAddrPort)
		if err != nil {
			Error.Println(err.Error())
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
		}()

		var wg sync.WaitGroup
		wg.Add(1)

		go handleConn(conn, ctx, &wg, ChatType_Normal)

		if chatMode == ChatMode_Private {
			sendMsg(conn, int8(CtoSMsgType_GetUserList), []byte(""))
		} else if chatMode == ChatMode_Broadcast {
			broadcastChat(conn, os.Stdin)
		}

		//		for {
		//			select {
		//			case <-sigs:
		//				fmt.Println("exitapp,sigs:", sigs)
		//				cancel()
		//				break
		//			default:
		//			}
		//		}

		wg.Wait()
		conn.Close()
	}
}

func handleConn(conn net.Conn, ctx context.Context, wg *sync.WaitGroup, chatType ChatType) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if !atEOF {
			if len(data) >= PackageLengthsz+MsgTypesz {
				length := int16(0)
				binary.Read(bytes.NewReader(data[:PackageLengthsz]), binary.BigEndian, &length)
				if int(length) <= len(data) {
					return int(length), data[:int(length)], nil
				}
			}
		}
		return
	})
SCAN:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break
		default:
		}

		scannedPack := new(Package)

		if err := scannedPack.Unpack(bytes.NewReader(scanner.Bytes())); err != nil {
			Error.Println(err.Error())
			break SCAN
		}

		switch StoCMsgType(scannedPack.MsgType) {
		case StoCMsgType_UserList:
			if string(scannedPack.Msg) == "" {
				fmt.Println("现在没有可选择的私聊用户")
				break SCAN
			}

			userList := strings.Split(string(scannedPack.Msg), "|")
			for i, user := range userList {
				fmt.Println(fmt.Sprintf("%d: %s", i+1, user))
			}

			fmt.Println("请选择一个聊天用户编号:")
			input := bufio.NewScanner(os.Stdin)
			var userNo int
		inputmode:
			for input.Scan() {
				var err error
				if userNo, err = strconv.Atoi(string(input.Bytes())); err != nil {
					fmt.Println("请选择一个聊天用户编号:")
					continue
				}

				if userNo <= 0 || userNo > len(userList) {
					fmt.Println("输入编号不合法")
					continue
				}

				break inputmode
			}

			wg.Add(1)
			go privateChat(conn, os.Stdin, userList[userNo-1], wg, ctx)
		case StoCMsgType_Chat:
			if chatType == ChatType_Normal {
				fmt.Println(string(scannedPack.Msg))
			}
		}
	}

	wg.Done()
}

func privateChat(dst io.Writer, src io.Reader, privateUser string, wg *sync.WaitGroup, ctx context.Context) {
	input := bufio.NewScanner(src)
	for input.Scan() {
		select {
		case <-ctx.Done():
			break
		default:
		}

		sendMsg(dst, int8(CtoSMsgType_Chat), marshalChat(int8(ChatMode_Private), []byte(privateUser), []byte(input.Text())))
	}

	wg.Done()
}

func broadcastChat(dst io.Writer, src io.Reader) {
	input := bufio.NewScanner(src)
	for input.Scan() {
		sendMsg(dst, int8(CtoSMsgType_Chat), marshalChat(int8(ChatMode_Broadcast), []byte(""), []byte(input.Text())))
	}
}

func marshalChat(chatMode int8, prAddrPort []byte, content []byte) []byte {
	chat := &Chat{
		ChatMode:   chatMode,
		PrAddrPort: prAddrPort,
		Content:    content,
	}

	chat.PrAddrPortLen = int16(len(chat.PrAddrPort))
	chat.Length = ChatLengthsz + ChatModesz + PrAddrPortLensz + int16(len(chat.PrAddrPort)) + int16(len(chat.Content))

	return chat.Pack()
}

func sendMsg(dst io.Writer, msgType int8, msg []byte) {
	pack := &Package{
		MsgType: msgType,
		Msg:     msg,
	}

	pack.Length = PackageLengthsz + MsgTypesz + int16(len(pack.Msg))
	pack.Pack(dst)
}
