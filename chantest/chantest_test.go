package chantest

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ezotaka/golib/channel"
)

func TestContextWithCountCancel(t *testing.T) {
	type args struct {
		parent context.Context
		cnt    int
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
				parent: context.Background(),
				cnt:    0,
			},
			want: 0,
		},
		{
			name: "cnt = 1",
			args: args{
				parent: context.Background(),
				cnt:    1,
			},
			want: 1,
		},
		{
			name: "nil parent",
			args: args{
				parent: nil,
				cnt:    0,
			},
			isPanic:  true,
			panicMsg: "cannot create context from nil parent",
		},
		{
			name: "cnt = -1",
			args: args{
				parent: context.Background(),
				cnt:    -1,
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
			got, _ := ContextWithCountCancel(tt.args.parent, tt.args.cnt)
			if cnt, ok := countToCancel(got); !ok || !reflect.DeepEqual(cnt, tt.want) {
				t.Errorf("ContextWithCountCancel() got = %v, want %v", cnt, tt.want)
			}
		})
	}
}

func TestForTest(t *testing.T) {
	cancelByCount := func(cnt int) context.Context {
		ctx, _ := ContextWithCountCancel(context.Background(), cnt)
		return ctx
	}
	type args struct {
		ctx context.Context
		c   <-chan int
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
				ctx: context.Background(),
				c:   channel.ToChan(1, 2),
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by context with count 0",
			args: args{
				ctx: cancelByCount(0),
				c:   channel.ToChan(1, 2),
			},
			want: []int{},
		},
		{
			name: "channel closed by context with count 1",
			args: args{
				ctx: cancelByCount(1),
				c:   channel.ToChan(1, 2),
			},
			want: []int{1},
		},
		{
			name: "channel closed by context with count 2",
			args: args{
				ctx: cancelByCount(2),
				c:   channel.ToChan(1, 2),
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by context with count 3",
			args: args{
				ctx: cancelByCount(3),
				c:   channel.ToChan(1, 2),
			},
			want: []int{1, 2},
		},
		{
			name: "nil context",
			args: args{
				ctx: nil,
			},
			isPanic:  true,
			panicMsg: "ctx must not be nil",
		},
		{
			name: "nil channel",
			args: args{
				ctx: context.Background(),
				c:   nil,
			},
			want: nil,
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

			got := ForTest(tt.args.ctx, tt.args.c)
			gotSlice := channel.ToSlice(context.Background(), got)
			if !reflect.DeepEqual(gotSlice, tt.want) {
				t.Errorf("ForTest() = %v, want %v", gotSlice, tt.want)
			}
		})
	}
}

func TestGetTestedValues(t *testing.T) {
	// function to be tested
	// count up every 100 msec until end or done
	counter := func(ctx context.Context, end int) <-chan int {
		valChan := make(chan int)
		go func() {
			defer close(valChan)
			count := 1
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					valChan <- count
					if count >= end {
						return
					}
					count++
				}
			}
		}()
		return valChan
	}
	type counterArgs struct {
		end int
	}

	type args struct {
		test   Case[int, counterArgs]
		caller func(context.Context, counterArgs) <-chan int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "not done, end 3",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "already done at first",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 10,
					},
					IsDoneAtFirst: true,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{},
		},
		{
			name: "done by index == 0",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 0 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by index == 1",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 1 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2},
		},
		{
			name: "not satisfied index == 5",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 5 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "done by value == 1",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 1 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by value == 2",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 2 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2},
		},

		{
			name: "not satisfied value == 5",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 5 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "done by time == 50 ms",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByTime: 50 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{},
		},
		{
			name: "done by time == 250 ms",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByTime: 250 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2},
		},
		{
			name: "not satisfied time == 500 ms",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByTime: 500 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "done by index and value",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 1 },
					IsDoneByValue: func(v int) bool { return v == 2 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1, 2},
		},
		{
			name: "done by index before value",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 0 },
					IsDoneByValue: func(v int) bool { return v == 2 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by value before index",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 2 },
					IsDoneByValue: func(v int) bool { return v == 1 },
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by index before time",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 0 },
					IsDoneByTime:  250 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by time before index",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 2 },
					IsDoneByTime:  150 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by value before time",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 1 },
					IsDoneByTime:  250 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by time before value",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 2 },
					IsDoneByTime:  150 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "all conditions",
			args: args{
				test: Case[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneAtFirst: true,
					IsDoneByIndex: func(i int) bool { return i == 2 },
					IsDoneByValue: func(v int) bool { return v == 2 },
					IsDoneByTime:  250 * time.Millisecond,
				},
				caller: func(ctx context.Context, a counterArgs) <-chan int {
					return counter(ctx, a.end)
				},
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetTestedValues(tt.args.test, tt.args.caller); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DoTest() = %v, want %v", got, tt.want)
			}
		})
	}
}
