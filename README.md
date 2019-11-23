# MocKuma
这是一款使用 Go 编写的 Http 接口 Mock 工具。该工具读取命令化的 Json 映射配置文件，并根据配置生成对应的 Mock 接口。

前、后端开发人员使用本工具可以模拟 Restful 接口以辅助开发以及单元测试；测试人员也可以使用本工具利用其命令式的参数匹配编写符合测试用例的接口辅助测试。


## 构建与运行
使用 `go get` 命令或下载压缩包解压至 `$GOPATH/github.com/kumasuke120/mockuma` 中，进入该目录并执行以下代码进行构建：
```
$ cd cmd && go build -o ../bin/mockuma
```

如果想要避免麻烦或者没有 Go 的开发环境，请[点此](https://github.com/kumasuke120/mockuma/releases)以下载已发布版本的可执行文件。

构建或下载成功并重命名后，可以使用以下命令运行：
```
$ cd bin && ./mockuma
```

### 命令行参数
虽然 MocKuma 可以直接执行，但是它也提供了一些参数供配置使用，以下是所有支持的命令行参数：

1. `-mapfile`: `MockuMappings` 映射配置文件路径，支持相对路径和绝对路径。特别的，MocKuma 的工作目录将会被设为该配置文件所在目录。
默认情况下，将会依次寻找当前目录下名为 `mockuMappings.json`，`mockuMappings.main.json` 的配置并读取；
2. `-p`: MocKuma 监听端口号，默认值是 `3214`；
3. `--help`: 查看帮助，帮助内容文本为英文；
4. `--version`: 查看当前 MockKuma 的版本信息。


## `MockuMappings` 基础
`MockuMappings` 是 MocKuma 的配置文件统称，其文件内容均为 `.json` 格式。

### 示例配置文件
为了便于理解，项目中提供了示例配置文件，位于 `example/` 文件夹中，可以[点此](example)查看这些配置文件。
其中位于 `example/single-file` 中的是单文件配置，而位于 `example/multi-file` 中的则是多文件配置。


为了便于充分理解以下部分的说明和示例返回，建议在 `$GOPATH/github.com/kumasuke120/mockuma` 目录下执行以下命令，使用示例配置文件启动 MocKuma：

(单文件模式)
```
$ bin/mockuma -mapfile=example/single-file/mockuMappings.json
```
(多文件模式)
```
$ bin/mockuma -mapfile=example/multi-file/mockuMappings.main.json
```

### 单文件配置与多文件配置
MocKuma 支持单文件和多文件两种配置模式。

其中单文件模式适合 Mock 接口数量少的使用场景，配置简单，可以快速成型。

而多文件模式则需要一个入口文件，适合 Mock 接口多，业务情况复杂的情况。
使用多文件模式可以将不同业务的接口 Mock 放在不同的文件中，多个文件毋需在统一目录下，可以建文件夹进行管理。

`1.1.0` 版本后，推荐使用多文件模式进行配置，以下的说明也以多文件模式为基准，涉及到单文件与多文件模式不同的部分将会特殊说明。

### 注释配置文件
由于 Json 格式不支持一般的注释语法，`MockuMappings` 中提供了一种添加注释的方式：
```json
{
  "@type": "...",
  "@comment": "这里填写注释"
}
```
上例中的 `@comment` 就是添加注释的方式，可以添加在任意层级、任意位置的 **Json 对象**（形如 `{...}`）中，其内容可以为任意类型。
MocKuma 在读取配置文件时，将自动去除 `@comment` 相关内容，不会对配置造成影响。
此外，形如 `@comment` 以 `@` 开头的属性在 `MockuMappings` 被称为指令(directive)，接下来我们还将会看到它们。

### 配置文件中的路径
在 `MockuMappings` 中很多地方支持引用其他文件，文件路径支持相对和绝对路径。
其中需要注意的是相对路径是**相对于主入口文件（单文件模式下为 `-mapfile` 指定文件）所在目录的**，无论引用其他文件的文件所在位置如何，都遵守这个规律。

### 主入口(main)配置
主入口(main)配置文件为多文件模式下的默认文件，推荐该类型文件使用 `.main.json` 作为后缀。
可以[点此](example/multi-file/mockuMappings.main.json)查看示例文件。

在未来的版本中，主入口文件将会增加全局的配置，以用于控制接口映射处理时的默认行为。

_在单文件模式下，不支持使用主入口文件_

以下是主入口配置的基本样式：
```json
{
  "@type": "main",
  "@comment": "这是主入口文件",
  "@include": {
    "mappings": [
      "hello.mappings.json"
    ]
  }
}
```
主入口配置顶层为 Json 对象，有以下属性 key：
- `@type`: `MockuMappings` 配置文件类型的标志，用于主入口配置时，其值须为 `main`；
- `@include`: 引入指令(directive)，仅可在主入口文件中使用。其值须为 Json 对象，对象的 key 为引入文件的类型(`@type`)，对象的 key 对应的 value
是一个 Json 数组，其中每一个值均为所引入文件的路径。如果被引用的文件和对象 key 指定的类型不一致，MocKuma 将会报错。目前该指令仅支持引入映射(mappings)文件。

### 映射(mappings)配置
映射(mappings)配置文件指定接口映射相关参数，推荐该类型文件使用 `.mappings.json` 作为后缀。
可以[点此](example/multi-file/mappings/hello.mappings.json)查看示例文件。
映射配置文件主要配置具体 Mock 接口的地址(uri)、请求方式(method)以及处理策略(policies)。

_在单文件模式下，该文件即为入口文件，但有些变化，将在下面说明_

以下是映射配置的基本样式：
```json
{
  "@type": "mappings",
  "mappings": [
    {
      "@comment": {
        "uri": "本地启动且端口号为 3214 时，访问 'http://localhost:3214/hello' 即可",
        "method": "仅 GET 方式可调用"
      },
      "uri": "/hello",
      "method": "GET",
      "policies": [
        {
          "when": {
            "params": {
              "@comment": "调用参数为 '/hello?lang=cn' 时匹配",
              "lang": "cn"
            }
          },
          "returns": {
            "headers": {
              "Content-Type": "application/json; charset=utf8"
            },
            "body": "{\"code\": 2000, \"message\": \"你好，世界！\"}"
          }
        },
        {
          "returns": {
            "headers": {
              "Content-Type": "application/json; charset=utf8"
            },
            "body": {
              "@comment": "复杂的 json 结构，可以像这样展开书写，便于查看",
              "code": 2000,
              "message": "Hello, World!"
            }
          }
        }
      ]
    },
    {
      "uri": "/hello",
      "method": "@any",
      "policies": {
        "returns": {
          "statusCode": 405
        }
      }
    }
  ]
}
```
映射配置顶层为 Json 对象，有以下属性 key：
- `@type`: `MockuMappings` 配置文件类型的标志，用于映射配置时，其值须为 `mappings`；
- `mappings`: 映射项配置集，其值一般为 Json 数组。特别的，如果只有一个映射项，可以不使用 Json 数组将映射项作为 `mappings` 的直接下级。
匹配时，从第一个配置项开始从上到下依次匹配，匹配到第一个符合映射项即停止，即数组下标越小，映射项优先级越高。

_在单文件模式下，须直接将 `mappings` 中的内容（必须是 Json 数组，无省略写法）作为顶层_

#### 映射(mappings)的映射项
目前映射项中有三个属性配置：`uri`, `method`, `policies`：

- `uri`: Mock 接口的 uri，必须以 `/` 开头，该参数必填；
- `method`: Mock 接口映射的请求方式，支持所有 [Http/1.1 的请求方式](https://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html)。
该参数非必填，没有填写或者填写 `@any` 时，将会映射所有的请求方式；
- `policies`: Mock 接口的处理策略配置集，其值一般为 Json 数组。特别的，如果只有一个处理策略项，可以不使用 Json 数组将处理策略项作为 `policies` 的直接下级。
执行处理策略项(policy)时，从上到下依次匹配，返回匹配到的第一个结果。

需要注意的是，`uri` 和 `method` 上没有重复性检查，当出现多个配置时，按照上文提到的优先级进行匹配处理。

#### 映射项的处理策略项(policy)
目前处理策略项中有两大属性配置：`when`, `returns`：

- `when` 类似程序语言中的 `if`。`when` 中为限定策略的条件，可以有多种条件限定。不填写或者填写空 Json 对象，则该处理策略项(policy)恒真。
一个 `when` 中出现多个条件时，所有条件取逻辑“与”操作。当 `when` 中约束的条件满足时，即匹配成功，此时会执行 `returns` 命令。
`when` 中的限定条件均为选填，目前有如下限定条件：

| **条件** | **说明** | **示例** |
|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|
| `params` | （选填）匹配请求中的 Url 参数，形如 `/uri?key=value`；<br>或是匹配 POST、PUT、DELETE 且`Content-Type` 为 `application/x-www-form-urlencoded` 的参数；<br> 其形式为 Json 对象，其中 `key` 为参数名称，`value` 为参数值；<br>需要匹配多个同名参数时，`value` 须为 Json 数组| `"params": {"value1": [1, 2], "value2": 2}` |
| `headers` | （选填）匹配请求头中的参数，其形式和 `params` 相同，同样支持一个或多个参数值 | `"headers": { "Authorization": "Basic a3VtYXN1a2UxMjAvcGEkJHcwcmQ=" }` |

- `returns` 指定了 `when` 匹配后的返回内容，`returns` 中有如下参数：

| **参数** | **说明** | **示例** |
|------------|--------------------------------|----------------------------------------------------|
| `statusCode` | （选填，默认 200）Http 状态码 | `503` |
| `headers` | （选填）Http 响应头 | `"Content-Type": "text/html"` |
| `body` | （选填，默认为 ""）Http 响应体，可以为字符串，也可以是展开的 Json 对象或数组 | `"{\"code\": 2000, \"message\": \"Hello, World!\"}"` |

此外，`body` 中支持 `@file` 文件指令。该指令指定一个文件路径（相对路径是相对 `mapfile` 所在目录），并读取其内容作为该参数的值。
推荐在返回体非常大时使用该指令，可以使得配置文件更加简洁。

### 示例配置返回展示
使用默认配置以及上述的多文件示例配置，在本地启动 MocKuma，使用 Http 工具请求并记录运行结果如下（与具体结果可能有细节上的差异，如时间，MocKuma 版本等）：

- 请求 `POST http://localhost:3214/api/hello?lang=cn&lang=en`，返回：
```
HTTP/1.1 200 OK
Server: HelloMock/1.0
Date: Sun, 17 Nov 2019 18:08:34 GMT
Content-Length: 43
Content-Type: text/plain; charset=utf-8

{"code": 2000, "message": "Hello, 世界!"}
```

- 请求 `GET http://localhost:3214/api/books?page=2&perPage=20`，返回：
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf8
Server: MocKuma/1.1.0
Date: Sun, 17 Nov 2019 18:09:52 GMT
Content-Length: 531

<文件 'books-page2.json' 的内容>
```

- 请求 `DELETE http://localhost:3214/api/notexists`，返回：
```
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf8
Server: MocKuma/1.1.0
Date: Sun, 17 Nov 2019 18:11:42 GMT
Content-Length: 43

{
  "statusCode": 404,
  "message": "Not Found"
}
```

- 请求 `GET http://localhost:3214/whoami`，返回：
```
HTTP/1.1 200 OK
Server: MocKuma/1.1.0
Date: Sun, 17 Nov 2019 14:36:17 GMT
Content-Length: 36
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html lang="en">
<head>
	<title>Whoami</title>
</head>
<body>
<h1>I am MocKuma</h1>
</body>
</html>
```

## `MockuMappings` 模板
`1.1.0` 版本起，MocKuma 添加了模板系统，模板声明时可以使用占位符(placeholder)引用变量，这样在应用模板时可以套用变量值动态生成。

### 模板声明(template)配置
模板声明(template)配置文件声明并定义一个新的模板，推荐该类型文件使用 `.template.json` 作为后缀。
可以[点此](example/multi-file/template/login-policy.template.json)查看示例文件。

以下是模板声明配置的基本样式：
```json
{
  "@type": "template",
  "template": {
    "when": {
      "params": {
        "username": "@{username}"
      },
      "headers": {
        "Authorization": "Basic @{authToken}"
      }
    },
    "returns": {
      "headers": {
        "Content-Type": "application/json; charset=utf8"
      },
      "body": {
        "@comment": "'@{}' 是 @vars 的占位符",
        "code": 2000,
        "message": "Welcome, @{username}!"
      }
    }
  }
}
```
模板声明配置顶层为 Json 对象，有以下属性 key：
- `@type`: `MockuMappings` 配置文件类型的标志，用于模板声明配置时，其值须为 `template`；
- `template`: 模板具体内容。支持字符串、Json 对象、Json 数组。

#### 模板声明中的占位符
在模板声明中，可以在任意位置、任意层级使用占位符 `@{varName}` 引用变量，其中 `varName` 为变量名。

如果占位符是字符串的一部分，渲染模板时，对应引用的变量值将被转换为字符串拼接到原字符串上。
如果整个字符串中仅有一个占位符，渲染模板时，对应引用的变量值将直接替换到占位符的位置，变量类型不变。

### 变量定义(vars)配置
变量定义(vars)配置文件定义了一系列变量和它们具体的值，推荐该类型文件使用 `.vars.json` 作为后缀。
可以[点此](example/multi-file/vars/login-policy.vars.json)查看示例文件。

以下是变量定义配置的基本样式：
```json
{
  "@type": "vars",
  "vars": [
    {
      "username": "kumasuke120",
      "authToken": "a3VtYXN1a2UxMjAvcGEkJHcwcmQ="
    },
    {
      "username": "jane.doe",
      "authToken": "amFuZS5kb2Vqb25lLmRvZQ=="
    },
    {
      "username": "jone.doe",
      "authToken": "am9uZS5kb2VqYW5lLmRvZQ=="
    }
  ]
}
```
变量定义配置顶层为 Json 对象，有以下属性 key：
- `@type`: `MockuMappings` 配置文件类型的标志，用于变量定义配置时，其值须为 `vars`；
- `vars`: 变量定义项集，其值一般为 Json 数组。特别的，如果只有一个变量定义项，可以不使用 Json 数组将变量定义项作为 `vars` 的直接下级。
每个变量定义项都是一个 Json 对象，对象的 key 是变量名，变量名须以字母开头，之后可以接字母与数字，即变量名须符合正则表达式 `/[a-z][a-z\d]*/i`。

### 模板(template)应用
声明模板并定义好变量后，就可以应用模板了，以下是应用模板的示例：
```json
{
  "@type": "mappings",
  "mappings": [
    {
      "@template": "template/hello.template.json",
      "@vars": "vars/hello.vars.json"
    },
    {
      "@template": "template/hello.template.json",
      "vars": [
        {
          "var1": "val1",
          "var2": "val2"
        }
      ]
    }
  ]
}
```
应用模板时，需要在一个单独的 Json 对象中使用 `@template` 模板应用指令，不能在最顶层使用模板应用指令。
应用模板时必须使用 `vars` 属性或 `@vars` 指令指定模板对应的变量。`vars` 属性的具体配置同变量定义(vars)配置中对应属性一致。

应用模板后，模板中的占位符将会被按照一定规则以变量值替换并生成对应 Json 结构。
生成 Json 结构之后，Json 结构将会被放置在 `@template` 模板应用指令所在的位置。

#### 模板渲染结果的放置方式
由于变量定义时可以定义多组，且可以在不同的位置应用模板，模板渲染时有一些特殊的规则。

模板渲染的样式和其应用模板时指定的变量组数有关，变量有多少组，模板就会渲染多少次。

- 当在 Json 数组中应用模板时，无论有多少组变量，模板渲染的结果将直接插入 `@template` 模板应用指令所在的对应位置。
- 当在 Json 对象中应用模板时，如果变量只有一组，渲染结果将会以模板原样放置；如果存在多组，所有渲染结果将会被放入一个 Json 数组中再放置到对应位置。

#### 模板渲染结果示例
假设有如下模板声明，其文件路径为 `template/alphabet-order.template.json`：
```json
{
  "@type": "template",
  "template": {
    "@comment": "字母大小写以及其序号",
    "alphabet": "@{alphabetUpper}|@{alphabetLower}",
    "order": "@{order}"
  }
}
```

- 如果应用模板方式如下所示：
```json
{
  "@type": "mappings",
  "mappings": {
    "@template": "template/alphabet-order.template.json",
    "vars": [
      {
        "alphabetUpper": "A",
        "alphabetLower": "a",
        "order": 0
      },
      {
        "alphabetUpper": "B",
        "alphabetLower": "b",
        "order": 1
      }
    ]
  }
}
```
则将被渲染成：
```json
{
  "@type": "mappings",
  "mappings": [
    {
      "alphabet": "A|a",
      "order": 0
    },
    {
      "alphabet": "B|b",
      "order": 1
    }
  ]
}
```