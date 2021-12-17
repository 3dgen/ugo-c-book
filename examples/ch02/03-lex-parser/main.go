package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

var code = "1+2*(3+4)"

func main() {
	expr_tokens := Lex(code)
	ast := ParseExpr(expr_tokens)
	fmt.Println(JSONString(ast))

	fmt.Println(run(ast))
}

func JSONString(x interface{}) string {
	d, _ := json.MarshalIndent(x, "", "    ")
	return string(d)
}

func run(node *ExprNode) int {
	compile(node)
	if err := exec.Command("./a.out").Run(); err != nil {
		return err.(*exec.ExitError).ExitCode()
	}
	return 0
}

func compile(node *ExprNode) {
	output := new(Compiler).GenC(node)

	os.WriteFile("a.out.c", []byte(output), 0666)
	exec.Command("gcc", "a.out.c").Run()
}