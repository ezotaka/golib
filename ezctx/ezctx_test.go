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

package ezctx

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
