package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ProgramCX/GoQQBot/pkg/message"
)

func digestMD5(b []byte) string { sum := md5.Sum(b); return hex.EncodeToString(sum[:]) }

func TestUploadFailuresIdentifyStageAndPreserveAPIError(t *testing.T) {
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		for _, stage := range []string{"prepare", "finish", "merge"} {
			t.Run(string(scene)+"/"+stage, func(t *testing.T) {
				client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
					fail := func() (*http.Response, error) {
						res := response(`{"code":40093004,"trace_id":"trace-test","message":"do-not-log-this-response"}`)
						res.StatusCode = 400
						return res, nil
					}
					switch {
					case r.URL.Path == "/app/getAppAccessToken":
						return response(`{"access_token":"token","expires_in":7200}`), nil
					case strings.HasSuffix(r.URL.Path, "/upload_prepare"):
						if stage == "prepare" {
							return fail()
						}
						return response(`{"upload_id":"task","block_size":"3","parts":[{"index":0,"block_size":"3","presigned_url":"https://storage.test/part?secret=do-not-log"}]}`), nil
					case r.Method == "PUT":
						_, _ = io.Copy(io.Discard, r.Body)
						return response(""), nil
					case strings.HasSuffix(r.URL.Path, "/upload_part_finish"):
						if stage == "finish" {
							return fail()
						}
						return response(""), nil
					case strings.HasSuffix(r.URL.Path, "/files"):
						var body map[string]json.RawMessage
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Fatal(err)
						}
						if _, exists := body["url"]; exists {
							t.Fatal("merge request contains a URL source")
						}
						return fail()
					default:
						t.Fatal("unexpected request")
						return nil, nil
					}
				})}))
				_, err := client.UploadFile(context.Background(), message.Target{Scene: scene, ID: "recipient"}, FileUpload{Reader: strings.NewReader("abc"), Size: 3, Name: "file.bin", Type: 4})
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.Code != 40093004 || apiErr.TraceID != "trace-test" || apiErr.StatusCode != 400 {
					t.Fatalf("lost API error: %v", err)
				}
				if !strings.Contains(err.Error(), stage+" media") || strings.Contains(err.Error(), "do-not-log") {
					t.Fatalf("missing stage or leaked response/URL: %v", err)
				}
			})
		}
	}
}

func TestUploadFilePreservesServerIndicesAndUsesRelativeOffsets(t *testing.T) {
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		for _, base := range []int{0, 1} {
			for _, block := range []int{4, 32} {
				t.Run(fmt.Sprintf("%s/base%d/block%d", scene, base, block), func(t *testing.T) {
					data := []byte("abcdefghijk")
					var parts []message.UploadPart
					for slot := 0; slot*block < len(data); slot++ {
						parts = append(parts, message.UploadPart{Index: slot + base, BlockSize: strconv.Itoa(min(block, len(data)-slot*block)), PresignedURL: fmt.Sprintf("https://storage.test/%d", slot)})
					}
					// Plans may arrive in reverse order; indices, not slice order, select offsets.
					// 计划可能倒序返回；由索引而非切片顺序确定偏移。
					for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
						parts[i], parts[j] = parts[j], parts[i]
					}
					plan, _ := json.Marshal(message.UploadPrepareResult{UploadID: "task", BlockSize: strconv.Itoa(block), Parts: parts})
					puts, finishes := 0, 0
					client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
						switch {
						case r.URL.Path == "/app/getAppAccessToken":
							return response(`{"access_token":"token","expires_in":7200}`), nil
						case strings.HasSuffix(r.URL.Path, "/upload_prepare"):
							return response(string(plan)), nil
						case r.URL.Host == "storage.test":
							slot, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
							if err != nil {
								t.Fatal(err)
							}
							got, err := io.ReadAll(r.Body)
							if err != nil {
								t.Fatal(err)
							}
							if slot != puts || !bytes.Equal(got, data[slot*block:min((slot+1)*block, len(data))]) {
								t.Fatalf("wrong file offset for base %d, slot %d", base, slot)
							}
							puts++
							return response(""), nil
						case strings.HasSuffix(r.URL.Path, "/upload_part_finish"):
							var input message.UploadPartFinishRequest
							if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
								t.Fatal(err)
							}
							start, end := finishes*block, min((finishes+1)*block, len(data))
							if input.PartIndex != finishes+base || input.MD5 != digestMD5(data[start:end]) || input.BlockSize != strconv.Itoa(end-start) {
								t.Fatalf("completion changed server index or checksum: index=%d", input.PartIndex)
							}
							finishes++
							return response(""), nil
						case strings.HasSuffix(r.URL.Path, "/files"):
							if puts != len(parts) || finishes != len(parts) {
								t.Fatal("merged before completing all parts")
							}
							return response(`{"file_info":"uploaded","ttl":60}`), nil
						default:
							t.Fatal("unexpected operation")
							return nil, nil
						}
					})}))
					result, err := client.UploadFile(context.Background(), message.Target{Scene: scene, ID: "recipient"}, FileUpload{Reader: bytes.NewReader(data), Size: int64(len(data)), Name: "image.jpg", Type: message.FileTypeImage})
					if err != nil || result == nil || result.FileInfo != "uploaded" {
						t.Fatalf("upload: %v", err)
					}
				})
			}
		}
	}
}

func TestUploadFileBothScenesAndAllMediaTypes(t *testing.T) {
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		for _, kind := range []message.FileType{message.FileTypeImage, message.FileTypeVideo, message.FileTypeVoice, message.FileTypeFile} {
			t.Run(fmt.Sprintf("%s/%d", scene, kind), func(t *testing.T) {
				data := []byte("abcdefghijk")
				prefix := "/v2/users/recipient"
				if scene == message.SceneGroup {
					prefix = "/v2/groups/recipient"
				}
				var steps []string
				client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
					if r.URL.Path == "/app/getAppAccessToken" {
						return response(`{"access_token":"test-token","expires_in":7200}`), nil
					}
					if r.URL.Host == "storage.test" {
						if r.Method != "PUT" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
							t.Fatal("invalid storage authorization")
						}
						idx, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
						start := idx * 4
						end := min(start+4, len(data))
						actual, err := io.ReadAll(r.Body)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(actual, data[start:end]) || r.ContentLength != int64(end-start) {
							t.Fatalf("wrong part %d", idx)
						}
						steps = append(steps, fmt.Sprintf("put%d", idx))
						return response(""), nil
					}
					if r.Header.Get("Authorization") != "QQBot test-token" {
						t.Fatal("API authentication absent")
					}
					switch r.URL.Path {
					case prefix + "/upload_prepare":
						var input message.UploadPrepareRequest
						if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
							t.Fatal(err)
						}
						sum := sha1.Sum(data)
						if input.FileType != kind || input.FileSize != "11" || input.FileName != "sample.bin" || input.MD5 != digestMD5(data) || input.MD5TenM != digestMD5(data) || input.SHA1 != hex.EncodeToString(sum[:]) {
							t.Fatalf("wrong metadata: type=%d size=%s", input.FileType, input.FileSize)
						}
						steps = append(steps, "prepare")
						return response(`{"upload_id":"task","block_size":"4","parts":[{"index":2,"block_size":"3","presigned_url":"https://storage.test/2"},{"index":0,"block_size":"4","presigned_url":"https://storage.test/0"},{"index":1,"block_size":"4","presigned_url":"https://storage.test/1"}]}`), nil
					case prefix + "/upload_part_finish":
						var input message.UploadPartFinishRequest
						if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
							t.Fatal(err)
						}
						start := input.PartIndex * 4
						end := min(start+4, len(data))
						if input.UploadID != "task" || input.BlockSize != strconv.Itoa(end-start) || input.MD5 != digestMD5(data[start:end]) {
							t.Fatal("invalid completion metadata")
						}
						steps = append(steps, fmt.Sprintf("finish%d", input.PartIndex))
						return response(""), nil
					case prefix + "/files":
						var input message.RichMediaRequest
						if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
							t.Fatal(err)
						}
						if input.UploadID != "task" || input.FileType != kind || input.FileName != "sample.bin" || input.SrvSendMsg || input.URL != "" {
							t.Fatal("merge sent proactively or used incorrect source")
						}
						steps = append(steps, "merge")
						return response(`{"file_info":"opaque","file_uuid":"uuid","ttl":60,"raw_url":"https://storage.test/download"}`), nil
					default:
						t.Fatalf("unexpected path: %s", r.URL.Path)
						return nil, nil
					}
				})}))
				result, err := client.UploadFile(context.Background(), message.Target{Scene: scene, ID: "recipient"}, FileUpload{Reader: bytes.NewReader(data), Size: int64(len(data)), Name: "sample.bin", Type: kind})
				if err != nil || result.FileInfo != "opaque" || result.TTL != 60 || result.RawURL == "" {
					t.Fatalf("upload: %v", err)
				}
				if strings.Join(steps, ",") != "prepare,put0,finish0,put1,finish1,put2,finish2,merge" {
					t.Fatalf("incorrect order: %v", steps)
				}
			})
		}
	}
}

func TestUploadHashPrefixAndShortFile(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 10002432+5)
	copy(data[10002432:], "bcdef")
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/app/getAppAccessToken" {
			return response(`{"access_token":"test-token","expires_in":7200}`), nil
		}
		var input message.UploadPrepareRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.MD5TenM != digestMD5(data[:10002432]) || input.MD5 != digestMD5(data) || input.MD5 == input.MD5TenM {
			t.Fatal("incorrect 10002432-byte prefix checksum")
		}
		return response(`{"upload_id":"task","block_size":"0"}`), nil
	})}))
	target := message.Target{Scene: message.SceneC2C, ID: "user"}
	if _, err := client.UploadFile(context.Background(), target, FileUpload{Reader: bytes.NewReader(data), Size: int64(len(data)), Name: "file", Type: 4}); !errors.Is(err, ErrInvalidMedia) {
		t.Fatalf("invalid manifest: %v", err)
	}
	if _, err := client.UploadFile(context.Background(), target, FileUpload{Reader: bytes.NewReader([]byte("a")), Size: 2, Name: "file", Type: 4}); !errors.Is(err, io.EOF) {
		t.Fatalf("short file: %v", err)
	}
}

func TestUploadPartSafetyCancellationAndFailure(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	storageURL, _ := url.Parse("https://storage.test")
	jar.SetCookies(storageURL, []*http.Cookie{{Name: "private", Value: "do-not-forward"}})
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Jar: jar, Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Fatal("storage leaked credentials")
		}
		switch r.URL.Path {
		case "/redirect":
			res := response("")
			res.StatusCode = 307
			res.Header.Set("Location", "https://elsewhere.test")
			return res, nil
		case "/fail":
			return nil, fmt.Errorf("network failed")
		default:
			return nil, r.Context().Err()
		}
	})}))
	for _, path := range []string{"redirect", "fail"} {
		err := client.UploadPart(context.Background(), message.UploadPart{PresignedURL: "https://storage.test/" + path + "?secret=do-not-log"}, strings.NewReader("bytes"), 5)
		if err == nil || strings.Contains(err.Error(), "do-not-log") {
			t.Fatalf("storage error: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.UploadPart(ctx, message.UploadPart{PresignedURL: "https://storage.test/cancel"}, strings.NewReader("b"), 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	if err := client.UploadPart(context.Background(), message.UploadPart{PresignedURL: "http://storage.test"}, strings.NewReader("b"), 1); !errors.Is(err, ErrInvalidMedia) {
		t.Fatal("insecure URL accepted")
	}
}

func TestUploadPartReleasesBorrowedReaderBeforeReturning(t *testing.T) {
	var retained io.ReadCloser
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		retained = r.Body
		return response(""), nil
	})}))
	source := strings.NewReader("borrowed")
	if err := client.UploadPart(context.Background(), message.UploadPart{PresignedURL: "https://storage.test/part"}, source, 8); err != nil {
		t.Fatal(err)
	}
	// A transport retaining the request cannot keep touching caller-owned data after return.
	// 持有请求的传输在返回后不得继续访问调用方数据。
	if data, err := io.ReadAll(retained); err != nil || len(data) != 0 {
		t.Fatalf("borrowed reader remained accessible: %v", err)
	}
	if source.Len() != 8 {
		t.Fatal("late transport consumed caller data")
	}
}

func TestUploadRejectsBadPlansAndStopsBeforeMerge(t *testing.T) {
	for name, plan := range map[string]string{
		"one-based-gap": `{"upload_id":"task","block_size":"2","parts":[{"index":1,"block_size":"2","presigned_url":"https://storage.test/a"},{"index":3,"block_size":"2","presigned_url":"https://storage.test/b"}]}`,
		"mixed-indices": `{"upload_id":"task","block_size":"2","parts":[{"index":0,"block_size":"2","presigned_url":"https://storage.test/a"},{"index":2,"block_size":"2","presigned_url":"https://storage.test/b"}]}`,
		"duplicate":     `{"upload_id":"task","block_size":"2","parts":[{"index":0,"block_size":"2","presigned_url":"https://storage.test/a"},{"index":0,"block_size":"2","presigned_url":"https://storage.test/b"}]}`,
		"wrong-size":    `{"upload_id":"task","block_size":"2","parts":[{"index":0,"block_size":"2","presigned_url":"https://storage.test/a"},{"index":1,"block_size":"9","presigned_url":"https://storage.test/b"}]}`,
		"unsafe-url":    `{"upload_id":"task","block_size":"2","parts":[{"index":0,"block_size":"2","presigned_url":"https://storage.test/a"},{"index":1,"block_size":"2","presigned_url":"http://storage.test/b"}]}`,
		"failed-put":    `{"upload_id":"task","block_size":"4","parts":[{"index":0,"block_size":"4","presigned_url":"https://storage.test/a"}]}`,
		"failed-finish": `{"upload_id":"task","block_size":"4","parts":[{"index":0,"block_size":"4","presigned_url":"https://storage.test/a"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			transfers, finishes := 0, 0
			client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				switch {
				case r.URL.Path == "/app/getAppAccessToken":
					return response(`{"access_token":"token","expires_in":7200}`), nil
				case strings.HasSuffix(r.URL.Path, "/upload_prepare"):
					return response(plan), nil
				case r.Method == "PUT":
					transfers++
					_, _ = io.Copy(io.Discard, r.Body)
					res := response("")
					if name == "failed-put" {
						res.StatusCode = 503
					}
					return res, nil
				case strings.HasSuffix(r.URL.Path, "/upload_part_finish"):
					finishes++
					return response(`{"code":40093001}`), nil
				default:
					t.Fatal("failure reached merge")
					return nil, nil
				}
			})}))
			_, err := client.UploadFile(context.Background(), message.Target{Scene: message.SceneGroup, ID: "group"}, FileUpload{Reader: strings.NewReader("abcd"), Size: 4, Name: "sample", Type: 4})
			if err == nil {
				t.Fatal("bad plan/transfer succeeded")
			}
			if strings.HasPrefix(name, "failed-") {
				if transfers != 1 || (name == "failed-put" && finishes != 0) || (name == "failed-finish" && finishes != 1) {
					t.Fatal("failed transfer was retried or acknowledged")
				}
			} else if transfers != 0 || !errors.Is(err, ErrInvalidMedia) {
				t.Fatal("invalid manifest transferred bytes")
			}
		})
	}
}

func TestRecallBothScenesEscapesAndPlatformErrors(t *testing.T) {
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		for _, code := range []int{0, 40062003, 40064004} {
			client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/app/getAppAccessToken" {
					return response(`{"access_token":"token","expires_in":7200}`), nil
				}
				if r.Method != "DELETE" || !strings.Contains(r.URL.EscapedPath(), "a%2Fb/messages/m%2Fn") {
					t.Fatal("recall did not escape IDs")
				}
				if code == 0 {
					return response(""), nil
				}
				return response(fmt.Sprintf(`{"code":%d}`, code)), nil
			})}))
			err := client.RecallMessage(context.Background(), message.Target{Scene: scene, ID: "a/b"}, "m/n")
			var failure *APIError
			if code == 0 {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.As(err, &failure) || failure.Code != code {
				t.Fatalf("recall: %v", err)
			}
		}
	}
}

func TestURLAndInlineUploadSources(t *testing.T) {
	for _, scene := range []message.Scene{message.SceneC2C, message.SceneGroup} {
		for _, proactive := range []bool{false, true} {
			for _, source := range []string{"url", "inline"} {
				calls := 0
				client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
					if r.URL.Path == "/app/getAppAccessToken" {
						return response(`{"access_token":"token","expires_in":7200}`), nil
					}
					calls++
					var req message.RichMediaRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Fatal(err)
					}
					if req.SrvSendMsg != proactive || (req.URL == "" && req.FileData == "") {
						t.Fatal("source or send flag lost")
					}
					return response(`{"file_info":"info","ttl":0}`), nil
				})}))
				input := message.RichMediaRequest{FileType: 1, SrvSendMsg: proactive, URL: "https://source.test/a.png"}
				if source == "inline" {
					input.URL = ""
					input.FileName = "a.png"
					input.FileData = "opaque-protocol-string"
				}
				if _, err := client.UploadMedia(context.Background(), message.Target{Scene: scene, ID: "recipient"}, input); err != nil || calls != 1 {
					t.Fatalf("upload: %v", err)
				}
				input.UploadID = "conflict"
				if _, err := client.UploadMedia(context.Background(), message.Target{Scene: scene, ID: "recipient"}, input); !errors.Is(err, ErrInvalidMedia) || calls != 1 {
					t.Fatal("conflicting sources reached network")
				}
			}
		}
	}
}
