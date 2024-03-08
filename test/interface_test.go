package test

import (
	"fmt"
	"testing"
)

func function(list interface{}) {
	newList, ok := list.([]interface{})
	if ok {
		print(1)
	} else {
		print(2)
		fmt.Printf("%q", newList)
	}

}

func TestInterface(t *testing.T) {
	a := []string{"1", "2", "3"}
	function(a)
}
