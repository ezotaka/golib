// Copyright (c) 2022 Takatomo Ezo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
