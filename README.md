# MocKuma [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Release](https://img.shields.io/github/release/kumasuke120/mockuma/all.svg)](https://github.com/kumasuke120/mockuma/releases/latest) [![Build Status](https://api.travis-ci.org/kumasuke120/mockuma.svg?branch=dev)](https://travis-ci.org/kumasuke120/mockuma) [![codecov](https://codecov.io/gh/kumasuke120/mockuma/branch/dev/graph/badge.svg)](https://codecov.io/gh/kumasuke120/mockuma)

[English | [中文](README_CN.md)]

MocKuma is an http mocking server. It reads command-like json mapping configuration file, generating
corresponding mock APIs dynamically.

Front/back end developers may use this tool to mock RESTful APIs, helping developments and unit testings;
Tester may also use this tool with its command-like mapping configuration, writing mock APIs to match the parameters
and testing it.

### Features
- maps responses based on requests's parameters/headers
- reloads automatically mappings when changed
- renders multiple mappings with user-defined templates and variables 
- supports references static files
- supports redirects and forwards


## Installation
Executes the following command to install in your environment:
```
$ go get -u github.com/kumasuke120/mockuma/cmd/mockuma
```

You may also click [here](https://github.com/kumasuke120/mockuma/releases) to download an executable file of the latest
release version if you don't own the Go development environment or if you wanna do it quickly.


## Quick Start

1. Makes sure `$GOPATH\bin` has been included to your environment variable `$PATH`;
2. Creates a new file called `mockuMappings.json` with its content as below:
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
3. Starts the MocKuma with the following command:
```
$ mockuma
```
4. Then you could access [http://localhost:3214/](http://localhost:3214/) or 
[http://localhost:3214/?lang=cn](http://localhost:3214/?lang=cn) to check out the result. 

#### Command-Line Arguments
Although you could run MocKuma directly like the example above, MocKuma provides a series of command line arguments:

1. `-mapfile=<filename>`: the path to the MockuMappings mapping configuration file, supports both relative and absolute path. 
Under the default circumstance, MocKuma will find a configuration file called `mockuMappings.json`, 
`mockuMappings.main.json` or `main.json` in the current working directory, reading and loading the file.
Specifically, the working directory of MocKuma will be set to the directory in which the mapfile resides if you specify it manually;
2. `-p=<port_number>`: the port number on which the MocKuma listens, the default value is 3214;
3. `--version`: views the version information of MocKuma.

#### More Examples
You could click [here](example) to see more examples.
