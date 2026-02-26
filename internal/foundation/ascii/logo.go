// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

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
