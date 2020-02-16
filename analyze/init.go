package analyze

var (
	memoryBuckets []int
)

func init() {
	for memory := 128; memory <= 3008; memory += 64 {
		memoryBuckets = append(memoryBuckets, memory)
	}
}
