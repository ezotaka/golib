package channel

import (
	"reflect"
	"testing"
	"time"
)

func TestGetTestedValues(t *testing.T) {
	// function to be tested
	// count up every 100 msec until end or done
	counter := func(done <-chan any, end int) <-chan int {
		valChan := make(chan int)
		go func() {
			defer close(valChan)
			count := 1
			for {
				select {
				case <-done:
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
		test   TestCase[int, counterArgs]
		caller func(<-chan any, counterArgs) <-chan int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "not done, end 3",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "already done at first",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 10,
					},
					IsDoneAtFirst: true,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{},
		},
		{
			name: "done by index == 0",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 0 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by index == 1",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 1 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2},
		},
		{
			name: "not satisfied index == 5",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 5 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "done by value == 1",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 1 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by value == 2",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 2 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2},
		},

		{
			name: "not satisfied value == 5",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 5 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "done by time == 50 ms",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByTime: 50 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{},
		},
		{
			name: "done by time == 250 ms",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByTime: 250 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2},
		},
		{
			name: "not satisfied time == 500 ms",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByTime: 500 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2, 3},
		},
		{
			name: "done by index and value",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 1 },
					IsDoneByValue: func(v int) bool { return v == 2 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1, 2},
		},
		{
			name: "done by index before value",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 0 },
					IsDoneByValue: func(v int) bool { return v == 2 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by value before index",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 2 },
					IsDoneByValue: func(v int) bool { return v == 1 },
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by index before time",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 0 },
					IsDoneByTime:  250 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by time before index",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByIndex: func(i int) bool { return i == 2 },
					IsDoneByTime:  150 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by value before time",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 1 },
					IsDoneByTime:  250 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "done by time before value",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneByValue: func(v int) bool { return v == 2 },
					IsDoneByTime:  150 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{1},
		},
		{
			name: "all conditions",
			args: args{
				test: TestCase[int, counterArgs]{
					Args: counterArgs{
						end: 3,
					},
					IsDoneAtFirst: true,
					IsDoneByIndex: func(i int) bool { return i == 2 },
					IsDoneByValue: func(v int) bool { return v == 2 },
					IsDoneByTime:  250 * time.Millisecond,
				},
				caller: func(done <-chan any, a counterArgs) <-chan int {
					return counter(done, a.end)
				},
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTestedValues(tt.args.test, tt.args.caller); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DoTest() = %v, want %v", got, tt.want)
			}
		})
	}
}
