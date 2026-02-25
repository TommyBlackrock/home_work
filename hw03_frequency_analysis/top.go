package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

const maxTopTerms = 10

type termFrequency struct {
	term string
	freq int
}

func Top10(text string) []string {
	tokens := tokenizeByWhitespace(text)
	if len(tokens) == 0 {
		return []string{}
	}

	frequencyByTerm := countTermFrequency(tokens)
	rankedTerms := rankTermFrequency(frequencyByTerm)

	return takeTopTerms(rankedTerms, maxTopTerms)
}

func tokenizeByWhitespace(text string) []string {
	return strings.Fields(text)
}

func countTermFrequency(tokens []string) map[string]int {
	frequencyByTerm := make(map[string]int, len(tokens))
	for _, token := range tokens {
		frequencyByTerm[token]++
	}
	return frequencyByTerm
}

func rankTermFrequency(frequencyByTerm map[string]int) []termFrequency {
	rankedTerms := make([]termFrequency, 0, len(frequencyByTerm))
	for term, frequency := range frequencyByTerm {
		rankedTerms = append(rankedTerms, termFrequency{term, frequency})
	}

	sort.Slice(rankedTerms, func(i, j int) bool {
		if rankedTerms[i].freq == rankedTerms[j].freq {
			return rankedTerms[i].term < rankedTerms[j].term
		}
		return rankedTerms[i].freq > rankedTerms[j].freq
	})
	return rankedTerms
}

func takeTopTerms(rankedTerms []termFrequency, limit int) []string {
	topN := minInt(limit, len(rankedTerms))
	result := make([]string, 0, topN)
	for _, tf := range rankedTerms[:topN] {
		result = append(result, tf.term)
	}

	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
