package cloudwatch

import (
	"bytes"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Reader is an io.Reader implementation that streams log lines from cloudwatch
// logs.
type Reader struct {
	group, stream, nextToken *string

	client client

	throttle <-chan time.Time

	b lockingBuffer
}

func NewReader(group, stream string, client *cloudwatchlogs.CloudWatchLogs) *Reader {
	r := &Reader{
		group:    aws.String(group),
		stream:   aws.String(stream),
		client:   client,
		throttle: time.Tick(readThrottle),
	}
	go r.start()
	return r
}

func (r *Reader) start() error {
	for {
		<-r.throttle
		if err := r.read(); err != nil {
			return err
		}
	}
}

func (r *Reader) read() error {
	resp, err := r.client.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  r.group,
		LogStreamName: r.stream,
		NextToken:     r.nextToken,
	})
	if err != nil {
		return err
	}

	// We want to re-use the existing token in the event that
	// NextForwardToken is nil, which means there's no new messages to
	// consume.
	if resp.NextForwardToken != nil {
		r.nextToken = resp.NextForwardToken
	}

	// If there are no messages, return so that the consumer can read again.
	if len(resp.Events) == 0 {
		return nil
	}

	for _, event := range resp.Events {
		r.b.WriteString(*event.Message)
	}

	return nil
}

func (r *Reader) Read(b []byte) (int, error) {
	// If there is not data right now, return. Reading from the buffer would
	// result in io.EOF being returned, which is not what we want.
	if r.b.Len() == 0 {
		return 0, nil
	}

	return r.b.Read(b)
}

// lockingBuffer is a bytes.Buffer that locks Reads and Writes.
type lockingBuffer struct {
	sync.Mutex
	bytes.Buffer
}

func (r *lockingBuffer) Read(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.Buffer.Read(b)
}

func (r *lockingBuffer) Write(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.Buffer.Write(b)
}
