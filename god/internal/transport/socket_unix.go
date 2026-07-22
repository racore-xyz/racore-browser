//go:build !windows

package transport

import "syscall"

func setReuseAddress(fd uintptr) error {
	return syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}

func shutdownRead(fd uintptr) {
	_ = syscall.Shutdown(int(fd), syscall.SHUT_RD)
}
