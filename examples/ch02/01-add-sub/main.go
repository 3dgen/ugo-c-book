package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	code, _ := io.ReadAll(os.Stdin)
	fmt.Println(run(string(code)))
}

func run(code string) int {
	compile(code)
	if err := exec.Command("./a.out").Run(); err != nil {
		return err.(*exec.ExitError).ExitCode()
	}
	return 0
}

func compile(code string) {
	tokens := parse_tokens(code)
	output := gen_c(tokens)

	os.WriteFile("a.out.c", []byte(output), 0666)
	exec.Command("gcc", "a.out.c").Run()
}

func parse_tokens(code string) (tokens []string) {
	for code != "" {
		if idx := strings.IndexAny(code, "+-"); idx >= 0 {
			if idx > 0 {
				tokens = append(tokens, strings.TrimSpace(code[:idx]))
			}
			tokens = append(tokens, code[idx:][:1])
			code = code[idx+1:]
			continue
		}

		tokens = append(tokens, strings.TrimSpace(code))
		return
	}
	return
}

func gen_c(tokens []string) string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, `int main() {`)

	var idx int
	for i, tok := range tokens {
		if i == 0 {
			fmt.Fprintf(&buf, "\tint t%d = 0 + %v;\n",
				idx, tokens[i],
			)
			continue
		}
		switch tok {
		case "+":
			idx++
			fmt.Fprintf(&buf, "\tint t%d = t%d + %v;\n",
				idx, idx-1, tokens[i+1],
			)
		case "-":
			idx++
			fmt.Fprintf(&buf, "\tint t%d = t%d - %v;\n",
				idx, idx-1, tokens[i+1],
			)
		}
	}
	fmt.Fprintf(&buf, "\treturn t%d;\n", idx)
	fmt.Fprintln(&buf, `}`)

	return buf.String()
}
