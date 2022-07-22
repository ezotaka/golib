package chantest

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestContextWithCountCancel(t *testing.T) {
	type args struct {
		cnt int
	}
	tests := []struct {
		name     string
		args     args
		want     int
		isPanic  bool
		panicMsg string
	}{
		{
			name: "cnt = 0",
			args: args{
				cnt: 0,
			},
			want: 0,
		},
		{
			name: "cnt = 1",
			args: args{
				cnt: 1,
			},
			want: 1,
		},
		{
			name: "cnt = -1",
			args: args{
				cnt: -1,
			},
			isPanic:  true,
			panicMsg: "cnt must be zero or positive",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); tt.isPanic && r == nil {
					t.Errorf("ContextWithCountCancel() doesn't panic, want '%v'", tt.panicMsg)
				} else if tt.isPanic && r != tt.panicMsg {
					t.Errorf("ContextWithCountCancel() panic '%v', want '%v'", r, tt.panicMsg)
				}
			}()
			got := ContextWithCountCancel(tt.args.cnt)
			if cnt, ok := countToCancel(got); !ok || !reflect.DeepEqual(cnt, tt.want) {
				t.Errorf("ContextWithCountCancel() got = %v, want %v", cnt, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	// count up every 100 msec until end
	counter := func(end int) <-chan int {
		valChan := make(chan int)
		go func() {
			defer close(valChan)
			count := 1
			for {
				time.Sleep(100 * time.Millisecond)
				valChan <- count
				if count >= end {
					return
				}
				count++
			}
		}()
		return valChan
	}
	counterCaller := func(ctx context.Context, end int) <-chan int {
		return counter(end)
	}
	type args struct {
		ctx    context.Context
		fn     func(context.Context, int) <-chan int
		fnArgs int
	}
	tests := []struct {
		name     string
		args     args
		want     []int
		isPanic  bool
		panicMsg string
	}{
		{
			name: "channel closed normally",
			args: args{
				ctx:    context.Background(),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by context with count 0",
			args: args{
				ctx:    ContextWithCountCancel(0),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{},
		},
		{
			name: "channel closed by context with count 1",
			args: args{
				ctx:    ContextWithCountCancel(1),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1},
		},
		{
			name: "channel closed by context with count 2",
			args: args{
				ctx:    ContextWithCountCancel(2),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by context with count 3",
			args: args{
				ctx:    ContextWithCountCancel(3),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by timeout 0",
			args: args{
				ctx:    ContextWithTimeout(50 * time.Millisecond),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{},
		},
		{
			name: "channel closed by timeout 1",
			args: args{
				ctx:    ContextWithTimeout(150 * time.Millisecond),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1},
		},
		{
			name: "channel closed by timeout 2",
			args: args{
				ctx:    ContextWithTimeout(250 * time.Millisecond),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "nil context is treated as context.Background()",
			args: args{
				ctx:    nil,
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "nil fn causes panic",
			args: args{
				ctx: context.Background(),
				fn:  nil,
			},
			isPanic:  true,
			panicMsg: "fn must not be nil",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); tt.isPanic && r == nil {
					t.Errorf("ForTest() doesn't panic, want '%v'", tt.panicMsg)
				} else if tt.isPanic && r != tt.panicMsg {
					t.Errorf("ForTest() panic '%v', want '%v'", r, tt.panicMsg)
				}
			}()

			got := Run(tt.args.ctx, tt.args.fn, tt.args.fnArgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ForTest() = %v, want %v", got, tt.want)
			}
		})
	}
}
