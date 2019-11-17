# MocKuma
这是一款使用 Go 编写的 Http 接口 Mock 工具。该工具读取命令化的 Json 映射配置文件，并根据配置 Mock 出对应的接口。

## 构建与运行
使用 `go get` 命令或下载压缩包解压在 `$GOPATH/github.com/kumasuke120/mockuma` 中，进入该目录并执行以下代码进行构建：
```
go build -o bin/mockuma github.com/kumasuke120/mockuma/cmd
```

构建成功后，使用以下命令运行：
```
cd bin && ./mockuma
```
使用默认参数运行时，工具程序将会读取当前目录下的 `mockuMappings.json`文件并生成对应 Mock 接口，以下是所有支持的命令行参数：

| **参数** | **说明** | **默认值** | **示例** |
|----------|------------------------------------------------|--------------------|-----------------------------|
| -mapfile | MockuMappings 映射配置文件 | mockuMappings.json | -mapfile=C:\mockuMappings.json |
| -p | 工具程序监听端口号 | 3214 | -p=3214 |
| --help | 查看帮助，内容为英文 | -- | -- |

## MockuMappings 配置
`MockuMappings` 是 MocKuma 自有的配置文件格式。它是一个有着特定规则的 `.json` 格式文件，以下为示例：
```json
[
  {
    "uri": "/api/get",
    "method": "GET",
    "policies": [
      {
        "when": {
          "params": {
            "value1": [
              1
            ],
            "value2": 2
          }
        },
        "returns": {
          "headers": {
            "Server": "MocKuma-Mappings/1.0"
          },
          "body": "{\"code\": 2000, \"message\": \"Hello, 世界!\"}"
        }
      },
      {
        "returns": {
          "statusCode": 201,
          "headers": {
            "Content-Type": "application/json; charset=utf8"
          },
          "body": "{\"code\": 2000, \"message\": \"Hello, World!\"}"
        }
      }
    ]
  }
]
```
`MockuMappings` 顶层为一个 Json 数组，数组的每一项均是一个 Json 对象。这种 Json 对象中有 3 个参数:

| **参数** | **说明** | **示例** |
|----------|-------------------------------------------------------------------|--------------|
| uri | （必填）Mock 接口的 Uri，必须以 / 开头 | `"/api/example"` |
| method | （选填，默认 Any）Mock 接口绑定的请求方式，支持 GET、POST、PUT、DELETE 以及 Any；<br>其中前四项为单独绑定， Any 为绑定所有请求方式 | `"GET"` |
| policies | （必填）Mock 接口的映射策略，返回 Mock 数据时，从上到下依次执行，返回匹配到的第一个结果 | -- |

此外，不同的 `uri` 和 `method` 组合，可以生成不同的 Mock 接口。当配置中出现多个相同的`uri` 和 `method` 组合时，只有数组下标最小的有效。

`policies` 是 MocKuma 的核心映射策略，目前支持两种命令，即 `when` 和 `returns`：
- `when` 类似程序语言中的 `if`。`when` 中为限定策略的条件，可以有多种条件限定（暂时只支持 `params`）。不填写或者填写空 Json 对象，则该 `Policy` 恒真。
一个 `when` 中出现多个条件时，所有条件取逻辑“与”操作。当 `when` 中约束的条件满足时，即匹配成功，此时会执行 `returns` 命令。
`when` 中的限定条件均为选填，目前有如下限定条件：

| **条件** | **说明** | **示例** |
|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|
| params | 匹配请求中的 Url 参数，形如 `/uri?key=value`；<br>或是匹配 POST、PUT、DELETE 且`Content-Type` 为 `application/x-www-form-urlencoded` 的参数；<br> 形式为 Json 对象，其中 `key` 为参数名称，`value` 为参数值；<br>如果需要匹配多个同名参数，`value` 须为 Json 数组| `"params": {"value1": [1, 2], "value2": 2}` |


- `returns` 指定了 `when` 匹配后的返回内容，`returns` 中有如下参数：

| **参数** | **说明** | **示例** |
|------------|--------------------------------|----------------------------------------------------|
| statusCode | （选填，默认 200）Http 状态码 | `503` |
| headers | （选填）Http 响应头 | `"Content-Type": "text/html"` |
| body | （选填，默认为 ""）Http 响应体 | `"{\"code\": 2000, \"message\": \"Hello, World!\"}"` |