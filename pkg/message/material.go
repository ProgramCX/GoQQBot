package message

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ImagePrepareRequest registers a Markdown image upload. Size is a JSON number, not a decimal string.
// ImagePrepareRequest 登记 Markdown 图片上传，Size 是 JSON 数字而非十进制字符串。
// ContentType must match the file extension and the subsequent storage PUT header.
// ContentType 必须与扩展名及后续存储 PUT 请求头一致。
type ImagePrepareRequest struct {
	FileName    string `json:"file_name"`
	FileSize    uint64 `json:"file_size"`
	ContentType string `json:"content_type"`
	MD5         string `json:"md5"`
	SHA1        string `json:"sha1"`
}

// ImageUploadPlan contains temporary upload credentials. Do not log PresignedURL or persist it as an image reference.
// ImageUploadPlan 包含临时上传凭证，不得记录 PresignedURL 或将其持久化为图片引用。
type ImageUploadPlan struct {
	UploadID     string          `json:"upload_id"`
	PresignedURL string          `json:"presigned_url"`
	ExpiresIn    uint32          `json:"expires_in"`
	Raw          json.RawMessage `json:"-"`
}

// ImageCompleteRequest acknowledges an upload using the storage ETag without its surrounding quotes.
// ImageCompleteRequest 使用去掉外围引号的存储 ETag 确认上传。
type ImageCompleteRequest struct {
	UploadID string `json:"upload_id"`
	MD5      string `json:"md5"`
}

// ImageMaterial is a bot-owned reusable Markdown image, independent of any conversation.
// ImageMaterial 是机器人持有且可复用的 Markdown 图片，独立于任何会话。
// Persist ID and URL against a business key; PreviewURL is temporary and must not be used as a durable reference.
// 请按业务键持久化 ID 和 URL；PreviewURL 是临时地址，不能用作持久引用。
// TTL is a sliding window in seconds, not an absolute expiry. List responses supply ExpiresAt in Unix seconds.
// TTL 是以秒计的滑动窗口，不是绝对过期时间；列表中的 ExpiresAt 是 Unix 秒。
type ImageMaterial struct {
	ID         string `json:"image_id"`
	URL        string `json:"url"`
	TTL        uint32 `json:"ttl,omitempty"`
	PreviewURL string `json:"preview_url,omitempty"`
	FileName   string `json:"file_name,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	ExpiresAt  int64  `json:"expires_at,omitempty"`
	// Width and Height are optional SDK display metadata, not fields promised by the material API.
	// Width 和 Height 是 SDK 可选显示元数据，不是素材 API 保证提供的字段。
	// UploadMarkdownImage fills supported local image dimensions; query results may leave them zero.
	// UploadMarkdownImage 填充支持的本地图片尺寸；查询结果中的尺寸可能为零。
	Width  int             `json:"-"`
	Height int             `json:"-"`
	Raw    json.RawMessage `json:"-"`
}

// UnmarshalJSON accepts both JSON numbers and decimal strings for the platform's 64-bit timestamps.
// UnmarshalJSON 接受平台 64 位时间戳的 JSON 数字及十进制字符串两种形式。
// Raw is an owned copy. Public fields remain int64, with no floating-point conversion or precision loss.
// Raw 为独立副本；公开字段保持 int64，不进行浮点转换或损失精度。
func (m *ImageMaterial) UnmarshalJSON(data []byte) error {
	type plain ImageMaterial
	var wire struct {
		plain
		CreatedAt json.RawMessage `json:"created_at"`
		ExpiresAt json.RawMessage `json:"expires_at"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	created, err := materialInt64(wire.CreatedAt)
	if err != nil {
		return fmt.Errorf("image created_at: %w", err)
	}
	expires, err := materialInt64(wire.ExpiresAt)
	if err != nil {
		return fmt.Errorf("image expires_at: %w", err)
	}
	result := ImageMaterial(wire.plain)
	result.CreatedAt = created
	result.ExpiresAt = expires
	result.Raw = append(json.RawMessage(nil), data...)
	*m = result
	return nil
}

// ImageMaterialPage is one owned page; an empty NextCursor marks the last page.
// ImageMaterialPage 是独立持有的一页结果，空 NextCursor 表示最后一页。
type ImageMaterialPage struct {
	Images     []ImageMaterial `json:"images"`
	NextCursor string          `json:"next_cursor"`
	Total      uint64          `json:"total"`
	Raw        json.RawMessage `json:"-"`
}

// UnmarshalJSON accepts a numeric or decimal-string total while preserving uint64 precision and strict bounds.
// UnmarshalJSON 接受数字或十进制字符串 total，保留 uint64 精度并严格校验边界。
func (p *ImageMaterialPage) UnmarshalJSON(data []byte) error {
	type plain ImageMaterialPage
	var wire struct {
		plain
		Total json.RawMessage `json:"total"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	var total uint64
	if len(wire.Total) > 0 {
		value, err := materialIntegerText(wire.Total)
		if err != nil {
			return fmt.Errorf("image list total: %w", err)
		}
		total, err = strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("image list total: %w", err)
		}
	}
	result := ImageMaterialPage(wire.plain)
	result.Total = total
	result.Raw = append(json.RawMessage(nil), data...)
	*p = result
	return nil
}

func materialIntegerText(data json.RawMessage) (string, error) {
	value := strings.TrimSpace(string(data))
	if strings.HasPrefix(value, `"`) {
		var decoded string
		if err := json.Unmarshal(data, &decoded); err != nil {
			return "", err
		}
		return decoded, nil
	}
	return value, nil
}

func materialInt64(data json.RawMessage) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}
	value, err := materialIntegerText(data)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(value, 10, 64)
}

// WithDimensions returns a copy with explicit Markdown display dimensions; both must be positive.
// WithDimensions 返回带显式 Markdown 显示尺寸的副本，宽高均须为正数。
// Validation is deferred to MarkdownText/Build. It never changes the opaque material URL.
// 校验延迟到 MarkdownText/Build，不修改不透明的素材 URL。
func (m ImageMaterial) WithDimensions(width, height int) ImageMaterial {
	m.Width = width
	m.Height = height
	return m
}

// MarkdownText returns an inline image reference for application templates, escaping its alt text.
// MarkdownText 返回可用于业务模板的行内图片引用，并转义替代文本。
// The platform reference is validated and preserved verbatim; zero or invalid materials return ErrInvalidMessage.
// 平台引用经过校验后原样保留；零值或无效素材返回 ErrInvalidMessage。
func (m ImageMaterial) MarkdownText(alt string) (string, error) {
	u, err := url.Parse(m.URL)
	if err != nil || u.Scheme != "qqbot-media" || u.Host != "image" || !strings.HasPrefix(u.Path, "/v1/") || len(u.Path) <= 4 || u.User != nil || u.Fragment != "" || strings.ContainsAny(m.URL, " \t\r\n()<>\\") {
		return "", fmt.Errorf("%w: valid qqbot-media image reference is required", ErrInvalidMessage)
	}
	alt = strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "#", "\\#", "\r", " ", "\n", " ").Replace(alt)
	if m.Width != 0 || m.Height != 0 {
		if m.Width <= 0 || m.Height <= 0 {
			return "", fmt.Errorf("%w: both image dimensions must be positive", ErrInvalidMessage)
		}
		alt = fmt.Sprintf("%s #%dpx #%dpx", alt, m.Width, m.Height)
	}
	return "![" + alt + "](" + m.URL + ")", nil
}

// Markdown creates an immutable outgoing message for this material; it can be combined with Buttons or Append.
// Markdown 为该素材创建不可变发送消息，可与 Buttons 或 Append 组合。
// Construction errors are reported by Build/Reply, with no network activity.
// 构造错误由 Build/Reply 报告，不执行网络操作。
func (m ImageMaterial) Markdown(alt string) Outgoing {
	content, err := m.MarkdownText(alt)
	if err != nil {
		return Markdown("").fail(err)
	}
	return Markdown(content)
}
