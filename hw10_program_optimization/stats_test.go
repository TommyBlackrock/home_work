//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	t.Run("ignore invalid email without at sign", func(t *testing.T) {
		data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"invalid-email","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}`

		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsecat.com": 1}, result)
	})
}

func TestGetDomainStatExactSuffixAndCase(t *testing.T) {
	data := `{"Email":"first@Example.BIZ"}
{"Email":"second@example.biz"}
{"Email":"third@example.notbiz"}
{"Email":"fourth@example.biz.test"}
{"Email":"invalid-email"}`

	result, err := GetDomainStat(bytes.NewBufferString(data), "BIZ")

	require.NoError(t, err)
	require.Equal(t, DomainStat{"example.biz": 2}, result)
}

func TestGetDomainStatReportsJSONLine(t *testing.T) {
	data := `{"Email":"first@example.biz"}
{"Email":"broken@example.biz"
{"Email":"third@example.biz"}`

	result, err := GetDomainStat(bytes.NewBufferString(data), "biz")

	require.Nil(t, result)
	require.ErrorContains(t, err, "decode user at line 2")
}

func TestGetDomainStatDoesNotHideInvalidJSONAfterEmail(t *testing.T) {
	data := `{"Email":"first@example.biz","Broken":}`

	result, err := GetDomainStat(bytes.NewBufferString(data), "biz")

	require.Nil(t, result)
	require.ErrorContains(t, err, "decode user at line 1")
}

func TestGetDomainStatReturnsReaderError(t *testing.T) {
	expectedErr := errors.New("read failure")
	reader := io.MultiReader(
		strings.NewReader(`{"Email":"first@example.biz"}`),
		errorReader{err: expectedErr},
	)

	result, err := GetDomainStat(reader, "biz")

	require.Nil(t, result)
	require.ErrorIs(t, err, expectedErr)
}

func TestGetDomainStatSupportsLargeLines(t *testing.T) {
	largeName := strings.Repeat("a", 128*1024)
	data := `{"Name":"` + largeName + `","Email":"first@example.biz"}`

	result, err := GetDomainStat(bytes.NewBufferString(data), "biz")

	require.NoError(t, err)
	require.Equal(t, DomainStat{"example.biz": 1}, result)
}

func TestGetDomainStatReportsLineAcrossBatches(t *testing.T) {
	const invalidLine = linesPerBatch + 25

	var data strings.Builder
	for line := 1; line <= invalidLine+10; line++ {
		if line == invalidLine {
			data.WriteString(`{"Email":"broken@example.biz"`)
		} else {
			data.WriteString(`{"Email":"valid@example.biz"}`)
		}
		data.WriteByte('\n')
	}

	result, err := GetDomainStat(strings.NewReader(data.String()), "biz")

	require.Nil(t, result)
	require.ErrorContains(t, err, fmt.Sprintf("decode user at line %d", invalidLine))
}

type errorReader struct {
	err error
}

func (r errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
