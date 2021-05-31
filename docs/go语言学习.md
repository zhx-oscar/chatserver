## 切片
1. 如下代码表示一个数组的完整切片
```
x := [3]string{"abc", "def", "ghi"}
s := x[:] // a slice referencing the storage of x
```
2. 切片初始化时，使用make，第二个参数表示切片的元素数量，第三个参数表示切片的容量，如果make之后调用append，则会从第二个参数所设定的位置之后开始添加
```
slice := make([]int, 2)
slice = append(slice, 1) // 此处从切片的第三个位置开始添加
```
## errors
1. errors包提供了New函数可以输出错误信息
```
package main

import (
	"errors"
	"fmt"
)

func main() {
	err := errors.New("emit macho dwarf: elf header corrupted")
	if err != nil {
		fmt.Print(err)
	}
}

```
## go module
没有gomod之前，所有go代码必须放在$GOPATH/src 目录下，这样对工程的管理造成了不便，使用gomod可以将代码放在自己指定的任意目录下，方便了多工程的管理
### 使用方法
1. 在包含代码的目录下执行以下命令会生成go.mod文件
```
go mod init 工程名
```
2. 执行go build 会在当前目录生成可执行文件 

## 发送通道
```
chan<- int
```
## 接收通道
```
<-chan int
```
## 读写互斥锁
读写互斥锁sync.RWMutex允许只读操作并发执行，但写操作需要获得完全独享的访问权限。
1. 加锁解锁
```
var mu sync.RWMutex
mu.Lock() // 加锁
// 代码段
mu.Unlock() // 解锁
```
2. 只读锁加锁解锁
```
var mu sync.RWMutex
mu.RLock() // 加只读锁
// 代码段
mu.RUnlock() // 解只读锁
```

## pprof用法
1. 在import中加入pprof包，示例代码如下：
```
import (
    _ "net/http/pprof"
)
```
2. 在main函数中加入下面这段代码，注意要放在循环前面确保执行
```
go func() {
    http.ListenAndServe("0.0.0.0:6060", nil)
}()

```
3. 可以通过http://ip地址:6060/debug/pprof/  访问web界面
4. 要显示可视化的结果，需要在windows上使用graphviz分析结果
5. 在windows上安装go语言，使用以下命令可以查看图形化的prof结果
```
go tool pprof http://服务器ip:6060/debug/pprof/ + 具体要查看的程序报告类型
```
6. 可以查看的程序报告类型
```
allocs
block
cmdline
goroutine
heap
mutex
profile?seconds=60
threadcreate
trace
```
7. 输入web可在浏览器中查看结果
8. 查看火焰图可使用如下命令
```
curl -o trace.out http://服务器ip:6060/debug/pprof/profile?seconds=10
go tool pprof -http=:6061 trace.out
```
## log定向到多个位置
使用io.MultiWriter将log同时输出到屏幕和文件，示例代码如下
```
    Error *log.Logger
    
    f, err := os.OpenFile("../log/errors.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
    if err != nil {
        panic(err)
    }
 
    // 设置日志输出到文件
    // 定义多个写入器
    writers := []io.Writer{
        f,
        os.Stdout
    }
 
    fileAndStdoutWriter := io.MultiWriter(writers...)
    // 创建新的log对象
    Error = log.New(fileAndStdoutWriter, "", log.Ldate|log.Ltime|log.Lshortfile)

```

## 多协程程序退出
在新创建的协程中调用signal.Notify函数注册退出消息，当命令行键入退出操作时，每个注册的协程都会收到消息，注册需要用到协程内的声明的通道，示例代码如下：
```
go func() {
     c := make(chan os.Signal, 1)
     signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

    // todo:
    
LOOP:
    for {
        select {
            case <-c:
                break LOOP
            default:
        }
        
        // todo:
    } 
}
```
## golang 中string和int类型相互转换
```
// string转成int：
    int, err := strconv.Atoi(string)
// string转成int64：
    int64, err := strconv.ParseInt(string, 10, 64)
// int转成string：
    string := strconv.Itoa(int)
// int64转成string：
    string := strconv.FormatInt(int64,10)
```
## select 用法
当所有情况都没有发生的情况下，select会永久等待，如果不需要等待，需要加一个default情况
```
select {
case <-ch:
    // todo
defalut:
    // no op
}
```
## context 用法
1. 使用context可以利用cancel函数将退出消息传给子协程，子协程通过select语句接收到退出消息后即可退出子协程
2. 示例代码如下
```
func reqTask(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop", name)
			return
		default:
			fmt.Println(name, "send request")
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go reqTask(ctx, "worker1")
	time.Sleep(3 * time.Second)
	cancel()
	time.Sleep(3 * time.Second)
}
```
## recover 用法
1. recover必须放在defer语句中，可以捕获本协程内触发的panic，示例代码如下：
```
defer func() {
    if err := recover(); err != nil {
        Error.Printf("捕获异常: %v\n", err)
    }
}()

```
2. 子协程panic时，需要在子协程的defer语句中捕获panic，然后通过通道向主协程发送消息，通知主协程退出，示例代码如下：
```
var PanicChan chan interface{}

func init() {
    PanicChan = make(chan interface{}, 1)
}

func main() {
    GoB()
    time.Sleep(1 * time.Second)
    go func() {
        select {
        case err := <-PanicChan:
            fmt.Printf("receive panic msg: %v \n", err)
        }
    }()
    for {
    }
    fmt.Println("hello")
}

func GoB() {
    defer (func() {
        if err := recover(); err != nil {
            PanicChan <- fmt.Sprint(err)
            fmt.Println("recover:", err)
        }
    })()
    panic("Panic GoB")
}
```