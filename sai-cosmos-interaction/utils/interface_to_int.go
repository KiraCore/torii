package utils

import "fmt"

func IfaceToInt64(digit interface{}) (int64, error) {
	digitFloat64, ok := digit.(float64)
	if !ok {
		return 0, fmt.Errorf("wrong interface type")
	}

	return int64(digitFloat64), nil
}
