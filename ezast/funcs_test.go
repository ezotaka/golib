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

import "testing"

func TestAstFile_GetFuncDecl(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name string
		ast  *AstFile
		args args
		want bool
	}{
		{
			name: "contains",
			ast:  toAst("testdata/src1.go"),
			args: args{
				fn: "test",
			},
			want: true,
		},
		{
			name: "not contains",
			ast:  toAst("testdata/src1.go"),
			args: args{
				fn: "notContains",
			},
			want: false,
		},
		{
			name: "empty func name",
			ast:  toAst("testdata/src1.go"),
			args: args{
				fn: "",
			},
			want: false,
		},
		{
			name: "source has no func",
			ast:  toAst("testdata/empty.go"),
			args: args{
				fn: "test",
			},
			want: false,
		},
		{
			name: "inner func is not target of ContainsFunc",
			ast:  toAst("testdata/inner.go"),
			args: args{
				fn: "inner",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.ast.GetFuncDecl(tt.args.fn)
			if tt.want {
				if got == nil || got.Name.Name != tt.args.fn {
					t.Errorf("AstFile.GetFuncDecl() = %v, want %v", got, nil)
				}
			} else {
				if got != nil {
					t.Errorf("AstFile.GetFuncDecl() = %v, want %v", got, nil)
				}
			}
		})
	}
}

func TestAstFile_ContainsFunc(t *testing.T) {
	type args struct {
		fn string
	}
	tests := []struct {
		name string
		ast  *AstFile
		args args
		want bool
	}{
		{
			name: "contains",
			ast:  toAst("testdata/src1.go"),
			args: args{
				fn: "test",
			},
			want: true,
		},
		{
			name: "not contains",
			ast:  toAst("testdata/src1.go"),
			args: args{
				fn: "notContains",
			},
			want: false,
		},
		{
			name: "empty func name",
			ast:  toAst("testdata/src1.go"),
			args: args{
				fn: "",
			},
			want: false,
		},
		{
			name: "source has no func",
			ast:  toAst("testdata/empty.go"),
			args: args{
				fn: "test",
			},
			want: false,
		},
		{
			name: "inner func is not target of ContainsFunc",
			ast:  toAst("testdata/inner.go"),
			args: args{
				fn: "inner",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.ast.ContainsFunc(tt.args.fn); got != tt.want {
				t.Errorf("astFile.ContainsFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAstFile_SwapFunc_isStrict(t *testing.T) {
	toAst := func(file string) *AstFile {
		f, _ := NewAstFile(file)
		return f
	}

	type args struct {
		fn1 string
		fn2 string
	}
	tests := []struct {
		name    string
		ast     *AstFile
		args    args
		want    *AstFile
		wantErr bool
	}{
		{
			name: "swap two different funcs",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "test",
				fn2: "more",
			},
			want: toAst("testdata/src3swap.go"),
		},
		{
			name: "if try to swap same funcs, do nothing",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "test",
				fn2: "test",
			},
			want: toAst("testdata/src3.go"),
		},
		{
			name: "first func is not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "notFound",
				fn2: "test",
			},
			wantErr: true,
		},
		{
			name: "second func is not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "test",
				fn2: "notFound",
			},
			wantErr: true,
		},
		{
			name: "both different funcs are not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "notFound1",
				fn2: "notFound2",
			},
			wantErr: true,
		},
		{
			name: "both same funcs are not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "notFound",
				fn2: "notFound",
			},
			wantErr: true,
		},
		{
			name: "first func has any args",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "hasArgs",
				fn2: "noArgsReturns",
			},
			wantErr: true,
		},
		{
			name: "second func has any args",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "noArgsReturns",
				fn2: "hasArgs",
			},
			wantErr: true,
		},
		{
			name: "first func has any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "hasReturns",
				fn2: "noArgsReturns",
			},
			wantErr: true,
		},
		{
			name: "second func has any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "noArgsReturns",
				fn2: "hasReturns",
			},
			wantErr: true,
		},
		{
			name: "first func has any args and any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "hasArgsAndReturns",
				fn2: "noArgsReturns",
			},
			wantErr: true,
		},
		{
			name: "second func has any args and any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "noArgsReturns",
				fn2: "hasArgsAndReturns",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.ast.SwapSimpleFuncs(tt.args.fn1, tt.args.fn2, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("AstFile.SwapFunc() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !AssertEqualAstFile(tt.want.File, tt.ast.File) {
				t.Errorf("AstFile.SwapFunc() changed = %v, want %v", tt.ast.String(), tt.want.String())
			}
		})
	}
}

func TestAstFile_SwapFunc_isNotStrict(t *testing.T) {
	type args struct {
		fn1 string
		fn2 string
	}
	tests := []struct {
		name    string
		ast     *AstFile
		args    args
		want    *AstFile
		wantErr bool
	}{
		{
			name: "swap tow different funcs",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "test",
				fn2: "more",
			},
			want: toAst("testdata/src3swap.go"),
		},
		{
			name: "if try to swap same funcs, do nothing",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "test",
				fn2: "test",
			},
			want: toAst("testdata/src3.go"),
		},
		{
			name: "first func is not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "notFound",
				fn2: "test",
			},
			want: toAst("testdata/src3notFound.go"),
		},
		{
			name: "second func is not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "test",
				fn2: "notFound",
			},
			want: toAst("testdata/src3notFound.go"),
		},
		{
			name: "both different funcs are not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "notFound1",
				fn2: "notFound2",
			},
			want: toAst("testdata/src3.go"),
		},
		{
			name: "both same funcs are not found",
			ast:  toAst("testdata/src3.go"),
			args: args{
				fn1: "notFound",
				fn2: "notFound",
			},
			want: toAst("testdata/src3.go"),
		},
		{
			name: "first func has any args",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "hasArgs",
				fn2: "noArgsReturns",
			},
			wantErr: true,
		},
		{
			name: "second func has any args",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "noArgsReturns",
				fn2: "hasArgs",
			},
			wantErr: true,
		},
		{
			name: "first func has any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "hasReturns",
				fn2: "noArgsReturns",
			},
			wantErr: true,
		},
		{
			name: "second func has any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "noArgsReturns",
				fn2: "hasReturns",
			},
			wantErr: true,
		},
		{
			name: "first func has any args and any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "hasArgsAndReturns",
				fn2: "noArgsReturns",
			},
			wantErr: true,
		},
		{
			name: "second func has any args and any return values",
			ast:  toAst("testdata/argsReturns.go"),
			args: args{
				fn1: "noArgsReturns",
				fn2: "hasArgsAndReturns",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.ast.SwapSimpleFuncs(tt.args.fn1, tt.args.fn2, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("AstFile.SwapFunc() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !AssertEqualAstFile(tt.want.File, tt.ast.File) {
				t.Errorf("AstFile.SwapFunc() changed = %v, want %v", tt.ast.String(), tt.want.String())
			}
		})
	}
}
