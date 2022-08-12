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

package conv

import (
	"reflect"
	"testing"
)

func TestChan(t *testing.T) {
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
			if got := Slice(Chan(tt.args.values...)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlice(t *testing.T) {
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
				c: Chan(1, 2),
			},
			want: []int{1, 2},
		},
		{
			name: "empty channel",
			args: args{
				c: Chan[int](),
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
			if got := Slice(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}
