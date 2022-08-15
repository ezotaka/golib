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
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
)

// ast.File with token.FileSet
type AstFile struct {
	*ast.File
	fset *token.FileSet
}

// Convert to minimized source code text.
func (a AstFile) String() string {
	pp := &printer.Config{}
	buf := bytes.NewBufferString("")

	pp.Fprint(buf, a.fset, a.File)
	return buf.String()
}

// NewAstFile parses the source code of a single Go source file and returns
// the corresponding ast.File node
func NewAstFile(file string) (*AstFile, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.Mode(0))

	if err != nil {
		return nil, err
	}

	return &AstFile{
		fset: fset,
		File: f,
	}, nil
}

// Save AST as source code to temp file.
//
// SaveTemp returns filePath, temp file cleaner func , and error.
// Cleaner func is not always nil, so you may call it any time.
func (a AstFile) SaveTemp(file string) (outFile string, cleaner func(), err error) {
	cleaner = func() {}

	if file == "" {
		err = errors.New("file must not be empty")
		return
	}

	dir, err := os.MkdirTemp("", "ezast")
	if err != nil {
		return
	}
	cleaner = func() { os.RemoveAll(dir) }

	outFile = filepath.Join(dir, file)
	err = os.WriteFile(outFile, []byte(a.String()), 0666)
	return
}
