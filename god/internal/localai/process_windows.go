//go:build windows

package localai

import (
	"os/exec"
	"syscall"
)

const (
	createNoWindow           = 0x08000000
	belowNormalPriorityClass = 0x00004000
)

func configureBackgroundProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNoWindow | belowNormalPriorityClass,
	}
}
