package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		// uncomment if task with asterisk completed
		// {input: `qwe\4\5`, expected: `qwe45`},
		// {input: `qwe\45`, expected: `qwe44444`},
		// {input: `qwe\\5`, expected: `qwe\\\\\`},
		// {input: `qwe\\\3`, expected: `qwe\3`},
		// additional test cases
		{input: "ф2", expected: "фф"},
		{input: "я3", expected: "яяя"},
		{input: "你2", expected: "你你"},
		{input: "あ2", expected: "ああ"},
		{input: "🙂2", expected: "🙂🙂"},
		{input: "😎3", expected: "😎😎😎"},
		{input: "😈0", expected: ""},
		// extended test cases
		{input: "!3", expected: "!!!"},
		{input: ",2", expected: ",,"},
		{input: "a!2", expected: "a!!"},
		{input: ".0", expected: ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{
		"3abc",
		"45",
		"aaa10b",
		"a12", // две цифры подряд
		"🙂34", // unicode + цифры подряд
		"😈99", // emoji + цифры подряд

		"1",   // одиночная цифра
		"123", // только цифры
		"00",  // только цифры
		"a01", // число
	}
	for _, tc := range invalidStrings {
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}
