# 安装 SDK

GoQQBot 作为模块导入业务项目。所需 Go 版本见仓库的 `go.mod`，最低为 **1.24.4**。

## 在业务项目中安装

先在业务目录初始化模块。

```sh
go mod init example.com/mybot
```

仓库及目标提交公开后，可以添加依赖。

```sh
go get github.com/ProgramCX/GoQQBot
```

发布后，业务项目可固定已验收的标签或提交。

## 使用本地工作区

开发尚未发布的代码时，在业务项目中添加 `replace`。Windows PowerShell 示例中的目录可换成实际位置。

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

用本地命令查看包和公开 API。

```sh
go doc github.com/ProgramCX/GoQQBot/pkg/goqqrobot
go doc github.com/ProgramCX/GoQQBot/shared/types.Context.Reply
```

普通编译无需 C 编译器。

## 注意事项

- SDK 以模块形式使用。`go install` 不会安装一个可运行的机器人程序。
- 添加依赖时，Go 读取远端已发布的代码。本地尚未推送的修改不会自动下载。文档不预设已发布的版本。
- 本地 `replace` 只作用于当前模块，不会随 SDK 传递。发布到其他机器前，把本地绝对路径换成可下载的版本。
- 在 Windows 等平台运行 `go test -race ./...`，需要满足 Go 的 CGO 和 C 工具链要求。

## 另请参阅

[快速开始](quick-start.md) · [配置文件](configuration.md) · [官方启动接入：机器人账号及接入票据](https://bot.q.qq.com/wiki/develop/api-v2/)
