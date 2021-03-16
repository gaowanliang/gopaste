[简体中文](readme_zh-CN.md)
# GoPaste

[![Go Version](https://img.shields.io/github/go-mod/go-version/gaowanliang/gopaste.svg?style=flat-square&label=Go&color=00ADD8)](https://github.com/gaowanliang/gopaste/blob/master/go.mod)
[![Release Version](https://img.shields.io/github/v/release/gaowanliang/gopaste.svg?style=flat-square&label=Release&color=1784ff)](https://github.com/gaowanliang/gopaste/releases/latest)
[![GitHub license](https://img.shields.io/github/license/gaowanliang/gopaste.svg?style=flat-square&label=License&color=2ecc71)](https://github.com/gaowanliang/gopaste/blob/master/LICENSE)
[![GitHub Star](https://img.shields.io/github/stars/gaowanliang/gopaste.svg?style=flat-square&label=Star&color=f39c12)](https://github.com/gaowanliang/gopaste/stargazers)
[![GitHub Fork](https://img.shields.io/github/forks/gaowanliang/gopaste.svg?style=flat-square&label=Fork&color=8e44ad)](https://github.com/gaowanliang/gopaste/network/members)

Modern self-hosted pastebin service written in Golang.

## Motivation
Many Pastebin services exist but all are more complicated than they need to be.
That is why I decided to write a pastebin service in golang.

although the code has severely been modified and functionalities altered.

Supports `mysql` or `sqlite3` database drivers.

![Paste here](https://s3.ax1x.com/2021/03/16/6svC8g.png)
![Paste content](https://s3.ax1x.com/2021/03/16/6svKGF.png)

Due to the time problem, at the same time this code is also archived code refurbished and re-uploaded, currently only the Chinese language interface is made, if you have time, you can submit the language pack for other languages, thank you very much.

## Advantage
* Easy to use and install
* Code highlighting in more than 240 languages supported
* Support for creating short links


## Getting started
### Prerequisities
* [go](https://golang.org/doc/install)
* gcc
* mysql (optionnal)


### Configuration

You should edit ```config.json``` before running GoPaste.
```
Address : Base url used for redirections
Port : HTTP listen port
Length : Length of paste ID
DBType : "sqlite3" or "mysql" based on your preferences
DBName : Database sqlite filename or mysql name 
DBUsername : Mysql database username (unused for sqlite)
DBPassword : Mysql database password (unused for sqlite)
```

### Run

GOPaste will automatically create the required SQL tables required.

1. Enter [releases](https://github.com/gaowanliang/gopaste/releases), download the latest version of the compressed package and unzip it to any folder (At present, only Windows and Linux AMD64 compilers are available. For other systems, please compile by yourself)
2. Give `gopaste` execution permission
3. Fill it in `config.json`
4. Run through the daemons or `screen`
