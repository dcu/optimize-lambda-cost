package analyze

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/require"
)

func Test_parseEvents(t *testing.T) {
	c := require.New(t)

	ep := newEventParser()
	ep.parseEvents([]*cloudwatchlogs.FilteredLogEvent{
		&cloudwatchlogs.FilteredLogEvent{
			Message: aws.String("REPORT RequestId: 0047a237-bc83-4e23-a034-6c7ab29baec2\tDuration: 2337.93 ms\tBilled Duration: 2400 ms\tMemory Size: 128 MB\tMax Memory Used: 64 MB\tInit Duration: 137.16 ms\tXRAY TraceId: 1-5e497526-830095f2b1d59a26d53bd30c\tSegmentId: 524edb79422ff8b3\tSampled: true"),
		},
		&cloudwatchlogs.FilteredLogEvent{
			Message: aws.String("REPORT RequestId: 76557f6b-e145-4ae6-89eb-c9c5810faca4\tDuration: 274.08 ms\tBilled Duration: 300 ms\tMemory Size: 128 MB\tMax Memory Used: 65 MB\tXRAY TraceId: 1-5e49754a-52385175ad4a3f50011ca95d\tSegmentId: 41a89d5e1e9bac01\tSampled: true"),
		},
		&cloudwatchlogs.FilteredLogEvent{
			Message: aws.String("REPORT RequestId: 3be36555-dcf5-468a-8a30-01aa72a69697\tDuration: 909.48 ms\tBilled Duration: 1000 ms\tMemory Size: 128 MB\tMax Memory Used: 66 MB\tXRAY TraceId: 1-5e49756a-99cfc8fe07661c39f686dea2\tSegmentId: 16b47763549375a3 Sampled: true"),
		},
		&cloudwatchlogs.FilteredLogEvent{
			Message: aws.String("REPORT RequestId: cc4b1ff6-5f22-46d2-86ee-0a6ae5ac929d\tDuration: 768.76 ms\tBilled Duration: 800 ms\tMemory Size: 128 MB\tMax Memory Used: 66 MB\tXRAY TraceId: 1-5e49756c-1794006a027481d541643d61\tSegmentId: 746de3434e670005\tSampled: true"),
		},
	})

	c.Equal(1, len(ep.buckets))

	bucket := ep.buckets[128]
	c.NotNil(bucket)

	bucket.Print()
	c.Equal(192, bucket.CalculateSuggestedMemory(0.01))
	c.Equal(896, bucket.CalculateSuggestedMemory(0.99))
}
