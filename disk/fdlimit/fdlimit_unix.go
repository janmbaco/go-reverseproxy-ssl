// +build linux netbsd openbsd solaris

package fdlimit

import "syscall"

// Current retrieves the number of file descriptors allowed to be opened by this
// process.
func Get() (int, error) {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return 0, err
	}
	return int(limit.Cur), nil
}
