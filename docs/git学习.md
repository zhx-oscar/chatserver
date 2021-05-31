## linux环境为GitLab帐号添加SSH keys并连接GitLab
1. ssh-keygen -t rsa -C ”yourEmail@example.com”
2. 打开~/.ssh/id_rsa.pub文件，并且复制全部内容
3. 打开gitlab，在settings中选择SSH Keys，将内容复制到Key框中，点击Add key
## git 常用命令
1. git ad .
2. git clone 仓库地址.git
3. git commit -m "提交注释"
4. git pull
5. git push
6. git checkout 文件名 //还原本地文件
7. git rm

## git 切分支步骤
1. 查看远端分支
```
git branch -a
```
2. 查看本地分支
```
git branch
```
3. 切换分支
```
git checkout -b 本地分支名 远端分支名
```
## git 查看本地和远端文件的差异
1. 查看所有所有不同的文件
```
git diff
```
2. 查看当前目录下所有不同的文件
```
git diff .
```