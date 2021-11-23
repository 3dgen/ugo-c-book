//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

var code = "42"

func main() {
	compile(string(code))
}

func compile(code string) {
	output := fmt.Sprintf(tmpl, code)
	os.WriteFile("a.out.c", []byte(output), 0666)
	exec.Command("gcc", "a.out.c").Run()
}

const tmpl = `
int main(){
	return %v;
}
`