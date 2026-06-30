package hw10programoptimization

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/valyala/fastjson" //nolint:depguard // Homework permits third-party JSON parsers.
)

const oldBenchmarkEnvironment = "HW10_BENCH_IMPL"

type domainStatFunc func(io.Reader, string) (DomainStat, error)

func BenchmarkGetDomainStat(b *testing.B) {
	data := readBenchmarkInput(b)
	benchmarkDomainStat(b, selectedBenchmarkImplementation(), data)
}

func readBenchmarkInput(b *testing.B) []byte {
	b.Helper()

	archive, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer archive.Close()

	if len(archive.File) != 1 {
		b.Fatalf("unexpected number of files in archive: %d", len(archive.File))
	}

	reader, err := archive.File[0].Open()
	if err != nil {
		b.Fatal(err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		b.Fatal(err)
	}

	return data
}

func BenchmarkGetDomainStatZIP(b *testing.B) {
	archive, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer archive.Close()

	if len(archive.File) != 1 {
		b.Fatalf("unexpected number of files in archive: %d", len(archive.File))
	}

	implementation := selectedBenchmarkImplementation()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader, err := archive.File[0].Open()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		stat, runErr := implementation(reader, "biz")

		b.StopTimer()
		closeErr := reader.Close()
		b.StartTimer()

		if runErr != nil {
			b.Fatal(runErr)
		}
		if closeErr != nil {
			b.Fatal(closeErr)
		}
		checkBenchmarkResult(b, stat)
	}
}

func benchmarkDomainStat(b *testing.B, implementation domainStatFunc, data []byte) {
	b.Helper()
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stat, err := implementation(bytes.NewReader(data), "biz")
		if err != nil {
			b.Fatal(err)
		}
		checkBenchmarkResult(b, stat)
	}
}

func selectedBenchmarkImplementation() domainStatFunc {
	switch os.Getenv(oldBenchmarkEnvironment) {
	case "old":
		return getDomainStatOldFastJSON
	case "original":
		return getDomainStatOriginal
	default:
		return GetDomainStat
	}
}

func checkBenchmarkResult(b *testing.B, stat DomainStat) {
	b.Helper()
	if stat["quinu.biz"] != 60 {
		b.Fatalf("unexpected quinu.biz count: %d", stat["quinu.biz"])
	}
}

func getDomainStatOldFastJSON(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	targetSuffix := "." + strings.ToLower(domain)
	var parser fastjson.Parser
	for scanner.Scan() {
		user, err := parser.ParseBytes(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("decode user: %w", err)
		}

		emailDomain, ok := normalizeEmailDomainOld(user.GetStringBytes("Email"))
		if !ok {
			continue
		}

		if strings.HasSuffix(emailDomain, targetSuffix) {
			result[emailDomain]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read users: %w", err)
	}
	return result, nil
}

func normalizeEmailDomainOld(email []byte) (string, bool) {
	atIndex := bytes.LastIndexByte(email, '@')
	if atIndex < 0 || atIndex == len(email)-1 {
		return "", false
	}
	return strings.ToLower(string(email[atIndex+1:])), true
}

type oldUsers [100_000]User

func getDomainStatOriginal(r io.Reader, domain string) (DomainStat, error) {
	users, err := getUsersOld(r)
	if err != nil {
		return nil, err
	}
	return countDomainsOld(users, domain)
}

func getUsersOld(r io.Reader) (result oldUsers, err error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return result, err
	}

	lines := strings.Split(string(content), "\n")
	for index, line := range lines {
		var user User
		if err = json.Unmarshal([]byte(line), &user); err != nil {
			return result, err
		}
		result[index] = user
	}
	return result, nil
}

func countDomainsOld(users oldUsers, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range users {
		matched, err := regexp.Match("\\."+domain, []byte(user.Email))
		if err != nil {
			return nil, err
		}

		if matched {
			emailDomain := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
			result[emailDomain]++
		}
	}
	return result, nil
}
