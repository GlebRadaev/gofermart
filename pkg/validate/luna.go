package validate

import (
	"github.com/ShiraazMoollatjie/goluhn"
)

func IsLuna(s string) bool {
	err := goluhn.Validate(s)
	return err == nil
}
