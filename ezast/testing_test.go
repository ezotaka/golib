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

package ezast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func toAst(file string) *AstFile {
	f, _ := NewAstFile(file)
	return f
}
func TestAssertEqualAstFile(t *testing.T) {
	const (
		src1                 = "testdata/src1.go"
		src1AndIgnorableText = "testdata/src2.go"
		src1AndMoreFunc      = "testdata/src3.go"
	)

	toAst := func(file string) *ast.File {
		f, _ := parser.ParseFile(token.NewFileSet(), file, nil, parser.Mode(0))
		return f
	}

	type args struct {
		want *ast.File
		got  *ast.File
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same completely",
			args: args{
				want: toAst(src1),
				got:  toAst(src1),
			},
			want: true,
		},
		{
			name: "blank and comment are ignored",
			args: args{
				want: toAst(src1),
				got:  toAst(src1AndIgnorableText),
			},
			want: true,
		},
		{
			name: "not same",
			args: args{
				want: toAst(src1),
				got:  toAst(src1AndMoreFunc),
			},
			want: false,
		},
		{
			name: "nil and nil",
			args: args{
				want: nil,
				got:  nil,
			},
			want: true,
		},
		{
			name: "not nil and nil",
			args: args{
				want: toAst(src1),
				got:  nil,
			},
			want: false,
		},
		{
			name: "nil and not nil",
			args: args{
				want: nil,
				got:  toAst(src1),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := AssertEqualAstFile(tt.args.want, tt.args.got); got != tt.want {
				t.Errorf("AssertAstFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
