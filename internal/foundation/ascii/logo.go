// Copyright 2026 ProgramCX. All rights reserved.
// 版权所有 2026 ProgramCX，保留所有权利。
// Use of this source code is governed by a MIT license
// 本源代码的使用受 MIT 许可证约束，
// that can be found in the LICENSE file.
// 许可证内容见 LICENSE 文件。

package ascii

import "fmt"

var version = "dev"
var buildTime = "unknown"
var gitCommit = "unknown"

const logo = `
   ______          ____    ____     ____           __
  / ____/  ____   / __ \  / __ \   / __ )  ____   / /_
 / / __   / __ \ / / / / / / / /  / __  | / __ \ / __/
/ /_/ /  / /_/ // /_/ / / /_/ /  / /_/ / / /_/ // /_
\____/   \____/ \___\_\ \___\_\ /_____/  \____/ \__/
`

func PrintLogo() {
	fmt.Print(logo)
	fmt.Printf("Version: %s\nBuild: %s\nCommit: %s\n", version, buildTime, gitCommit)
}
