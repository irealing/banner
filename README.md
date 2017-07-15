# binner

扫描HTTP服务器Server及默认页面title.

使用go语言开发.

[TOC]

## 安装

在开始这一切之前，我们假设你已经有了**golang**的编译环境。如果没有，请根据实际情况选择安装还是绕道:doge:

* 克隆项目到本地
* 在项目目录中执行go build 命令


```shell
$ cd $GOPATH/src
$ git clone gitee.com/irealing/banner.git
$ cd banner
$ go build
```

如果以上流程顺利执行没有报错的话，那么这个项目就成功编译了。

## 执行

### 命令行参数

```
-if : 输入文件;
-of : 输出文件;
-go : 启动的协程数;
-log : 日志级别(debug/info/warn/error)
-port : 端口文件名
```
*Unix-like系统限制ulimit数，go参数设置过大可能导致请求失败*


### 示例

```
$./binner -if input.txt -of target.txt -go 100
```

### 输入文件格式

 每行一个主机地址,如下:
 
```
118.244.113.239
42.225.64.15
211.149.249.24
118.192.164.61
118.126.142.116
```

### 输出文件格式

```
118.244.113.239,***,***
211.149.249.24,***,***
118.126.142.116,***,***
```

### 自定义端口及协议

添加`pots.csv`文件(目前未实现自定义端口文件名称),格式如下:

```csv
https,443
https,8443
http,80
http,8088
```
