# netstat 命令
可以查询到目前主机开启的网络服务端口
例：netstat -aptn 可查询当前打开的端口，配合grep命令
```
netstat -aptn | grep 端口号
```
可查看某个端口是否打开
# read 命令
读取键盘输入的变量，用法:
```
read variable
```
# pkill 命令
杀死所有指定名称的进程，用法：
```
pkill pname
```

# top 命令
查看所有进程的信息，例:
```
top -u username   查看指定用户的进程
```

# lsof 命令
查看某个端口是否打开
```
lsof -i:端口号
```

# 杀死指定进程
1. 用命令ps aux | grep 进程名查找到pid
```
zx@zhangxiao:~$ ps aux | grep chat_client
zx        143850 25.0  0.2 1224388 8064 pts/3    Sl+  04:38   0:00 ./chat_client -type=1 -saddrport=172.16.16.19:8000
zx        143861  0.0  0.0   6432   676 pts/0    S+   04:38   0:00 grep --color=auto chat_client
```
其中143850就是 chat_client进程的pid

2. 用命令kill -9 pid即可杀死进程

# xshell 快捷键
1. 不小心按下 ctrl + s 导致xshell没反应，只要按下ctrl + q即可恢复

# ulimit
1. 查看系统最大文件打开数
```
ulimit -n
```
2. 设置系统最大文件打开数
```
ulimit -n 524288 
```

# 在目录中搜索关键字
1. 显示所有用到的行
```
rg -w 'name' ./
```
2. 显示所有用到的文件
```
rg -w 'name' ./ -l
```

# rm 命令
1. -f 即使原档案属性设为唯读，亦直接删除，无需逐一确认。
2. -r 将目录及以下之档案亦逐一删除。
例：删除目录和目录下所有文件一并删除
```
rm -rf 目录名
rm -rf * 删除当前目录下所有文件
```
