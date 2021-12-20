package weather

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	res, err := Get("101250105")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%+v\n", res)
}
