// Copyright 2026 ProgramCX. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package foundation

import (
	"os"
	"os/signal"
	"syscall"
)

func Notify() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	return sigChan
}
