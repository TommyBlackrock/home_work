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
	hasPrev := false
	prevWasDigit := false
	isFirstChar := true

	flushPrev := func() {
		if hasPrev {
			builder.WriteRune(prevRune)
			hasPrev = false
		}
	}

	for _, r := range inputStr {
		if !unicode.IsDigit(r) {
			flushPrev()
			prevRune = r
			hasPrev = true
			prevWasDigit = false
			isFirstChar = false
			continue
		}

		if isFirstChar {
			return "", ErrInvalidString
		}

		if prevWasDigit {
			return "", ErrInvalidString
		}

		if !hasPrev {
			prevWasDigit = false
			isFirstChar = false
			continue
		}

		count, err := strconv.Atoi(string(r))
		if err != nil {
			return "", ErrInvalidString
		}

		if count == 0 {
			hasPrev = false
			prevWasDigit = true
			isFirstChar = false
			continue
		}

		builder.WriteRune(prevRune)
		if count > 1 {
			builder.WriteString(strings.Repeat(string(prevRune), count-1))
		}

		hasPrev = false
		prevWasDigit = true
		isFirstChar = false
	}
	flushPrev()
	return builder.String(), nil
}
