# Analyze and optimize AWS Lambda costs

This is a little tool that parses the Lambda Function logs to and
suggests changes in memory usage.

The suggestion made by this tool are not 100% accurate since tasks that
are IO bound can't be optimized by adding more CPU power.

## Installation

```
go get github.com/dcu/optimize-lambda-cost
```

## Usage

```
optimize-lambda-cost --profile="aws-profile-with-read-access-to-cloudwatch-logs" --since="20 hours ago" my-lambda-function
```

And the response would look like:

```
2020/02/16 15:53:20 Start fetching logs for my-lambda-function starting on 2020-02-16 05:53:20.461067 -0500 -05 m=-35999.993904283
.......................
>> Analyzing stats for memory bucket: 192 (total requests: 31273 )
> Distribution for durations
1th 6.86 ms billed: 100 ms
25th 57.12 ms billed: 100 ms
50th 104.11 ms billed: 200 ms
75th 178.52 ms billed: 200 ms
99th 605.36 ms billed: 700 ms
max 1952.67 ms billed: 2000 ms

> Distribution for used memory
1th  64 MB
25th 74 MB
50th 76 MB
75th 78 MB
99th 85 MB
max 85 MB

> Suggested memory based on your usage
Suggestion for 1th percentile: 192 MB
Suggestion for 25th percentile: 192 MB
Suggestion for 75th percentile: 192 MB
Suggestion for 99th percentile: 384 MB
```

Please note that 1th percentile means 99% of the requests and 99th
percentile means the 1% of the requests.

