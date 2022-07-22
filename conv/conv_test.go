package conv

import (
	"context"
	"reflect"
	"testing"
)

func TestToChan(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "length=0",
			args: args{
				values: []int{},
			},
			want: []int{},
		},
		{
			name: "length=1",
			args: args{
				values: []int{1},
			},
			want: []int{1},
		},
		{
			name: "length=2",
			args: args{
				values: []int{1, 2},
			},
			want: []int{1, 2},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ToSlice(context.Background(), ToChan(tt.args.values...)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToChan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSlice(t *testing.T) {
	type args struct {
		c <-chan int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "{1, 2}",
			args: args{
				c: ToChan(1, 2),
			},
			want: []int{1, 2},
		},
		{
			name: "empty channel",
			args: args{
				c: ToChan[int](),
			},
			want: []int{},
		},
		{
			name: "nil channel",
			args: args{
				c: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ToSlice(context.Background(), tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToChan() = %v, want %v", got, tt.want)
			}
		})
	}
}
