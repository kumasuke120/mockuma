# MocKuma [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Release](https://img.shields.io/github/release/kumasuke120/mockuma/all.svg)](https://github.com/kumasuke120/mockuma/releases/latest) [![Build Status](https://api.travis-ci.org/kumasuke120/mockuma.svg?branch=dev)](https://travis-ci.org/kumasuke120/mockuma) [![codecov](https://codecov.io/gh/kumasuke120/mockuma/branch/dev/graph/badge.svg)](https://codecov.io/gh/kumasuke120/mockuma)

[[English](README.md) | 中文]

MocKuma 是一款 API 接口的 Mock 工具。该工具读取命令化的 Json 映射配置文件，并根据配置生成对应的 Mock 接口。

前、后端开发人员使用本工具可以模拟 RESTful 接口以辅助开发以及单元测试； 
测试人员也可以使用本工具利用其命令式的映射配置进行参数匹配编写符合测试用例的接口辅助测试。

### 特性
- 根据请求参数/请求头映射返回
- 映射改变时，自动重新加载
- 使用用户定义的模板和变量渲染映射
- 支持静态文件引用
- 支持跳转和转发


## 安装
执行以下命令以在你的环境中安装 MocKuma
```
$ go get -u github.com/kumasuke120/mockuma/cmd/mockuma
```

如果想要避免麻烦或者没有 Go 的开发环境，请[点此](https://github.com/kumasuke120/mockuma/releases)以下载已发布版本的可执行文件。


## 快速开始

1. 确认 `$GOPATH\bin` 已经被添加到你的 `$PATH` 环境变量中；
2. 创建名为 `mockuMappings.json` 的文件，内容如下：
```json
[
  {
    "uri": "/",
    "method": "GET",
    "policies": [
      {
        "when": { "params": { "lang": "cn" } },
        "returns": {
          "headers": { "Content-Type": "text/plain; charset=utf-8" },
          "body": "你好，世界！"
        }
      },
      {
        "returns": {
          "headers": { "Content-Type": "text/plain" },
          "body": "Hello, World!"
        }
      }
    ]
  }
]
```
3. 以如下命令启动 MocKuma：
```
$ mockuma
```
4. 这样你就可以访问 [http://localhost:3214/](http://localhost:3214/) 或
[http://localhost:3214/?lang=cn](http://localhost:3214/?lang=cn) 来查看结果。 

#### 命令行参数
虽然 MocKuma 可以直接执行，但是它也提供了一些命令行参数供配置使用，以下是所有支持的命令行参数：

1. `-mapfile`: `MockuMappings` 映射配置文件路径，支持相对路径和绝对路径。
默认情况下，将会依次寻找当前目录下名为 `mockuMappings.json`，`mockuMappings.main.json` 的配置文件并读取加载。
特别的，MocKuma 的工作目录将会被设为该配置文件所在目录；
2. `-p`: MocKuma 监听端口号，默认值为 `3214`；
3. `--version`: 查看当前 MocKuma 的版本信息。

#### 更多示例
你可以点击[此处](example)来查看更多示例。
