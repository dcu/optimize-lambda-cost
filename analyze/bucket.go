package analyze

import (
	"fmt"
	"math"
	"sort"

	"github.com/valyala/histogram"
)

// Bucket has the information for a collection of requests associated to the same memory size
type Bucket struct {
	Size                  int64
	Count                 int
	DurationHist          *histogram.Fast
	MemoryHist            *histogram.Fast
	CountByBilledDuration map[int64]int64
}

func newBucket(size int64) *Bucket {
	return &Bucket{
		Size:                  size,
		DurationHist:          histogram.NewFast(),
		MemoryHist:            histogram.NewFast(),
		CountByBilledDuration: map[int64]int64{},
	}
}

func (b *Bucket) update(duration float64, memoryUsed float64, billedDuration int64) {
	b.Count++
	b.DurationHist.Update(duration)
	b.MemoryHist.Update(memoryUsed)
	b.CountByBilledDuration[billedDuration]++
}

// CalculateSuggestedMemory returns a suggestion for optimizing the given percentile
func (b *Bucket) CalculateSuggestedMemory(percentile float64) int {
	duration := b.DurationHist.Quantile(percentile)
	targetDurationPos, _ := findBilledDuration(duration)

	currentMemoryPos, _ := findMemoryIndex(int(b.Size)) // memory matters because it's usually proportional to current duration if the task is cpu bounded

	avgPos := (currentMemoryPos + targetDurationPos) / 2
	if avgPos >= len(memoryBuckets) {
		avgPos = len(memoryBuckets) - 1
	}

	return memoryBuckets[avgPos]
}

// Print information associated with the bucket
func (b *Bucket) Print() {
	fmt.Println(">> Analyzing stats for memory bucket:", b.Size, "(total requests:", b.Count, ")")

	fmt.Println("> Top requests per billed duration")
	billedDurations := make([][]int64, 0, len(b.CountByBilledDuration))
	for billedDuration, count := range b.CountByBilledDuration {
		billedDurations = append(billedDurations, []int64{billedDuration, count})
	}

	maxCount := int64(0)
	sort.Slice(billedDurations, func(i, j int) bool {
		if billedDurations[i][1] > maxCount {
			maxCount = billedDurations[i][1]
		}

		return billedDurations[i][0] < billedDurations[j][0]
	})

	for _, billedDuration := range billedDurations {
		if billedDuration[1] > int64(float64(maxCount)*0.1) {
			fmt.Printf("%d ms: %d (%0.2f%%)\n", billedDuration[0], billedDuration[1], (float64(billedDuration[1])/float64(b.Count))*100)
		}
	}

	fmt.Println("")
	fmt.Println("> Distribution for durations")
	p1duration := b.DurationHist.Quantile(0.01)
	p25duration := b.DurationHist.Quantile(0.25)
	p50duration := b.DurationHist.Quantile(0.50)
	p75duration := b.DurationHist.Quantile(0.75)
	p99duration := b.DurationHist.Quantile(0.99)
	maxDuration := b.DurationHist.Quantile(1)

	printDurationPercentile("1th", p1duration)
	printDurationPercentile("25th", p25duration)
	printDurationPercentile("50th", p50duration)
	printDurationPercentile("75th", p75duration)
	printDurationPercentile("99th", p99duration)
	printDurationPercentile("max", maxDuration)

	fmt.Println("")
	fmt.Println("> Distribution for used memory")
	p1memory := b.MemoryHist.Quantile(0.01)
	p25memory := b.MemoryHist.Quantile(0.25)
	p50memory := b.MemoryHist.Quantile(0.50)
	p75memory := b.MemoryHist.Quantile(0.75)
	p99memory := b.MemoryHist.Quantile(0.99)
	maxMemory := b.MemoryHist.Quantile(1)

	fmt.Println("1th ", p1memory, "MB")
	fmt.Println("25th", p25memory, "MB")
	fmt.Println("50th", p50memory, "MB")
	fmt.Println("75th", p75memory, "MB")
	fmt.Println("99th", p99memory, "MB")
	fmt.Println("max", maxMemory, "MB")

	fmt.Println("")
	fmt.Println("> Suggested memory based on your usage")

	fmt.Printf("Suggestion for 1th percentile: %d MB\n", b.CalculateSuggestedMemory(0.01))
	fmt.Printf("Suggestion for 25th percentile: %d MB\n", b.CalculateSuggestedMemory(0.25))
	fmt.Printf("Suggestion for 75th percentile: %d MB\n", b.CalculateSuggestedMemory(0.75))
	fmt.Printf("Suggestion for 99th percentile: %d MB\n", b.CalculateSuggestedMemory(0.99))

	fmt.Println("")
}

func printDurationPercentile(label string, value float64) {
	_, billed := findBilledDuration(value)
	fmt.Println(label, value, "ms", "billed:", billed, "ms")
}

func findMemoryIndex(usedMemory int) (int, int) {
	for i, memory := range memoryBuckets {
		if memory >= usedMemory {
			return i + 1, memory
		}
	}

	return 0, 128
}

func findBilledDuration(duration float64) (int, int) {
	incSize := 100
	durI := int(math.Ceil(duration))
	index := 0
	for dur := incSize; dur <= 900000; dur += incSize {
		if dur > durI {
			return index, dur
		}

		index++
	}

	return 0, incSize
}
