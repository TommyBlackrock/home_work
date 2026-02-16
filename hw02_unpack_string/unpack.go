package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(inputStr string) (string, error) {
	if inputStr == "" {
		return "", nil
	}

	var builder strings.Builder

	var prevRune rune
	var hasPrev bool = false
	var prevWasDigit bool = false

	flushPrev := func() {
		if hasPrev {
			builder.WriteRune(prevRune)
			hasPrev = false
		}
	}

	for _, r := range inputStr {
		if unicode.IsDigit(r) {

			if !hasPrev {
				prevWasDigit = false
				continue
			}

			if prevWasDigit {
				return "", ErrInvalidString
			}

			count, err := strconv.Atoi(string(r))
			if err != nil {
				return "", ErrInvalidString
			}

			if count == 0 {
				hasPrev = false
			} else {
				builder.WriteRune(prevRune)
				if count-1 > 0 {
					builder.WriteString(strings.Repeat(string(prevRune), count-1))
				}
				hasPrev = false
			}
			prevWasDigit = true
			continue
		}

		flushPrev()
		prevRune = r
		hasPrev = true
		prevWasDigit = false

	}
	flushPrev()
	return builder.String(), nil
}
