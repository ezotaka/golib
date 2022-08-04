package channel

import (
	"reflect"
	"testing"

	"github.com/ezotaka/golib/conv"
)

func TestEnumerate(t *testing.T) {
	type args struct {
		c <-chan string
	}
	closed := make(chan string)
	close(closed)
	tests := []struct {
		name string
		args args
		want []forEnum[string]
	}{
		{
			name: "normal",
			args: args{
				c: conv.Chan("a", "b"),
			},
			want: []forEnum[string]{
				{
					I: 0,
					V: "a",
				},
				{
					I: 1,
					V: "b",
				},
			},
		},
		{
			name: "closed channel",
			args: args{
				c: closed,
			},
			want: []forEnum[string]{},
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
			if got := conv.Slice(Enumerate(tt.args.c)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enumerate() = %v, want %v", got, tt.want)
			}
		})
	}
}
