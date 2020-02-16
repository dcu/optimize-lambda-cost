package analyze

import (
	"log"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var (
	lambdaReqRx = regexp.MustCompile(`REPORT RequestId: ([\w-]+)\tDuration: (\d+\.\d+) ms\tBilled Duration: (\d+) ms\tMemory Size: (\d+) MB\tMax Memory Used: (\d+) MB`)
)

// EventParser is in charge of parsing the lambda logs
type EventParser struct {
	buckets map[int64]*Bucket
}

func newEventParser() *EventParser {
	return &EventParser{
		buckets: map[int64]*Bucket{},
	}
}

func (p *EventParser) parseEvents(events []*cloudwatchlogs.FilteredLogEvent) {
	for _, event := range events {
		p.parseEvent(*event.Message)
	}
}

func (p *EventParser) parseEvent(message string) {
	matches := lambdaReqRx.FindStringSubmatch(message)
	if len(matches) == 0 {
		log.Println("couldn't parse:", message)
		return
	}

	_, duration, billedDuration, memory, memoryUsed := matches[1], matches[2], matches[3], matches[4], matches[5]

	memoryI, _ := strconv.ParseInt(memory, 10, 64)

	bucket, ok := p.buckets[memoryI]
	if !ok {
		bucket = newBucket(memoryI)
		p.buckets[memoryI] = bucket
	}

	durationF, _ := strconv.ParseFloat(duration, 64)
	memoryUsedF, _ := strconv.ParseFloat(memoryUsed, 64)
	billedDurationI, _ := strconv.ParseInt(billedDuration, 10, 64)

	bucket.update(durationF, memoryUsedF, billedDurationI)
}
