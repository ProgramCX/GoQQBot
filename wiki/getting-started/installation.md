# 安装 SDK

在 Go 项目里添加 GoQQBot 模块，就能导入 SDK。仓库的 `go.mod` 要求 Go **1.24.4** 或更高版本。

## 在业务项目中安装

先在业务目录初始化模块。

```sh
go mod init example.com/mybot
```

仓库和所需的提交都公开后，就可以添加依赖。

```sh
go get github.com/ProgramCX/GoQQBot
```

项目可以将 SDK 依赖固定到指定版本标签或提交。

## 使用本地工作区

要在业务项目里试用本地代码，可以添加 `replace`。把下面 PowerShell 示例中的目录换成你本机的 SDK 路径。

```powershell
go mod edit -require=github.com/ProgramCX/GoQQBot@v0.0.0
go mod edit -replace=github.com/ProgramCX/GoQQBot=E:/Document/GoQQBot
```

将[快速开始](quick-start.md)的代码保存为业务项目的 `main.go`，再运行。

```sh
go mod tidy
go run .
```

## 查看 API

安装后，可以用以下命令查看 SDK 的包和公开 API。

```sh
go doc github.com/ProgramCX/GoQQBot/pkg/goqqrobot
go doc github.com/ProgramCX/GoQQBot/shared/types.Context.Reply
```

普通编译无需 C 编译器。

## 注意事项

- SDK 是供业务项目导入的模块。运行 `go install` 不会得到一个可执行的机器人。
- 添加依赖时，Go 只会下载远端已有的代码。如需使用尚未推送的修改，可以配置本地工作区。示例未固定到特定发布版本。
- 本地 `replace` 只对当前模块生效。业务项目部署到其他机器前，把指向你本机的绝对路径换成可下载的版本。
- 在 Windows 等平台运行 `go test -race ./...`，需要满足 Go 的 CGO 和 C 工具链要求。

## 另请参阅

[快速开始](quick-start.md) · [配置文件](configuration.md) · [官方启动接入：机器人账号及接入票据](https://bot.q.qq.com/wiki/develop/api-v2/)
