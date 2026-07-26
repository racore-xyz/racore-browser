//go:build !windows

package localai

import "os/exec"

func configureBackgroundProcess(_ *exec.Cmd) {}
