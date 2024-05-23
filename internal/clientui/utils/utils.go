package utils

import (
	"fmt"
	"io"

	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

func Str(w io.Writer, s string, a ...any) {
	_, _ = fmt.Fprintf(w, s, a...)
}

func AddrToString(addr *pb.SockAddr) string {
	return fmt.Sprintf("%s:%d", addr.Ip, addr.Port)
}
