package services

import (
	"math"
	"sort"
)

type WordStat struct {
	Word  string
	TF    float64
	Count int
	IDF   float64
	TFIDF float64
}

func ComputeTFIDF(words []string) []WordStat {
	wordCount := make(map[string]int)
	for _, w := range words {
		wordCount[w]++
	}

	totalWords := len(words)

	stats := make([]WordStat, 0, len(wordCount))
	for w, count := range wordCount {
		tf := float64(count) / float64(totalWords)

		idf := math.Log(float64(totalWords) / float64(count))

		stats = append(stats, WordStat{
			Word:  w,
			TF:    tf,
			Count: count,
			IDF:   idf,
			TFIDF: tf * idf,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TF > stats[j].TF
	})

	return stats
}
