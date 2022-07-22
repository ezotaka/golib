package conv

import (
	"context"
	"reflect"
	"testing"

	"github.com/ezotaka/golib/chantest"
)

func TestToChan(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []chantest.Case[int, args]{
		{
			Name: "length=0",
			Args: args{
				values: []int{},
			},
			Want: []int{},
		},
		{
			Name: "length=1",
			Args: args{
				values: []int{1},
			},
			Want: []int{1},
		},
		{
			Name: "length=2",
			Args: args{
				values: []int{1, 2},
			},
			Want: []int{1, 2},
		},
	}
	caller := func(ctx context.Context, a args) <-chan int {
		return ToChan(a.values...)
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := chantest.GetTestedValues(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("ToChan() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestToSlice(t *testing.T) {
	type args struct {
		ctx context.Context
		c   <-chan int
	}
	tests := []chantest.Case[int, args]{
		{
			Name: "{1, 2}",
			Args: args{
				ctx: context.Background(),
				c:   ToChan(1, 2),
			},
			Want: []int{1, 2},
		},
		{
			Name: "empty channel",
			Args: args{
				ctx: context.Background(),
				c:   ToChan[int](),
			},
			Want: []int{},
		},
		{
			Name: "nil channel",
			Args: args{
				ctx: context.Background(),
				c:   nil,
			},
			Want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			if got := ToSlice(tt.Args.ctx, tt.Args.c); !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("ToSlice() = %v, want %v", got, tt.Want)
			}
		})
	}
}
