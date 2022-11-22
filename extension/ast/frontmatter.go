package ast

import (
	"fmt"
	"github.com/yuin/goldmark/ast"
)

type FrontmatterStart struct {
	ast.BaseBlock
	MetaData map[interface{}]interface{}
	Status   string
}

var KindFrontmatterStart = ast.NewNodeKind("FrontmatterStart")

func (f FrontmatterStart) Kind() ast.NodeKind {
	return KindFrontmatterStart
}

func (f *FrontmatterStart) Dump(source []byte, level int) {
	m := map[string]string{}
	m["_status"] = fmt.Sprintf("%v", f.Status)
	for key, value := range f.MetaData {
		m[fmt.Sprintf("%v", key)] = fmt.Sprintf("%v", value)
	}
	ast.DumpHelper(f, source, level, m, nil)
}
