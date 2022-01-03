package wk5

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	// Could be set from configuration
	DurationPerBucket = time.Duration(1 * time.Second)
	MaxBucketNumber   = 10
)

type bucket struct {
	startTime time.Time
	succeeded int64
	failed    int64
	timeout   int64
	rejected  int64

	next *bucket
}

type rollingCounter struct {
	bucketNumber int32
	head         *bucket
	tail         *bucket

	totalSucceeded int64
	totalFailed    int64
	totalTimeout   int64
	totalRejected  int64

	mu sync.Mutex
}

type matrix struct {
	totalSucceeded int64
	totalFailed    int64
	totalTimeout   int64
	totalRejected  int64
}

func NewBucket() *bucket {
	return &bucket{
		startTime: time.Now(),
		succeeded: rand.Int63n(100),
		failed:    rand.Int63n(5),
		timeout:   rand.Int63n(5),
		rejected:  rand.Int63n(10),
	}
}

func NewRollingCounter() *rollingCounter {
	return &rollingCounter{
		bucketNumber: 0,
		head:         nil,
		tail:         nil,
	}
}

func (r *rollingCounter) addBucket(ctx context.Context, b *bucket) (*rollingCounter, error) {
	// The added bucket can be merged with the tail bucket
	if b.startTime.Sub(r.tail.startTime) < DurationPerBucket {
		fmt.Printf("Merge the new bucket with the tail bucket")
		r.mu.Lock()
		n := r.addTotalWithBucket(ctx, b)
		r.mu.Unlock()
		return n, nil
	}

	// The window reaches the maximum bucket number
	if r.bucketNumber == MaxBucketNumber {
		fmt.Printf("The head bucket should be removed")
		r.mu.Lock()
		r.subTotalWithBucket(ctx, r.head)
		h := r.head
		r.head = r.head.next
		h.next = nil

		r.addTotalWithBucket(ctx, b)
		// To keep duration of each bucket to DurationPerBucket
		b.startTime = r.tail.startTime.Add(DurationPerBucket)
		r.tail.next = b
		r.tail = r.tail.next
		r.mu.Unlock()
		fmt.Printf("The new bucket is added to the tail")

		return r, nil
	}

	// bucket number is between [0, MaxBucketNumber)
	r.mu.Lock()
	r.addTotalWithBucket(ctx, b)
	if r.bucketNumber == 0 {
		r.head = b
	} else {
		r.tail.next = b
	}
	r.tail = b
	r.tail.next = nil
	r.bucketNumber = r.bucketNumber + 1
	r.mu.Unlock()

	return r, nil
}

func (r *rollingCounter) addTotalWithBucket(ctx context.Context, b *bucket) *rollingCounter {
	r.totalSucceeded = r.totalSucceeded + b.succeeded
	r.totalFailed = r.totalFailed + b.failed
	r.totalTimeout = r.totalTimeout + b.timeout
	r.totalRejected = r.totalRejected + b.rejected
	return r
}

func (r *rollingCounter) subTotalWithBucket(ctx context.Context, b *bucket) *rollingCounter {
	r.totalSucceeded = r.totalSucceeded - b.succeeded
	r.totalFailed = r.totalFailed - b.failed
	r.totalTimeout = r.totalTimeout - b.timeout
	r.totalRejected = r.totalRejected - b.rejected
	return r
}

func (r *rollingCounter) snapshot() *matrix {
	return &matrix{
		totalSucceeded: r.totalSucceeded,
		totalFailed:    r.totalFailed,
		totalTimeout:   r.totalTimeout,
		totalRejected:  r.totalRejected,
	}
}

func (r *rollingCounter) runOnce(ctx context.Context, tick time.Ticker, in chan bucket, out chan matrix) {
	fmt.Printf("Rolling Counter is started")

	go func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Need to stop input handling and stop output channel")
				close(out)
				return ctx.Err()
			case t := <-tick.C:
				fmt.Println("Tick at", t)
				fmt.Printf("Output the current snapshot/n")
				s := r.snapshot()
				out <- *s
			case b := <-in:
				fmt.Printf("Receving new bucket data")
				newRC, err := r.addBucket(ctx, &b)
				if err != nil {
					fmt.Printf("there is someting wrong, error: %v", err)
					close(out)
					return err
				}
				r = newRC
			}
		}
	}()
}

func main() {
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	g, errCtx := errgroup.WithContext(ctx)

	rc := NewRollingCounter()
	ticker := time.NewTicker(500 * time.Millisecond)
	in := make(chan bucket, MaxBucketNumber)
	out := make(chan matrix, MaxBucketNumber)

	g.Go(func() error {
		fmt.Printf("Printing output matrix")
		select {
		case <-errCtx.Done():
			return errCtx.Err()
		case m := <-out:
			fmt.Printf("Succeed Number: %v/n", m)
		}
		return nil
	})

	g.Go(func() error {
		fmt.Printf("Generating sample bucket")
		for {
			data := NewBucket()

			select {
			case in <- *data:
				time.Sleep(time.Duration(rand.Int63n(500)*int64(time.Millisecond)) + DurationPerBucket)

			case <-errCtx.Done():
				close(in)
				return errCtx.Err()
			}
		}
	})

	rc.runOnce(errCtx, *ticker, in, out)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	g.Go(func() error {
		fmt.Printf("Waiting for OS signal")
		select {
		case <-errCtx.Done():
			return errCtx.Err()
		case <-ch:
			cancelFunc()
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("group wait error '%s'", err.Error())
	}
	close(ch)
}
