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

func TestAstFile_ChangePackage(t *testing.T) {
	type args struct {
		newPkg string
	}
	tests := []struct {
		name    string
		file    string
		args    args
		want    *AstFile
		wantErr bool
	}{
		{
			name: "package name is changed normally",
			file: "testdata/src1.go",
			args: args{
				newPkg: "pkg",
			},
			want: toAst("testdata/src1pkg.go"),
		},
		{
			name: "empty package name",
			file: "testdata/src1.go",
			args: args{
				newPkg: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAstFile(tt.file)
			err := a.ChangePackage(tt.args.newPkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("AstFile.ChangePackage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !AssertEqualAstFile(tt.want.File, a.File) {
				t.Errorf("AstFile.ChangePackage() changed = %v, want %v", a.String(), tt.want.String())
			}
		})
	}
}
