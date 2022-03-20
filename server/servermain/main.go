package servermain

import (
	"fmt"
	"os"
)

func Main() {
	s, err := NewServer()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	s.StartServer()
}
