package main

import (
	"fmt"
	"testing"
)

func TestDNS(t *testing.T) {
	t.Run("NewQuery", func(t *testing.T) {
		h := NewQuery([]byte(""))
		fmt.Printf("%v\n", h)
		fmt.Printf("%v", h.Bytes())
	})
}
