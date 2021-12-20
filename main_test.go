package main

import (
	"fmt"
	"testing"
)

func TestGetAllFile(t *testing.T) {
	var list []string
	GetAllFile("./dist", &list)
	fmt.Println(len(list))
}
