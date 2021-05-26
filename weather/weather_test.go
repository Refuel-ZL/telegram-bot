package weather

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	fmt.Printf("%+v\n", Get("101250105"))
}
