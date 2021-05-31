# 在vim中打开另一个文件
```
:e 文件路径
```

# vim go 安装
1. 首先确认GOPROXY环境变量是否设置

2. 先创建两个目录：
```
mkdir -p ~/.vim/autoload ~/.vim/bundle
```
3. 从https://github.com/tpope/vim-pathogen下载 vim-pathogen 到 ~/.vim/autoload 目录下
4. 编辑  ~/.vimrc  ，加入如下几行内容，再重启vim
```
execute pathogen#infect()
syntax on
filetype plugin indent on
```
5. 下载vim-go
```
cd ~/.vim/bundle/
git clone https://github.com/fatih/vim-go.git
```
6. 进入vim，执行vim-go提供的 :GoInstallBinaries

# vim 跳转到某一行
```
:数字
```
# 查看当前文件路径
ctrl + g

# 替换
全文替换：
```
:%s/foo/bar/g
```