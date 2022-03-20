package serverhttp

import "strconv"

type ID uint64

// 以下代码取自etcd源码：client\pkg\types\id.go
func (i ID) String() string {
	return strconv.FormatUint(uint64(i), 16)
}
