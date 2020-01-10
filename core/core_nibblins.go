package core

import (
	"fmt"
	"strings"
	"strconv"
)

func NibblinsToDollarString(nibblins int64) string {
	cents := float64(nibblins) / 1e9
	dollars := cents / 100
	return fmt.Sprintf("%f", dollars)
}

func DollarStringToNibblins(str string) (int64, error) {
	if len(str) == 0 {
		return 0, nil
	}
	decimal := strings.Index(str, ".")
	if decimal == -1 {
		decimal = len(str)
	}

	start := 0
	if str[0] == '-' {
		start = 1
	}

	dollars := int64(0)
	if decimal > start {
		var err error
		dollars, err = strconv.ParseInt(str[start:decimal], 10, 64)
		if err != nil {
			return 0, err
		}
	}
	nibblins := int64(0)
	if decimal+1 < len(str) {
		length := len(str) - (decimal + 1)
		if length < 11 {
			length = 11
		}
		for i := 0; i < length; i += 1 {
			if i < 11 {
				nibblins *= 10
			}
			index := decimal + 1 + i
			if index < len(str) {
				char := str[index]
				if char < byte('0') || char > byte('9') {
					return 0, fmt.Errorf("invalid dollar string: %s", str)
				}
				if i < 11 {
					nibblins += int64(char - byte('0'))
				}
			}
		}
	}
	if str[0] == '-' {
		dollars = -dollars
		nibblins = -nibblins
	}
	return (dollars * 1e11) + nibblins, nil
}
