package retry

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type errTemporary struct{}

func (e errTemporary) Error() string   { return "error" }
func (e errTemporary) Temporary() bool { return true }

func TestIfTemporary_RetryCases(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "should retry if err isTemporary",
			err:  &errTemporary{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			count := 0
			sleep = func(d time.Duration) {}
			_ = IfTemporary(ctx, func() error {
				count++
				return test.err
			})
			retried := count > 1
			assert.True(t, retried)
		})
	}
}

func TestIfTemporary_NoRetryCases(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "should not retry if err is not temporary",
			err:  errors.New("whatever"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			count := 0
			sleep = func(d time.Duration) {}
			_ = IfTemporary(ctx, func() error {
				count++
				return test.err
			})
			retried := count > 1
			assert.False(t, retried)
		})
	}
}

func TestIfTemporary_RetryTimes(t *testing.T) {
	tests := []struct {
		name        string
		callResults [3]error
		expected    int
	}{
		{
			name: "should try 3 times by default",
			callResults: [3]error{
				&errTemporary{},
				&errTemporary{},
				&errTemporary{},
			},
			expected: 3,
		},
		{
			name: "should try 2 times if the second call works",
			callResults: [3]error{
				&errTemporary{},
				nil,
				nil,
			},
			expected: 2,
		},
		{
			name: "should only try once if the first call works",
			callResults: [3]error{
				nil,
				nil,
				nil,
			},
			expected: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			count := 0
			sleep = func(d time.Duration) {}
			_ = IfTemporary(ctx, func() error {
				err := test.callResults[count]
				count++
				return err
			})
			assert.Equal(t, test.expected, count)
		})
	}
}

func TestIfTemporary_Sleep(t *testing.T) {
	tests := []struct {
		name           string
		callResults    [3]error
		sleepDurations [3]time.Duration
	}{
		{
			name: "should not sleep in the first try",
		},
		{
			name:           "should sleep for baseTimeout in the second try",
			callResults:    [3]error{&errTemporary{}},
			sleepDurations: [3]time.Duration{defaultRetrier.baseTimeout},
		},
		{
			name: "should sleep for 2*baseTimeout in the third try",
			callResults: [3]error{
				&errTemporary{},
				&errTemporary{},
			},
			sleepDurations: [3]time.Duration{
				defaultRetrier.baseTimeout,
				defaultRetrier.baseTimeout * 2,
			},
		},
		{
			name: "should not sleep after it fails in the last try",
			callResults: [3]error{
				&errTemporary{},
				&errTemporary{},
				&errTemporary{},
			},
			sleepDurations: [3]time.Duration{
				defaultRetrier.baseTimeout,
				defaultRetrier.baseTimeout * 2,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			count := 0
			actualDurations := [3]time.Duration{}
			sleep = func(d time.Duration) {
				actualDurations[count] = d
			}
			_ = IfTemporary(ctx, func() error {
				err := test.callResults[count]
				count++
				return err
			})
			assert.ElementsMatch(t, test.sleepDurations, actualDurations)
		})
	}
}

func TestIfTemporary_Return(t *testing.T) {
	exampleErr := errors.New("whatever")
	tests := []struct {
		name        string
		callResults [3]error
		expected    error
	}{
		{
			name: "should return nil if no error",
		},
		{
			name:        "should return the error when not temporary",
			callResults: [3]error{exampleErr},
			expected:    exampleErr,
		},
		{
			name: "should return the last error when retrying",
			callResults: [3]error{
				&errTemporary{},
				exampleErr,
			},
			expected: exampleErr,
		},
		{
			name: "should return the last error when finished retrying",
			callResults: [3]error{
				&errTemporary{},
				&errTemporary{},
				exampleErr,
			},
			expected: exampleErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			count := 0
			sleep = func(d time.Duration) {}
			actual := IfTemporary(ctx, func() error {
				err := test.callResults[count]
				count++
				return err
			})
			assert.Equal(t, test.expected, actual)
		})
	}
}
