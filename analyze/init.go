package analyze

var (
	memoryBuckets []int
	costMapping   = map[int64]float64{}
)

func init() {
	i := 1
	for memory := 128; memory <= 3008; memory += 64 {
		memoryBuckets = append(memoryBuckets, memory)

		costMapping[int64(memory)] = 0.000000208 * float64(i)

		i++
	}
}
