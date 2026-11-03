# 安装 SDK

需要 Go **1.24.4 或以上**。SDK 是供业务项目导入的模块，不是通过 `go install` 安装的机器人程序。

## 在业务项目中安装

在自己的业务目录初始化模块：

```sh
go mod init example.com/mybot
```

仓库及目标提交已经公开可下载时，可添加依赖：

```sh
go get github.com/ProgramCX/GoQQBot
```

这个命令取决于远端已发布的代码，不保证包含本地工作区尚未推送的修改。发布后应在业务项目中固定经过验收的标签或提交；本文不假设已有某个发布版本。

## 使用本地工作区

尚未发布，或者需要同时开发 SDK 和机器人业务时，在业务项目目录添加 `replace`。下面是 Windows PowerShell 示例，目录按实际位置修改：

```powershell
go mod edit -require=github.com/ProgramCX/GoQQBot@v0.0.0
go mod edit -replace=github.com/ProgramCX/GoQQBot=E:/Document/GoQQBot
```

将[快速开始](quick-start.md)的代码保存为业务项目的 `main.go`，然后运行：

```sh
go mod tidy
go run .
```

本地 `replace` 只影响当前模块，不会随 SDK 自动传递给其他项目。将业务发布到其他机器前，移除本地绝对路径并使用可下载的版本。

## 查看 API

```sh
go doc github.com/ProgramCX/GoQQBot/pkg/goqqrobot
go doc github.com/ProgramCX/GoQQBot/shared/types.Context.Reply
```

普通编译不要求 C 编译器。在 Windows 等平台运行 `go test -race ./...` 时，需要满足 Go 竞态检测的 CGO 与 C 工具链要求。

## 另请参阅

- [快速开始](quick-start.md)
- [配置文件](configuration.md)
- [官方启动接入：机器人账号及接入票据](https://bot.q.qq.com/wiki/develop/api-v2/)
