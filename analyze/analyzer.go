package analyze

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

const maxFetches = 100

// Analyzer is in charge of analyzing the lambda logs
type Analyzer struct {
	client cloudwatchlogsiface.CloudWatchLogsAPI
}

// New creates a new instance of an Analyzer
func New(awsProfile string) *Analyzer {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	if awsProfile != "" {
		opts.Profile = awsProfile
	}

	sess := session.Must(session.NewSessionWithOptions(opts))
	svc := cloudwatchlogs.New(sess)

	return &Analyzer{
		client: svc,
	}
}

// FetchBuckets fetches the memory buckets found in logs
func (a *Analyzer) FetchBuckets(functionName string, startTime time.Time, timeout time.Duration) (map[int64]*Bucket, error) {
	req := &cloudwatchlogs.FilterLogEventsInput{
		StartTime:     aws.Int64(startTime.UnixNano() / int64(time.Millisecond)),
		LogGroupName:  aws.String("/aws/lambda/" + functionName),
		FilterPattern: aws.String("REPORT RequestId"), // https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html
	}

	timeoutCh := time.After(timeout)

	parser := newEventParser()

loop:
	for i := 0; i < maxFetches; i++ {
		select {
		case <-timeoutCh:
			log.Println("fetching logs timed out")
			break loop
		default:
			resp, err := a.client.FilterLogEvents(req)
			if isThrottle(err) {
				fmt.Print("T")
				time.Sleep(1 * time.Second)
				continue
			}

			if err != nil {
				return nil, err
			}

			parser.parseEvents(resp.Events)

			if resp.NextToken == nil {
				break
			}

			req.NextToken = resp.NextToken
			fmt.Print(".")
		}
	}
	fmt.Println("")

	return parser.buckets, nil
}

func isThrottle(err error) bool {
	var awsErr awserr.Error
	if ok := errors.As(err, &awsErr); ok && awsErr.Code() == "ThrottlingException" {
		return true
	}

	return false
}
