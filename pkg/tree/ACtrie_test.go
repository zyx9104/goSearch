package tree

import (
	"testing"
)

func TestACtire(t *testing.T) {
	act := NewAC()
	type Test struct {
		str string
		res bool
	}
	strs := []string{"123", "123456", "一二三四", "一二三四五", "一二三四", "一二", "今天"}
	test := []Test{
		{"123125,3346436346", true},
		{"yi安抚一二", true},
		{"三四五六", false},
		{"五六七八", false},
		{"今天星期一", true},
		{"明天星期几", false},
	}

	act.Build(strs)
	for _, item := range test {
		if act.Find(item.str) != item.res {
			t.Error(item.str, "not matched")
		}
	}
}
