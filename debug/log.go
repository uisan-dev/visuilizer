package debug

import (
	"fmt"
	"visuilizer/config"
)

func Debugf(str string, a ...any) {
	if config.Debug {
		fmt.Printf("[DEBUG] "+str, a...)
	}
}
