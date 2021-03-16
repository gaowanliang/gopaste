# GoPaste

[![Go Version](https://img.shields.io/github/go-mod/go-version/gaowanliang/gopaste.svg?style=flat-square&label=Go&color=00ADD8)](https://github.com/gaowanliang/gopaste/blob/master/go.mod)
[![Release Version](https://img.shields.io/github/v/release/gaowanliang/gopaste.svg?style=flat-square&label=Release&color=1784ff)](https://github.com/gaowanliang/gopaste/releases/latest)
[![GitHub license](https://img.shields.io/github/license/gaowanliang/gopaste.svg?style=flat-square&label=License&color=2ecc71)](https://github.com/gaowanliang/gopaste/blob/master/LICENSE)
[![GitHub Star](https://img.shields.io/github/stars/gaowanliang/gopaste.svg?style=flat-square&label=Star&color=f39c12)](https://github.com/gaowanliang/gopaste/stargazers)
[![GitHub Fork](https://img.shields.io/github/forks/gaowanliang/gopaste.svg?style=flat-square&label=Fork&color=8e44ad)](https://github.com/gaowanliang/gopaste/network/members)

用Golang编写的现代自托管的剪切板服务。

## 开发缘由
存在许多剪切板服务，但都比它们原本需要的功能要复杂。
这就是为什么我决定用golang写一个剪切板服务。


支持`mysql`或`sqlite3`数据库驱动程序。

![Paste here](https://s3.ax1x.com/2021/03/16/6svC8g.png)
![Paste content](https://s3.ax1x.com/2021/03/16/6svKGF.png)

由于时间问题，同时此代码也被存档代码翻新并重新上传，目前只做了中文界面，如果您有时间，可以提交其他语言的语言包，非常感谢。

## 优点
* 使用简单，安装快捷
* 支持超过240种语言的代码高亮
* 支持创建短链接


## 开始
### 先决条件
* [go](https://golang.org/doc/install)
* gcc
* mysql (可选)


### Configuration

在运行pastengo之前，你应该编辑 ```config.json``` 。
```
Address : 用于重定向的基本url
Port : HTTP侦听端口
Length : 粘贴ID的长度
DBType : “sqlite3”或“mysql”，这取决于您的偏好
DBName : 数据库sqlite文件名或mysql名称
DBUsername : Mysql数据库用户名（sqlite不填写）
DBPassword : Mysql数据库密码 （sqlite不填写）
```

### Run

GoPaste将自动创建所需的SQL表。

1. 进入 [releases](https://github.com/gaowanliang/gopaste/releases) 下载最新版本压缩包，解压到任意文件夹（目前仅提供Windows amd64和 Linux amd64的编译程序，其他系统请自行编译）
2. 赋予`GoPaste`执行权限
3. 填写好config.json
4. 通过守护进程或者使用screen等程序运行


