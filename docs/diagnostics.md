# 日志、错误通知和 panic

SDK 使用标准库 `log/slog`。默认静默，不创建日志文件、不写标准输出、不改变应用的 slog.Default。每个 Bot 持有自己的 logger；API Client 和配置加载器只返回错误，不重复记录消息载荷。

## 注入日志与控制级别

```go
var level slog.LevelVar
level.Set(slog.LevelInfo)
appLog := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: &level,
}))
bot, err := goqqrobot.NewBot("config.yaml", nil,
    goqqrobot.WithLogger(appLog),
)
if err != nil { return err }
// 运行期间可安全调整：
level.Set(slog.LevelDebug)
```

可替换为 JSONHandler 或符合 slog.Handler 的现有日志适配器。`WithLogger(nil)` 明确关闭 SDK 日志；不传该选项的效果相同。SDK 不修改应用全局日志设置，调用方注入的文件、writer 和 handler 由调用方管理、关闭。

命令注册等过程信息为 Debug，实际错误为 Error。未知命令只记录事件类别，不记录用户输入的命令或参数。错误日志保留阶段、事件类型，以及可识别 API 错误的状态、错误码和 trace；不会格式化任意 handler error、panic Value、Stack、消息正文或整个 Context。旧频道接口的消息体 Debug 日志已移除。

APIError.Error 仍省略远端响应。需要平台说明时，由业务显式读取 DiagnosticMessage，按自己的隐私策略处理；`examples/media -debug-events` 是显式开启该诊断的示例，和 SDK 的日志级别分开。

## 实时接收错误和 panic

```go
var failures, panics atomic.Uint64
bot, err := goqqrobot.NewBot("config.yaml", nil,
    goqqrobot.WithLogger(appLog),
    goqqrobot.WithErrorHandler(func(ctx *types.Context, err error) {
        failures.Add(1)
        var recovered *types.PanicError
        if errors.As(err, &recovered) {
            panics.Add(1)
            // 可显式检查 recovered.Value、recovered.Stack。
            // 它们可能含业务敏感数据，不要直接写进公共日志。
        }
        // 业务可在这里记录指标、触发告警或按策略取消应用 context。
    }),
)
```

回调收到 `*types.HandlerError`，包含 EventType、Stage 和可解包的 Err。普通错误可继续用 `errors.Is` / `errors.As` 查找原始原因；panic 对应 `*types.PanicError`。如果 panic 的值本身是 error，其错误链也会保留。

handler 的 panic 在调用边界捕获，错误回调拿到同一个 Context，因此命令、参数、原始 Payload 和自动回复序号仍属于同一次事件。panic 不会杀掉 worker，后续事件继续处理，已接收任务的完成通知仍被执行。任务池还保留对调度层意外 panic 的兜底捕获，此时 Stage=task。

回调在当前 worker 内同步调用，不持有 SDK 注册锁或生命周期锁；多个 worker 可并发调用，业务状态需要自己同步。回调应及时返回，不要把整个 Payload、Context 或错误对象交给另一个 goroutine 并修改它们。应用需要长期保留数据时，明确复制需要的业务字段。

错误回调或错误日志 handler 自身 panic 时，也会被捕获并保留，分别使用 Stage=error_handler/logger；不会递归再次调用错误回调。

## 停止时不会丢失失败

处理过程继续运行，SDK 在固定大小的状态中保留：

- 第一个事件处理失败（含解析错误和 panic）；
- 第一个错误观察器失败（日志 handler 或错误回调 panic）。

正常结束时，Run、Start、StartWithGracefulShutdown 和等待完成的 Shutdown 会将已记录失败与传输/关闭错误合并返回。即使未配置 logger 和错误回调，失败仍可从结束结果观察，不再静默吞掉。它不是所有历史错误的列表；需要逐次通知、计数或外部存储时使用 WithErrorHandler。

```go
err = bot.Run(ctx)
var failure *types.HandlerError
if errors.As(err, &failure) {
    // 应用决定如何处理期间发生的事件失败。
    return err
}
if errors.Is(err, context.Canceled) {
    return nil
}
return err
```

关闭截止时间不能强行停止不响应取消的用户代码；超时后仍在运行的回调，无法保证其未来错误进入已经返回的结果。请保持 handler 和错误回调遵守 context 与关闭期限。

## 迁移说明

- 以前自动输出的注册日志现在默认关闭。应用需要输出时显式传 `WithLogger`，仓库示例已这样配置。
- 以前 handler 返回错误或 panic 后可能只写日志/静默继续；现在实时回调可观察，且停止结果可能为非 nil。即使错误回调已做业务处理，停止结果仍保留最早失败。请检查调用方只判断 context.Canceled 就忽略所有错误的逻辑。
- 未改变 handler 函数签名、OnCommand/OnMessage 注册方式及 SDK 外部依赖；无需自定义 SDK logger 接口，也不必导入 internal 包。

## Windows 竞态检查

本机已通过 WinGet 安装用户级 WinLibs GCC（POSIX/UCRT）。安装后重新打开 PowerShell，使持久化 PATH 生效：

```powershell
gcc --version
$env:CGO_ENABLED = "1"
go test -race -p 1 ./...
```

`-p 1` 降低本机并行构建的内存占用，不关闭竞态检测。资源充足时可去掉该参数。检查应与普通测试、go vet 一起执行；竞态检测只覆盖实际运行的路径，并不能证明所有执行路径均无数据竞争。[Go 官方说明](https://go.dev/doc/articles/race_detector)
