package icp

import (
	"fmt"
	"testing"
)

func TestBeiAn(t *testing.T) {
	resp, err := BeiAn("baidu.com")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%s 查询,共 %d 结果\n", resp.UnitName, resp.Total)
	for _, info := range resp.List {
		fmt.Printf("%#v\n", info)
	}
}
