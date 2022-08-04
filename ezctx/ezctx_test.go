package ctx

import (
	"reflect"
	"testing"
	"time"

	"github.com/ezotaka/golib/channel/ctxpl"
)

func TestWithDone(t *testing.T) {
	closed := make(chan struct{})
	close(closed)
	type args struct {
		done chan struct{}
	}
	tests := []struct {
		name      string
		args      args
		doneCount int
		want      []int
	}{
		{
			name: "normal done",
			args: args{
				done: make(chan struct{}),
			},
			doneCount: 2,
			want:      []int{1, 1},
		},
		{
			name: "already done",
			args: args{
				done: closed,
			},
			want: []int{},
		},
		{
			name: "nil done",
			args: args{
				done: nil,
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithDone(tt.args.done)
			c := ctxpl.Repeat(ctx, 1)
			got := []int{}
			for v := range c {
				got = append(got, v)
				if tt.doneCount == len(got) {
					close(tt.args.done)
					time.Sleep(100 * time.Millisecond)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDone() = %v, want %v", got, tt.want)
			}
		})
	}
}
