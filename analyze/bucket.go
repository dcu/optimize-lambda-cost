package analyze

import (
	"fmt"
	"io"
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

	avgPos := targetDurationPos + (currentMemoryPos / 2)
	if avgPos >= len(memoryBuckets) {
		avgPos = len(memoryBuckets) - 1
	}

	return memoryBuckets[avgPos]
}

// Print information associated with the bucket
func (b *Bucket) Print(output io.Writer) {
	percentiles := []float64{
		0.01, 0.25, 0.50, 0.75, 0.99,
	}

	fmt.Fprintln(output, ">> Analyzing stats for memory bucket:", b.Size, "MB (total requests:", b.Count, ")")

	fmt.Fprintln(output, "> Top requests per billed duration")
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

	estimatedCostPerMillion := 0.20
	for _, billedDuration := range billedDurations {
		percent := (float64(billedDuration[1]) / float64(b.Count))
		if billedDuration[1] > int64(float64(maxCount)*0.1) {
			fmt.Fprintf(output, "%d ms: %d (%0.2f%%)\n", billedDuration[0], billedDuration[1], percent*100)
		}

		totalRequests := 1_000_000.0 * percent
		units := billedDuration[0] / 100

		estimatedCostPerMillion += costMapping[b.Size] * float64(totalRequests) * float64(units)
	}

	fmt.Fprintf(output, "Estimated cost per million requests: %0.2f$\n", estimatedCostPerMillion)

	fmt.Fprintln(output, "")
	fmt.Fprintln(output, "> Distribution for durations")
	for _, percentile := range percentiles {
		pDuration := b.DurationHist.Quantile(percentile)
		printDurationPercentile(output, fmt.Sprintf("%dth percentile", int(percentile*100)), pDuration)
	}
	fmt.Fprintln(output, "")

	fmt.Fprintln(output, "> Distribution for used memory")
	for _, percentile := range percentiles {
		pMemory := b.MemoryHist.Quantile(percentile)
		fmt.Fprintf(output, "%dth percentile: %0.1f MB\n", int(percentile*100), pMemory)
	}
	fmt.Fprintln(output, "")

	fmt.Fprintln(output, "> Suggested memory based on your usage")
	for _, percentile := range percentiles {
		fmt.Fprintf(output, "Suggestion for %dth percentile: %d MB\n", int(percentile*100), b.CalculateSuggestedMemory(percentile))
	}

	fmt.Fprintln(output, "")
}

func printDurationPercentile(output io.Writer, label string, value float64) {
	_, billed := findBilledDuration(value)
	fmt.Fprintln(output, label, value, "ms", "billed:", billed, "ms")
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
