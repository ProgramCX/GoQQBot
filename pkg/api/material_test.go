package api

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func TestMarkdownImageFullUploadPreservesMIMEAndETag(t *testing.T) {
	for _, format := range []struct{ extension, mime string }{{".png", "image/png"}, {".jpeg", "image/jpeg"}, {".jpg", "image/jpeg"}, {".gif", "image/gif"}, {".webp", "image/webp"}} {
		t.Run(format.extension, func(t *testing.T) {
			data := []byte("stable generated image bytes")
			var encoded bytes.Buffer
			picture := image.NewRGBA(image.Rect(0, 0, 13, 7))
			var encodeErr error
			switch format.mime {
			case "image/png":
				encodeErr = png.Encode(&encoded, picture)
			case "image/jpeg":
				encodeErr = jpeg.Encode(&encoded, picture, nil)
			case "image/gif":
				encodeErr = gif.Encode(&encoded, picture, nil)
			}
			if encodeErr != nil {
				t.Fatal(encodeErr)
			}
			if encoded.Len() > 0 {
				data = encoded.Bytes()
			}
			sum := digestMD5(data)
			var steps []string
			jar, _ := cookiejar.New(nil)
			storageURL, _ := url.Parse("https://storage.test")
			jar.SetCookies(storageURL, []*http.Cookie{{Name: "private", Value: "cookie"}})
			client := NewClient("test", "secret", WithHTTPClient(&http.Client{Jar: jar, Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/app/getAppAccessToken" {
					return response(`{"access_token":"token","expires_in":7200}`), nil
				}
				if r.URL.Host == "storage.test" {
					if r.Method != "PUT" || r.Header.Get("Content-Type") != format.mime || r.ContentLength != int64(len(data)) || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
						t.Fatal("incorrect MIME, size or storage authentication")
					}
					got, err := io.ReadAll(r.Body)
					if err != nil || !bytes.Equal(got, data) {
						t.Fatalf("binary bytes: %v", err)
					}
					steps = append(steps, "put")
					res := response("")
					res.Header.Set("ETag", `"`+strings.ToUpper(sum)+`"`)
					return res, nil
				}
				if r.Header.Get("Authorization") != "QQBot token" {
					t.Fatal("API lacks bot token")
				}
				switch r.URL.Path {
				case "/v2/media-images/prepare":
					var input message.ImagePrepareRequest
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						t.Fatal(err)
					}
					sha := sha1.Sum(data)
					if input.FileSize != uint64(len(data)) || input.ContentType != format.mime || input.FileName != "chart"+format.extension || input.MD5 != sum || input.SHA1 != hex.EncodeToString(sha[:]) {
						t.Fatal("incorrect image registration")
					}
					steps = append(steps, "prepare")
					return response(`{"upload_id":"task","presigned_url":"https://storage.test/image?signature=hidden","expires_in":3600}`), nil
				case "/v2/media-images/complete":
					var input message.ImageCompleteRequest
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						t.Fatal(err)
					}
					if input.UploadID != "task" || input.MD5 != sum {
						t.Fatal("completion lost ETag or upload ID")
					}
					steps = append(steps, "complete")
					return response(`{"image_id":"image","url":"qqbot-media://image/v1/ref","ttl":7776000,"preview_url":"https://storage.test/preview","future":true}`), nil
				default:
					t.Fatal("upload unexpectedly sent a message")
					return nil, nil
				}
			})}))
			material, err := client.UploadMarkdownImage(context.Background(), ImageUpload{Reader: bytes.NewReader(data), Size: int64(len(data)), Name: "chart" + format.extension})
			if err != nil || material == nil || material.ID != "image" || material.TTL != 7776000 || !strings.Contains(string(material.Raw), `"future":true`) {
				t.Fatalf("upload result: %v", err)
			}
			if format.mime != "image/webp" && (material.Width != 13 || material.Height != 7) {
				t.Fatal("local image dimensions not retained")
			}
			if strings.Join(steps, ",") != "prepare,put,complete" {
				t.Fatalf("upload steps: %v", steps)
			}
		})
	}
}

func TestMarkdownImageListPaginationAndDelete(t *testing.T) {
	cursor := "next+/=&?"
	calls := 0
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/app/getAppAccessToken" {
			return response(`{"access_token":"token","expires_in":7200}`), nil
		}
		calls++
		if r.Method == "DELETE" {
			if r.URL.EscapedPath() != "/v2/media-images/id%2Fwith%3Fquery/delete" {
				t.Fatal("image ID was not escaped")
			}
			if calls == 4 {
				res := response(`{"code":40035008}`)
				res.StatusCode = 400
				return res, nil
			}
			return response(""), nil
		}
		if r.URL.Path != "/v2/media-images" {
			t.Fatal("incorrect list endpoint")
		}
		if calls == 1 {
			if r.URL.RawQuery != "" {
				t.Fatal("first page sent cursor")
			}
			return response(`{"images":[{"image_id":"image","url":"qqbot-media://image/v1/ref","file_name":"chart.png","created_at":"100","expires_at":"200"}],"next_cursor":"next+/=&?","total":"1"}`), nil
		}
		if r.URL.Query().Get("cursor") != cursor {
			t.Fatal("cursor was not encoded losslessly")
		}
		return response(`{"images":[],"next_cursor":"","total":1}`), nil
	})}))
	first, err := client.ListMarkdownImages(context.Background(), "")
	if err != nil || first.NextCursor != cursor || first.Images[0].ExpiresAt != 200 {
		t.Fatalf("first page: %v", err)
	}
	last, err := client.ListMarkdownImages(context.Background(), first.NextCursor)
	if err != nil || last.NextCursor != "" || len(last.Images) != 0 {
		t.Fatalf("last page: %v", err)
	}
	if err := client.DeleteMarkdownImage(context.Background(), "id/with?query"); err != nil {
		t.Fatal(err)
	}
	err = client.DeleteMarkdownImage(context.Background(), "id/with?query")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 40035008 {
		t.Fatal("delete suppressed platform not-found")
	}
}

func TestMarkdownImageRejectsBadInputBeforeNetwork(t *testing.T) {
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { t.Fatal("invalid input reached network"); return nil, nil })}))
	for _, file := range []ImageUpload{{Name: "image.png", Size: 1}, {Reader: strings.NewReader("x"), Name: "image.bmp", Size: 1}, {Reader: strings.NewReader("x"), Name: "image.png", ContentType: "image/jpeg", Size: 1}, {Reader: strings.NewReader("x"), Name: "image.png", Size: MaxImageMaterialSize + 1}, {Reader: strings.NewReader("x"), Name: "image.png", Size: 0}} {
		if _, err := client.UploadMarkdownImage(context.Background(), file); !errors.Is(err, ErrInvalidImageMaterial) {
			t.Fatalf("invalid image accepted: %v", err)
		}
	}
	if _, err := client.UploadMarkdownImage(context.Background(), ImageUpload{Reader: strings.NewReader("x"), Name: "image.png", Size: 2}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("short file: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.UploadMarkdownImage(ctx, ImageUpload{Reader: strings.NewReader("x"), Name: "image.png", Size: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := client.UploadMarkdownImage(nil, ImageUpload{}); !errors.Is(err, ErrInvalidImageMaterial) {
		t.Fatal("nil operation context")
	}
	if err := client.DeleteMarkdownImage(context.Background(), ""); !errors.Is(err, ErrInvalidImageMaterial) {
		t.Fatal("empty deletion")
	}
	if _, err := client.UploadMarkdownImageFile(context.Background(), t.TempDir()); !errors.Is(err, ErrInvalidImageMaterial) {
		t.Fatal("directory accepted")
	}
	if _, err := client.UploadMarkdownImageFile(context.Background(), filepath.Join(t.TempDir(), "missing.png")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file error not wrapped: %v", err)
	}
}

func TestMarkdownImageStopsOnStorageAndCompletionFailures(t *testing.T) {
	data := []byte("abc")
	sum := digestMD5(data)
	for _, mode := range []string{"prepare", "put", "redirect", "missing-etag", "bad-etag", "mismatch", "complete", "bad-response"} {
		t.Run(mode, func(t *testing.T) {
			puts, completes := 0, 0
			client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				switch {
				case r.URL.Path == "/app/getAppAccessToken":
					return response(`{"access_token":"token","expires_in":7200}`), nil
				case r.URL.Path == "/v2/media-images/prepare":
					if mode == "prepare" {
						return response(`{"code":40035007}`), nil
					}
					return response(`{"upload_id":"task","presigned_url":"https://storage.test/image?secret=private","expires_in":3600}`), nil
				case r.URL.Host == "storage.test":
					puts++
					_, _ = io.Copy(io.Discard, r.Body)
					res := response("")
					res.Header.Set("ETag", sum)
					switch mode {
					case "put":
						res.StatusCode = 403
					case "redirect":
						res.StatusCode = 307
						res.Header.Set("Location", "https://elsewhere.test")
					case "missing-etag":
						res.Header.Del("ETag")
					case "bad-etag":
						res.Header.Set("ETag", "not-an-md5")
					case "mismatch":
						res.Header.Set("ETag", strings.Repeat("0", 32))
					}
					return res, nil
				case r.URL.Path == "/v2/media-images/complete":
					completes++
					if mode == "complete" {
						return response(`{"code":40035003}`), nil
					}
					return response(`{"image_id":"created","url":"qqbot-media://image/v1/ref","ttl":"invalid"}`), nil
				default:
					t.Fatal("unexpected request/retry")
					return nil, nil
				}
			})}))
			result, err := client.UploadMarkdownImage(context.Background(), ImageUpload{Reader: bytes.NewReader(data), Size: 3, Name: "a.png"})
			if err == nil || strings.Contains(err.Error(), "secret=private") {
				t.Fatalf("failure or signed URL leak: %v", err)
			}
			if mode == "prepare" {
				if puts != 0 {
					t.Fatal("prepare failure transferred data")
				}
			} else if puts != 1 {
				t.Fatal("PUT was retried")
			}
			if mode == "complete" || mode == "bad-response" {
				if completes != 1 {
					t.Fatal("completion retried or skipped")
				}
			} else if completes != 0 {
				t.Fatal("invalid transfer was completed")
			}
			if mode == "bad-response" && (result == nil || !json.Valid(result.Raw)) {
				t.Fatal("lost successful completion response")
			}
		})
	}
}
