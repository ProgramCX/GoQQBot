// management 是 QQ 群聊管理、频道管理和频道消息的实机验收程序。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/ProgramCX/GoQQBot/pkg/api"
	"github.com/ProgramCX/GoQQBot/pkg/api/channel"
	"github.com/ProgramCX/GoQQBot/pkg/api/group"
	"github.com/ProgramCX/GoQQBot/pkg/api/robot"
	"github.com/ProgramCX/GoQQBot/pkg/config"
)

type operation struct {
	ids   string
	input string
	write bool
	run   func(context.Context, []string, []byte) (any, error)
}

type noInput struct{}

func add[T any, R any](ops map[string]operation, name, ids string, write bool, call func(context.Context, []string, T) (R, error)) {
	var zero T
	input := fmt.Sprintf("%T", zero)
	if _, ok := any(zero).(noInput); ok {
		input = ""
	}
	ops[name] = operation{ids: ids, input: input, write: write, run: func(ctx context.Context, ids []string, raw []byte) (any, error) {
		var body T
		if len(raw) > 0 {
			dec := json.NewDecoder(strings.NewReader(string(raw)))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&body); err != nil {
				return nil, fmt.Errorf("解析 %s 参数: %w", input, err)
			}
			if !json.Valid(raw) {
				return nil, errors.New("输入必须是一个 JSON 值")
			}
		}
		return call(ctx, ids, body)
	}}
}

func ok(err error) (any, error) {
	return map[string]bool{"ok": err == nil}, err
}

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		var failure *api.APIError
		if errors.As(err, &failure) {
			log.Printf("QQ API 诊断: %s  audit_id=%s", failure.DiagnosticMessage(), failure.MessageAuditID)
		}
		log.Fatal(err)
	}
}

func run() error {
	path := flag.String("config", "config.yaml", "机器人配置文件")
	name := flag.String("op", "listen", "测试操作名称")
	input := flag.String("input", "", "JSON 参数文件")
	write := flag.Bool("write", false, "允许所选操作修改 QQ 数据或发送消息")
	list := flag.Bool("list", false, "列出操作、位置参数和 JSON 参数类型")
	check := flag.Bool("check", false, "批量运行只读验收并输出结果")
	image := flag.String("image-file", "", "监听模式下 /sdk-image 使用的本地图片")
	flag.Parse()
	ops := make(map[string]operation)
	registerGroup(ops, group.NewClient(nil))
	registerChannel(ops, channel.NewClient(nil))
	registerRobot(ops, robot.NewClient(nil))
	if *list {
		names := make([]string, 0, len(ops))
		for key := range ops {
			names = append(names, key)
		}
		sort.Strings(names)
		fmt.Println("listen  接收 QQ 事件与被动回复")
		for _, key := range names {
			op := ops[key]
			fmt.Printf("%-29s %-33s %s", key, op.ids, op.input)
			if op.write {
				fmt.Print("  [-write]")
			}
			fmt.Println()
		}
		return nil
	}
	if *check && (*input == "" || *write || *image != "" || flag.NArg() != 0 || *name != "listen") {
		return errors.New("-check 需要 -input 指定验收参数，不能与单次操作或 -write 混用")
	}
	ids := flag.Args()
	if *check {
		return runChecks(*path, *input)
	}
	if *name != "listen" {
		op, found := ops[*name]
		if !found {
			return fmt.Errorf("未知操作 %q，请用 -list 查看", *name)
		}
		if len(ids) != len(strings.Fields(op.ids)) {
			return fmt.Errorf("%s 需要位置参数: %s", *name, op.ids)
		}
		for _, id := range ids {
			if strings.TrimSpace(id) == "" {
				return errors.New("ID 不能为空")
			}
		}
		if op.write && !*write {
			return errors.New("该操作需要显式指定 -write")
		}
		if *input != "" && op.input == "" {
			return errors.New("该操作无需 JSON 参数")
		}
	} else if len(ids) > 0 || *input != "" {
		return errors.New("listen 不接受位置参数或 JSON 文件")
	}
	cfg, err := config.LoadConfig(*path)
	if err != nil {
		return err
	}
	if *name == "listen" {
		return listen(*cfg, *image, *write)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	client := api.NewClient(cfg.Bot.AppId, cfg.Bot.AppSecret)
	registerGroup(ops, group.NewClient(client))
	registerChannel(ops, channel.NewClient(client))
	registerRobot(ops, robot.NewClient(client))
	var raw []byte
	if *input != "" {
		raw, err = os.ReadFile(*input)
		if err != nil {
			return err
		}
		raw = []byte(strings.TrimPrefix(string(raw), "\ufeff"))
		if len(strings.TrimSpace(string(raw))) == 0 {
			return errors.New("JSON 参数文件为空")
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	result, err := ops[*name].run(ctx, ids, raw)
	if err != nil {
		return fmt.Errorf("%s: %w", *name, err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
