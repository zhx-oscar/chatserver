# 概述
聊天服务器分为客户端和服务端两部分，工程地址为：
git@gitlab.lovengame.com:zhangx1/chatserver.git 
可支持广播和私聊
# 使用方法
## 编译生成文件
输入make指令
## 启动服务端
1. cd 到 bin 目录下
2. 执行./chat_server -h 可查看参数
```
Usage of ./chat_server:
  -lport string
    	服务器监听端口 (default "8000")
  -pport string
    	服务器pprof端口 (default "6060")
```
3. 使用./chat_server 加参数启动或者直接启动
## 启动客户端
1. cd 到 bin目录下
2. 执行./chat_client -h 可查看参数
```
Usage of ./chat_client:
  -saddrport string
    	服务器ip和端口 (default "localhost:8000")
  -type int
    	聊天类型:0.普通1.测试

```
3. 使用./chat_client 加参数启动还或者直接启动
## 发起私聊
1. 客户端连接服务器成功后，可选择私聊模式
2. 客户端会显示所有在线的用户，可选择一个用户进行私聊
3. 连接成功后，只有输入的私聊客户端才能收到消息

## 执行服务器压测
1. cd 到bin目录下
2. 输入./chat_client -type=1 -saddrport="服务器ip端口" -type=1
3. 根据提示输入测试协程数
4. 根据提示输入消息发送间隔
