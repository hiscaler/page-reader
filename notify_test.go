package pagereader

import (
	"fmt"
	"testing"
)

func TestNotify_String(t *testing.T) {
	notify := NewNotify("Test", "Hello")
	notify.AddLog("abc")
	notify.AddLogf("abc%sef", "d")
	fmt.Println(notify.String())
}
