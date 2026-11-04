package api

import "errors"

const (
	codeAPIAccessDenied       = 40012010
	codeLegacyAPIAccessDenied = 11253
	codeOperationDenied       = 40062003
	codeRecallWindowExpired   = 40064004
	codeChannelRecallDenied   = 306004
	codeInvalidParameters     = 40061001
	codeInvalidMessageID      = 40061002
	codeRecallFailed          = 50065001
	codeChannelRecallInvalid  = 306001
	codeChannelRecallBadID    = 306002
	codeChannelMessageLookup  = 306003
	codeChannelRecallFailed   = 306005
	codeRecallChannelLookup   = 306006
)

var (
	// ErrAPIAccessDenied 表示机器人应用没有接口访问权限。
	ErrAPIAccessDenied = errors.New("QQ application cannot access this API")
	// ErrOperationDenied 表示 QQ 拒绝对指定对象执行操作。
	ErrOperationDenied = errors.New("QQ operation permission denied")
	// ErrRecallWindowExpired 表示已超过 QQ 允许的消息撤回时限。
	ErrRecallWindowExpired = errors.New("QQ message recall window expired")
	// ErrInvalidAPIParameters 表示 QQ 拒绝了请求参数。
	ErrInvalidAPIParameters = errors.New("QQ request parameters are invalid")
	// ErrInvalidMessageID 表示 QQ 拒绝了请求中的消息 ID。
	ErrInvalidMessageID = errors.New("QQ message ID is invalid")
	// ErrRecallMessageLookup 表示 QQ 在撤回时未能取得消息。
	ErrRecallMessageLookup = errors.New("QQ failed to look up the message for recall")
	// ErrRecallChannelLookup 表示 QQ 在撤回时未能取得子频道。
	ErrRecallChannelLookup = errors.New("QQ failed to look up the channel for recall")
	// ErrRecallFailed 表示 QQ 返回撤回失败，调用方可核对状态后决定是否重试。
	ErrRecallFailed = errors.New("QQ message recall failed")
)

// Is 按平台错误码匹配稳定错误，原始响应仍可用 errors.As 取得。
func (e *APIError) Is(target error) bool {
	if e == nil {
		return false
	}
	switch e.Code {
	case codeAPIAccessDenied, codeLegacyAPIAccessDenied:
		return target == ErrAPIAccessDenied
	case codeOperationDenied, codeChannelRecallDenied:
		return target == ErrOperationDenied
	case codeRecallWindowExpired:
		return target == ErrRecallWindowExpired
	case codeInvalidParameters, codeChannelRecallInvalid:
		return target == ErrInvalidAPIParameters
	case codeInvalidMessageID, codeChannelRecallBadID:
		return target == ErrInvalidMessageID
	case codeChannelMessageLookup:
		return target == ErrRecallMessageLookup
	case codeRecallChannelLookup:
		return target == ErrRecallChannelLookup
	case codeRecallFailed, codeChannelRecallFailed:
		return target == ErrRecallFailed
	default:
		return false
	}
}
