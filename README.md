# MocKuma
See in other languages: English | [中文](README_CN.md) 

This is a Http API mocking server written in Go. It reads command-like json mapping configuration file, generating
corresponding mock API interfaces dynamically.

Front/back end developers may use this tool to mock RESTful API interfaces, helping developments and unit testings;
Tester may also use this tool with its command-like mapping configuration, writing mock APIs to match the parameters
and testing your software with your own test cases.



## Build & Run
Run `go get` or download source codes to `$GOPATH/github.com/kumasuke120/mockuma`, entering the directory and run the
following command:
```
$ cd cmd && go build -o ../bin/mockuma
```

You may click [here](https://github.com/kumasuke120/mockuma/releases) to download a executable of the newest
release version, if you don't own the Go development environment or you wanna do it quickly.

If you have got the executable, you could run the following command:
```
$ cd bin && ./mockuma
```


### Command Line Arguments
Although you could run MocKuma directly, MocKuma provides a series of command line arguments:

1. `-mapfile`: the path to the `MockuMappings` mapping configuration file, supports both relative and absolute path.
Specifically, the working directory of MocKuma will be set to the directory in which the `mapfile` resides.
Under the default circumstance, MocKuma will find a configuration file called `mockuMappings.json` or 
`mockuMappings.main.json` in the starting directory, reading and loading the file;
2. `-p`: the port number on which the MocKuma listens, the default value is `3214`;
3. `--help`: views the help content of command line arguments;
4. `--version`: views the version information of MocKuma;



## `MockuMappings` Essentials
`MockuMappings` is the unified name of MocKuma configuration files whose file formats are all of `.json`.


### Example Configuration Files
This repo provides example configuration files which help your understanding, resides in the `example/` directory,
you could view these files by clicking [here](example).
The files in the `example/single-file` directory are used in single-file mode, however those in the `example/multi-file` 
directory are used in multi-file mode.

It is recommended that run the following commands in the `$GOPATH/github.com/kumasuke120/mockuma` directory, starting
a MocKuma instance using example configuration files.

(Single-file Mode)
```
$ bin/mockuma -mapfile=example/single-file/mockuMappings.json
```
(Multi-file Mode)
```
$ bin/mockuma -mapfile=example/multi-file/mockuMappings.main.json
```


### Single-file & Multi-file Configuration
MocKuma supports both single-file and multi-file mode.

Single-file mode is suitable for the scenarios when the number of mock APIs is small and the logic of mock APIs is
not quite complex.

However, the multi-file mode needs a main entrance file. It is suitable for the scenarios when there are plenty of
APIs or complex logic.
You may put different API mappings into different files according your business in multi-file. Multiple files **don't**
have to be in the same directories. You could creating directories for different purposes to manage your configuration
files.

From `v1.1.0`, **multi-file mode is recommended**. In addition, the following parts are based on multi-file mode.
When it comes to the disparity between single-file and multi-file mode, there will be a specific explanation.


### Commenting Configuration Files
Because json does not have normal comment styles, `MockuMappings` provides a way to comment on your configuration file:
```json
{
  "@type": "...",
  "@comment": "Comment here..."
}
```
`@comment` in the above sample is the method to add comment. You could add comments in the **json object** 
(in the form of `{...}`) at any level, any position. The comment value could be of any type.

Besides, in the `MockuMappings`, all attributes like `@comment` which start with `@` are called directives. 
We will see them again in the following part.


### Paths in Configuration Files
Lots of parts of `MockuMappings` require include other files. The paths of those files support relative and absolute path.
It bears noting that the relative paths are relative to **the directory of main configuration file (specified by `-mapfile` in single-file mode)**.
No matter where the included files exists, this rule will always be complied.


### Main Configuration
Main configuration file is the default file in the multi-file mode. 
It is recommended to use `.main.json` as a suffix for this type of file.
You may click [here](example/multi-file/mockuMappings.main.json) to view the example file.

In the future release, the main configuration file will add some global configurations that control the default behavior
when the API interfaces mapping.

_Main configuration are not supported in single-file mode_

The following are the basic structures of the main configuration:
```json
{
  "@type": "main",
  "@comment": "This is main configuration",
  "@include": {
    "mappings": [
      "hello.mappings.json"
    ]
  }
}
```
The top level of a main configuration is a json object with following attributes:
- `@type`: type flag of `MockuMappings` files. when using in the main configuration, its value must be `main`;
- `@include`: include directive, can be used in main configuration only. its value should be a json object. 
The key of the json object is the `@type` of the included file; the value of the json object is a json array whose values
are the paths to the included files.
If the included file does not match the type specified by the object key, MocKuma will report an error. 
Currently this directive only supports the introduction of mappings files.


### Mappings Configuration
Mappings configuration file specifies API mapping related parameters.
It is recommended to use `.mappings.json` as a suffix for this type of file.
You may click [here](example/multi-file/mappings/hello.mappings.json) to view the example file.
The mappings configuration file mainly configures the uri, request method , and processing policies of 
the specific Mock APIs.

_In single-file mode, with some modifications this file is the main file. The differences will be explained later_

The following are the basic structures of the mappings configuration:
```json
{
  "@type": "mappings",
  "mappings": [
    {
      "@comment": {
        "uri": "when MocKuma starts locally and the port is 3214, you may access'http://localhost:3214/hello'",
        "method": "only GET method is mapped"
      },
      "uri": "/hello",
      "method": "GET",
      "policies": [
        {
          "when": {
            "params": {
              "@comment": "matches when the parameter is '/hello?lang=cn'",
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
              "@comment": "complex jsons can be written like this for easy viewing",
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
The top level of a mapping configuration is a json object with following attributes:
- `@type`: type flag of `MockuMappings` files. when using in the mappings configuration, its value must be `mappings`;
- `mappings`: a mapping configuration item set whose value is typically a json array. 
In particular, if there is only one mapping configuration item, you could put the item as the direct child of `mappings`.
Matching starts from first item and stops at the first matched item, that is, the smaller the array index, 
the higher the mapping item priority.

_In single-file mode, the contents of `mappings` (must be json array, no omission) as the direct top level _

#### Mapping Items in Mappings Configuration
There are currently three attributes for configuration in the mapping item: `uri`, `method`, `policies`:

- `uri`: the uri of the mock API, must start with `/`, which is required;
- `method`: the mock API mapped request method, supports all [Http/1.1 request methods](https://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html).
This attribute is optional. If you do not fill it in or fill it in `@any`, all request methods will be mapped.
- `policies`: a mapping policy item set whose value is typically a json array. 
In particular, if there is only one mapping policy item, you could put the item as the direct child of `policies`.
When processing a policy item, it matches **from top to bottom**, and returns the first result that matches.

It bears noting that there is no **repeatability check** on `uri` and `method`. 
When there are multiple pairs of `uri` and `method`, the matching is performed according to the priority mentioned above.

#### Policies in Mappings Configuration
There are currently two major attributes for configuration in the mapping policy item: `when`, `returns`:

- `when` is similar to `if` in the programming language. `when` defines a variety of conditions to limits policies. 
If you do not fill it in or fill it in an empty json object, then the mapping policy item is **always true**.
If there are multiple conditions in a single `when`, all conditions take a logical "** and **" operation. 
When the conditions in `when` is satisfied, the match is successful and the `returns` command is executed.
The conditions in `when` are optional. Currently we have the following conditions:

| **Condition** | **Description** | **Example** |
|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|
| `params` | (Optional) matches the url parameters in the request, in the form of `/uri?key=value`; <br>or matches POST, PUT, DELETE method with `Content-Type` of `application/x-www-form-urlencoded`; <br> it is of form json object, where key is the parameter name and value is the parameter value; <br> when you need to match multiple parameters with the same name, the value must be a json array | `"params": {"value1": [1, 2], "value2": 2}` |
| `headers` | (Optional) matches the parameters in the request headers in the same form as `params`, which also supports one or more parameter values. | `"headers": { "Authorization": "Basic a3VtYXN1a2UxMjAvcGEkJHcwcmQ=" }` |

- `returns` specifies the return value if the `when` matches, and the `returns` has the following parameters:

| **Condition** | **Description** | **Example** |
|------------|--------------------------------|----------------------------------------------------|
| `statusCode` | (Optional, default 200) Http status code | `503` |
| `headers` | (Optional) Http response header | `"Content-Type": "text/html"` |
| `body` | (Optional, default "") Http response body, either a string or an expanded json object or array | `"{\"code\": 2000, \"message\": \"Hello, World!\"}"` |

In addition, the `@file` file directive is supported in `body`. 
This directive specifies a file path (the relative path is relative to the directory where `mapfile` is located) 
and reads its contents as the value of this parameter.
It is recommended to use this command when the **return body is large**, which can make the configuration file more concise.


### Sample Responses of Example Configurations
You could use the default configuration and the multi-file example configuration above, start MocKuma locally, 
take a Http tool to request and log the results as follows
(there may be some differences with your responses, such as time, MocKuma version, etc.):

- Request `POST http://localhost:3214/api/hello?lang=cn&lang=en`, it returns:
```
HTTP/1.1 200 OK
Server: HelloMock/1.0
Date: Sun, 17 Nov 2019 18:08:34 GMT
Content-Length: 43
Content-Type: text/plain; charset=utf-8

{"code": 2000, "message": "Hello, 世界!"}
```

- Request `GET http://localhost:3214/api/books?page=2&perPage=20`, it returns:
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf8
Server: MocKuma/1.1.0
Date: Sun, 17 Nov 2019 18:09:52 GMT
Content-Length: 531

<The conetent of 'books-page2.json'>
```

- Request `DELETE http://localhost:3214/api/notexists`, it returns:
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

- Request `GET http://localhost:3214/whoami`, it returns:
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



## `MockuMappings` Template Engine
From `v1.1.0`, MocKuma adds a template engine. 
When you declare a template, you could use placeholders to reference variables. When applying this kind of template, 
MocKuma will generate dynamically using given variable sets.

### Template Declaration Configurations
Template declaration configuration file declares and defines a new template.
It is recommended to use `.template.json` as a suffix for this type of file.
You may click [here](example/multi-file/template/login-policy.template.json) to view the example file.

The following are the basic structures of the template declaration configuration:
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
        "@comment": "'@{}' is the placeholder for @vars",
        "code": 2000,
        "message": "Welcome, @{username}!"
      }
    }
  }
}
```
The top level of a template declaration configuration is a json object with following attributes:
- `@type`: type flag of `MockuMappings` files. when using in the template declaration configuration, its value must be `template`;
- `template`: the content for the actual template,  supports strings, json objects, json arrays.

#### Placeholders in Template Declaration
In template declaration, you could reference variables at any level, any position with the placeholder `@{varName}`.
The `varName` in the placeholder is variable name.
In particular, if you want to display `@{varName}` directly in the template without replacing it,
you can ** double write ** `@` to escape (`@@{varName}`).


### Vars Definition Configurations
Vars definition configuration file defines a series of variables and their actual values..
It is recommended to use `.vars.json` as a suffix for this type of file.
You may click [here](example/multi-file/vars/login-policy.vars.json) to view the example file.

The following are the basic structures of the vars definition configuration:
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
The top level of a vars definition configuration is a json object with following attributes:
- `@type`: type flag of `MockuMappings` files. when using in the vars definition configuration, its value must be `vars`;
- `vars`: a var definition item set whose value is typically a json array. 
In particular, if there is only one mapping policy item, you could put the item as the direct child of `vars`.
Each var definition item is a json object, the key of the object is the variable name, 
the variable name must **begin with a letter, following with the letters and numbers**, that is, 
the variable name must matches the regular expression `/[az][ Az\d]*/i`.

### Template Applying
Once you have declared the template and defined the variables, 
you can apply the template. Here is an example of an application template:
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
When applying a template, you need to use the `@template` template application directive in a separate json object, 
you can't use the template application directive **at the top level**.
When applying a template, you must use the `vars` attribute or the `@vars` directive to specify the variable 
corresponding to the template. The specific configuration of the `vars` attribute is consistent with the corresponding 
attribute in the vars definition configuration.

After applying the template, the placeholders in the template will be replaced with variable values 
according to certain rules and a corresponding json structure will be generated.
After generating the json structure, the json structure will be placed in the location **where the `@template`
template** applying directive is located.

#### Placements of Templates Rendering Results
Since we could define multiple groups of variables, and we could apply template te at almost any position, there
are some special rules MocKuma will follow when rendering templates.

The rendering of template is related to the number of variable groups you specify when applying the template. 
The template renders as many times as the number of groups of variables.

- When applying templates in the json array, no matter how many groups of variables there are, 
the result of template rendering **will be directly inserted into the corresponding position** of `@template`
template applying directive;
- When a template is applied in a json object, if there is only one group of variables, 
the rendering results will be placed as where the template places; if there are multiple groups, 
all the rendering results will be **put into a json array** and then placed in the corresponding location.


### Sample Usages of Template Engine
Given the following template declaration, whose path is `template/alphabet-order.template.json`:
```json
{
  "@type": "template",
  "template": {
    "@comment": "the upper and lowe cases of alphabet and its order",
    "alphabet": "@{alphabetUpper}|@{alphabetLower}",
    "order": "@{order}"
  }
}
```

- if applying the template as follows:
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
will be rendered as:
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

- if applying the template as follows:
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
      }
    ]
  }
}
```
will be rendered as:
```json
{
  "@type": "mappings",
  "mappings": {
    "alphabet": "A|a",
    "order": 0
  }
}
```
