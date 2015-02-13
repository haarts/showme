package main

import (
	"fmt"
	"testing"
)

func TestNewIndex(t *testing.T) {
	index := NewIndex()
	fmt.Printf("index %+v\n", index)
}
