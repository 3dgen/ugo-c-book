package main

import (
	"bytes"
	"fmt"
	"io"
)

type Compiler struct {
	nextId int
}

func (p *Compiler) GenC(node *ExprNode) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "int main() {\n")
	fmt.Fprintf(&buf, "\treturn %s;\n", p.genValue(&buf, node))
	fmt.Fprintf(&buf, "}\n")

	return buf.String()
}

func (p *Compiler) genValue(w io.Writer, node *ExprNode) (id string) {
	if node == nil {
		return ""
	}
	id = p.genId()
	switch node.Token.Type {
	case ADD:
		fmt.Fprintf(w, "\tint %s = %s + %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	case SUB:
		fmt.Fprintf(w, "\tint %s = %s - %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	case MUL:
		fmt.Fprintf(w, "\tint %s = %s * %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	case DIV:
		fmt.Fprintf(w, "\tint %s = %s / %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	default:
		fmt.Fprintf(w, "\tint %s = %s;\n",
			id, node.Token.Val,
		)
	}
	return
}

func (p *Compiler) genId() string {
	id := fmt.Sprintf("t%d", p.nextId)
	p.nextId++
	return id
}
