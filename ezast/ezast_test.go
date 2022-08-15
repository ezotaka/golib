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
	"os"
	"path/filepath"
	"testing"
)

func TestAstFile_SaveTemp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		t.Parallel()

		const file = "main.go"

		ast, err := NewAstFile("testdata/src1.go")
		if err != nil {
			t.Fatalf("NewAstFile() returns err %v", err)
		}

		// execute method to be tested
		gotOutFile, gotCleaner, err := ast.SaveTemp(file)

		if gotCleaner == nil {
			t.Error("AstFile.SaveTemp() gotCleaner = nil, want not nil")
		}

		if err != nil {
			t.Errorf("AstFile.SaveTemp() error = %v, want no errors", err)
			return
		}

		if _, err := os.Stat(gotOutFile); err != nil {
			t.Errorf("AstFile.SaveTemp() file %v is not found", gotOutFile)
			return
		}

		defer func() {
			gotCleaner()

			if _, err := os.Stat(gotOutFile); err == nil {
				t.Errorf("AstFile.SaveTemp() gotCleaner does not clean file %v", gotOutFile)
				return
			}
		}()

		if filepath.Base(gotOutFile) != file {
			t.Errorf("AstFile.SaveTemp() gotOutFile = %v, want file name %v", gotOutFile, file)
		}

		gotAst, err := NewAstFile(gotOutFile)
		if err != nil {
			t.Errorf("AstFile.SaveTemp() error from NewAstFile(gotOutFile) = %v, want no errors", err)
		}

		if !AssertEqualAstFile(ast.File, gotAst.File) {
			t.Errorf("AST of gotOutFile = %v, want %v", gotAst.String(), ast.String())
		}
	})

	t.Run("file is empty", func(t *testing.T) {
		t.Parallel()

		ast, err := NewAstFile("testdata/src1.go")
		if err != nil {
			t.Fatalf("NewAstFile() returns err %v", err)
		}

		// execute method to be tested
		_, gotCleaner, err := ast.SaveTemp("")

		if gotCleaner == nil {
			t.Error("AstFile.SaveTemp() gotCleaner = nil, want not nil")
		}

		if err == nil {
			t.Error("AstFile.SaveTemp() error = nil, want error telling that file not found")
		}
	})
}
