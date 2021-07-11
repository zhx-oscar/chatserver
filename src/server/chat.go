package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	. "util"
)

func init() {
	errf, err := os.OpenFile("../log/errors.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	infof, err := os.OpenFile("../log/infos.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// 设置日志输出到文件
	// 定义多个写入器
	writers := []io.Writer{
		errf,
		os.Stdout}

	fileAndStdoutWriter := io.MultiWriter(writers...)
	// 创建新的log对象
	Error = log.New(fileAndStdoutWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infof, "", log.Ldate|log.Ltime|log.Lshortfile)
	//syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())) // 将 stderr 重定向到 f
}

var (
	Error      *log.Logger
	Info       *log.Logger
	listenPort = flag.String("lport", "8000", "服务器监听端口")
	pprofPort  = flag.String("pport", "6060", "服务器pprof端口")
	panicChan  = make(chan interface{}, 1)
)

func main() {
	srv := &http.Server{Addr: "0.0.0.0:" + *pprofPort}
	go func() {
		srv.ListenAndServe()
	}()

	flag.Parse()

	listener, err := net.Listen("tcp", "0.0.0.0:"+*listenPort)
	if err != nil {
		Error.Printf("listen err: %s", err.Error())
		timeout, _ := context.WithTimeout(context.Background(), 5*time.Minute)
		srv.Shutdown(timeout)

		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err := recover(); err != nil {
			Error.Printf("捕获异常: %v\n", err)
		}

		cancel()
		listener.Close()
		timeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(timeout)
	}()

	var wgMain sync.WaitGroup

	wgMain.Add(1)
	go serve(listener, ctx, &wgMain)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1,
		syscall.SIGUSR2, syscall.SIGTSTP)

	for {
		select {
		case <-sigs:
			fmt.Println("exitapp,sigs:", sigs)
			return
		case err := <-panicChan:
			fmt.Printf("receive panic msg:%v\n", err)
			return
		default:
		}
	}

	wgMain.Wait()
}

// 监听客户端连接请求，并创建连接
func serve(listener net.Listener, ctx context.Context, wgMain *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			panicChan <- fmt.Sprint(err)
		}

		fmt.Println("serve 退出")
	}()

	wgMain.Add(1)
	go broadcaster(ctx, wgMain)
	wgMain.Add(1)
	go logInfo(ctx, wgMain)

	var tempDelay time.Duration
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done")
			return
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				Error.Printf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)

				continue
			}

			select {
			default:
				Error.Printf(fmt.Sprintf("accept err: ", err))
			}

			if conn != nil {
				conn.Close()
			}

			break
		}

		wgMain.Add(1)
		go handleConn(conn, ctx, wgMain)
	}

	wgMain.Done()
}

type client chan<- *Package

var (
	mu        sync.RWMutex
	clients   = make(map[string]client)
	entering  = make(chan client)
	leaving   = make(chan client)
	messages  = make(chan *Package, 10000)
	bcMsgSent int64
)

// 处理广播聊天消息
func broadcaster(ctx context.Context, wgMain *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			panicChan <- fmt.Sprint(err)
		}

		fmt.Println("broadcaster退出")
	}()
	bcClients := make(map[client]bool)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messages:
			for cli := range bcClients {
				cli <- msg
				bcMsgSent++
			}
		case cli := <-entering:
			bcClients[cli] = true
		case cli := <-leaving:
			delete(bcClients, cli)
			close(cli)
		}
	}

	wgMain.Done()
}

// 统计消息处理数量
func logInfo(ctx context.Context, wgMain *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			panicChan <- fmt.Sprint(err)
		}

		fmt.Println("logInfo退出")
	}()
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			Info.Printf("广播模式在1秒内发送了%d条消息\n", bcMsgSent)
			Info.Printf("当前广播消息缓冲通道内的消息有%d条\n", len(messages))
			bcMsgSent = 0
		}
	}

	wgMain.Done()
}

// 客户端发送消息的处理
func handleConn(conn net.Conn, ctx context.Context, wgMain *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			panicChan <- fmt.Sprint(err)
		}

		fmt.Println("handleConn退出")
	}()

	var wg sync.WaitGroup

	// 读缓冲
	rch := make(chan string, 100)
	// 写缓冲
	wch := make(chan *Package, 100)

	// 消息逻辑处理
	wg.Add(1)
	go execMsgLogic(conn, rch, wch, &wg, ctx)

	// 写处理
	wg.Add(1)
	go clientWriter(conn, wch, &wg, ctx)

	who := conn.RemoteAddr().String()
	mu.Lock()
	clients[who] = wch
	mu.Unlock()

	sendMsg(wch, int8(StoCMsgType_Chat), []byte("You are "+who))
	sendMsg(messages, int8(StoCMsgType_Chat), []byte(who+" has arrived"))

	entering <- wch

	var buff, content []byte
	buff = make([]byte, 10240)
	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		readLen, err := conn.Read(buff)
		if err != nil {
			fmt.Println("读数据错误:", err)
			break
		}

		content = append(content, buff[:readLen]...)
		for len(content) >= int(PackageLengthSize+MsgTypeSize) {
			length := int16(0)
			binary.Read(bytes.NewReader(content[:PackageLengthSize]), binary.BigEndian, &length)
			if int(length) > len(content) {
				break
			}

			rch <- string(content[:int(length)])
			content = content[length:]
		}
	}

	close(rch)
	leaving <- wch
	sendMsg(messages, int8(StoCMsgType_Chat), []byte(who+" has left"))
	wg.Wait()
	conn.Close()
	mu.Lock()
	delete(clients, who)
	mu.Unlock()

	wgMain.Done()
}

// 消息相关逻辑处理
func execMsgLogic(conn net.Conn, rch <-chan string, wch chan<- *Package, wg *sync.WaitGroup, ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			panicChan <- fmt.Sprint(err)
		}

		fmt.Println("execMsgLogic退出")
	}()

	for msg := range rch {
		select {
		case <-ctx.Done():
			break
		default:
		}

		scannedPack := new(Package)
		if err := scannedPack.Unpack(bytes.NewReader([]byte(msg))); err != nil {
			Error.Println(err.Error())
			break
		}

		switch CtoSMsgType(scannedPack.MsgType) {
		case CtoSMsgType_GetUserList:
			var userList string
			mu.RLock()
			for who, _ := range clients {
				if who == conn.RemoteAddr().String() {
					continue
				}

				if userList != "" {
					userList += "|"
				}

				userList += who
			}
			mu.RUnlock()

			sendMsg(wch, int8(StoCMsgType_UserList), []byte(userList))

		case CtoSMsgType_Chat:
			chat := &Chat{}
			chat.Unpack(scannedPack.Msg)

			message := conn.RemoteAddr().String() + ":" + string(chat.Content)
			switch ChatMode(chat.ChatMode) {
			case ChatMode_Broadcast:
				sendMsg(messages, int8(StoCMsgType_Chat), []byte(message))
			case ChatMode_Private:
				var friendCh client
				var ok bool
				mu.RLock()
				friendCh, ok = clients[string(chat.PrAddrPort)]
				mu.RUnlock()
				if !ok {
					sendMsg(wch, int8(StoCMsgType_Chat), []byte("未找到私聊目标客户端"))
				}

				sendMsg(friendCh, int8(StoCMsgType_Chat), []byte(message))
			}
		}
	}

	wg.Done()
}

// 负责向客户端发送消息
func clientWriter(conn net.Conn, ch <-chan *Package, wg *sync.WaitGroup, ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			panicChan <- fmt.Sprint(err)
		}

		fmt.Println("clientWriter退出")
	}()

	for pack := range ch {
		select {
		case <-ctx.Done():
			break
		default:
		}

		if err := pack.Pack(conn); err != nil {
			Error.Println(err.Error())
		}
	}

	wg.Done()
}

// 消息发送函数封装
func sendMsg(ch chan<- *Package, msgType int8, msg []byte) {
	pack := &Package{
		MsgType: msgType,
		Msg:     msg,
	}
	pack.Length = PackageLengthSize + MsgTypeSize + uint16(len(pack.Msg))
	ch <- pack
}
