package hw10programoptimization

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/valyala/fastjson" //nolint:depguard // Homework permits third-party JSON parsers.
)

const (
	// 4pcs of workers.
	domainStatWorkers = 4

	// 512 lines pack.
	linesPerBatch = 512

	// 8 reusable buffers so that the reader and worker can operate in parallel without allocating
	// a new []byte for every string.
	batchPoolSize       = domainStatWorkers * 2
	batchInitialSize    = 128 * 1024
	scannerInitialSize  = 4 * 1024
	scannerMaxTokenSize = 1024 * 1024
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

// lineBatch stores few JSON-strings in one continuous buffer.
// firstLine for error message purpose.
type lineBatch struct {
	data      []byte
	firstLine int
	lineCount int
}

// workerResult for specific worker.
type workerResult struct {
	stat DomainStat
	err  *lineError
}

type lineError struct {
	line int
	err  error
}

// domainCounter internals: key strings are created only on first occurrence
// subsequent addresses increment the counter by index without a new allocation.
type domainCounter struct {
	index   map[string]int
	keys    []string
	counts  []int
	scratch []byte
}

func newDomainCounter() *domainCounter {
	const expectedDomainsPerWorker = 128

	return &domainCounter{
		index:  make(map[string]int, expectedDomainsPerWorker),
		keys:   make([]string, 0, expectedDomainsPerWorker),
		counts: make([]int, 0, expectedDomainsPerWorker),
	}
}

// GetDomainStat reads JSON Lines in a streaming manner and counts email domains ending with the given top-level domain.

// Processing is structured as a pipeline.
func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	targetSuffix := []byte("." + strings.ToLower(domain))
	// Scanner reads lines and collects them into reusable batches.
	jobs := make(chan *lineBatch, batchPoolSize)
	freeBatches := make(chan *lineBatch, batchPoolSize)
	for i := 0; i < batchPoolSize; i++ {
		freeBatches <- &lineBatch{
			data: make([]byte, 0, batchInitialSize),
		}
	}

	results := make([]workerResult, domainStatWorkers)
	var workers sync.WaitGroup
	workers.Add(domainStatWorkers)

	for workerID := 0; workerID < domainStatWorkers; workerID++ {
		go func() {
			defer workers.Done()
			runDomainWorker(jobs, freeBatches, targetSuffix, &results[workerID])
		}()
	}

	readErr := produceBatches(r, jobs, freeBatches)
	close(jobs)
	workers.Wait()

	if readErr != nil {
		return nil, fmt.Errorf("read users: %w", readErr)
	}
	if validationErr := firstLineError(results); validationErr != nil {
		return nil, fmt.Errorf("decode user at line %d: %w", validationErr.line, validationErr.err)
	}
	return mergeWorkerStats(results), nil
}

// produceBatches copies data from the Scanner into its own batch buffers
// сopying is mandatory because the Scanner reuses memory after the next Scan call.
func produceBatches(
	r io.Reader,
	jobs chan<- *lineBatch,
	freeBatches <-chan *lineBatch,
) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, scannerInitialSize), scannerMaxTokenSize)

	batch := <-freeBatches
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		if batch.lineCount == 0 {
			batch.firstLine = lineNumber
		}

		batch.data = append(batch.data, scanner.Bytes()...)
		batch.data = append(batch.data, '\n')
		batch.lineCount++

		if batch.lineCount == linesPerBatch {
			jobs <- batch
			batch = <-freeBatches
		}
	}

	if batch.lineCount > 0 {
		jobs <- batch
	} else {
		resetBatch(batch)
	}

	return scanner.Err()
}

func runDomainWorker(
	jobs <-chan *lineBatch,
	freeBatches chan<- *lineBatch,
	targetSuffix []byte,
	result *workerResult,
) {
	counter := newDomainCounter()
	var parser fastjson.Parser

	for batch := range jobs {
		if err := processBatch(batch, targetSuffix, &parser, counter); err != nil {
			if result.err == nil || err.line < result.err.line {
				result.err = err
			}
		}

		resetBatch(batch)
		freeBatches <- batch
	}

	result.stat = counter.toDomainStat()
}

func processBatch(
	batch *lineBatch,
	targetSuffix []byte,
	parser *fastjson.Parser,
	counter *domainCounter,
) *lineError {
	data := batch.data
	lineNumber := batch.firstLine

	for offset := 0; offset < batch.lineCount; offset++ {
		newline := bytes.IndexByte(data, '\n')
		if newline < 0 {
			return &lineError{
				line: lineNumber,
				err:  fmt.Errorf("internal batch format: newline not found"),
			}
		}

		line := data[:newline]
		data = data[newline+1:]

		user, err := parser.ParseBytes(line)
		if err != nil {
			return &lineError{line: lineNumber, err: err}
		}

		emailDomain, ok := extractEmailDomain(user.GetStringBytes("Email"))
		if ok && hasDomainSuffix(emailDomain, targetSuffix) {
			counter.add(emailDomain)
		}

		lineNumber++
	}

	return nil
}

func extractEmailDomain(email []byte) ([]byte, bool) {
	atIndex := bytes.LastIndexByte(email, '@')
	if atIndex < 0 || atIndex == len(email)-1 {
		return nil, false
	}

	return email[atIndex+1:], true
}

// hasDomainSuffix checks the TLD directly in the []byte. No domain string is created.
func hasDomainSuffix(domain, targetSuffix []byte) bool {
	if len(domain) < len(targetSuffix) {
		return false
	}

	return bytes.EqualFold(domain[len(domain)-len(targetSuffix):], targetSuffix)
}

func (c *domainCounter) add(domain []byte) {
	c.scratch = append(c.scratch[:0], domain...)

	ascii := true
	for i, char := range c.scratch {
		switch {
		case char >= 'A' && char <= 'Z':
			c.scratch[i] = char + ('a' - 'A')
		case char >= utf8.RuneSelf:
			ascii = false
		}
	}

	if !ascii {
		c.addNormalized(strings.ToLower(string(domain)))
		return
	}

	// Go compiler converts []byte to string when reading from the map without storing a temporary string
	if index, ok := c.index[string(c.scratch)]; ok {
		c.counts[index]++
		return
	}

	key := string(c.scratch)
	c.index[key] = len(c.keys)
	c.keys = append(c.keys, key)
	c.counts = append(c.counts, 1)
}

func (c *domainCounter) addNormalized(domain string) {
	if index, ok := c.index[domain]; ok {
		c.counts[index]++
		return
	}

	c.index[domain] = len(c.keys)
	c.keys = append(c.keys, domain)
	c.counts = append(c.counts, 1)
}

func (c *domainCounter) toDomainStat() DomainStat {
	result := make(DomainStat, len(c.keys))
	for index, key := range c.keys {
		result[key] = c.counts[index]
	}
	return result
}

func firstLineError(results []workerResult) *lineError {
	var first *lineError
	for index := range results {
		current := results[index].err
		if current != nil && (first == nil || current.line < first.line) {
			first = current
		}
	}
	return first
}

func mergeWorkerStats(results []workerResult) DomainStat {
	totalDomains := 0
	for index := range results {
		totalDomains += len(results[index].stat)
	}

	merged := make(DomainStat, totalDomains)
	for index := range results {
		for domain, count := range results[index].stat {
			merged[domain] += count
		}
	}
	return merged
}

func resetBatch(batch *lineBatch) {
	batch.data = batch.data[:0]
	batch.firstLine = 0
	batch.lineCount = 0
}
