# 加减法表达式

在第一章中，我们通过最小编译器将一个整数编译为返回值等于该整数的程序，本节尝试将加减法表达式编译为同样的程序。

对于特定的加减法表达式，比如`1+3-2`，它对应的C程序如下：

```c
int maint() {
    //1 + 3 - 2
    int t0 = 0 + 1;
    int t1 = t0 + 3;
    int t2 = t1 - 2;
    return t2;
}
```

> **info** 我们没有直接使用`return 1 + 3 -2;`这样的代码，是为了与源项目以LLVM IR输出时保持一致，以此为基础逐步揭开创建AST等功能的面纱。

如果将输入的`1+3-2`转化为`[]string{"1", "+", "3", "-", "2"} `形式，我们则可以通过以下代码输出对应的C程序：

```go
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
```

而如何将输入的字符串拆分为词法单元（既编译器领域常说的Token）数组本质上属于词法分析的问题。我们先以最简单的方式实现：

```go
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
```

基本思路是调用`strings.IndexAny(code, "+-")`函数，根据`+-`字符拆分，返回拆分后的词法单元列表。

对上个版本的编译函数`compile`稍加改造以支持加法和减法的算术表达式：

```go
func compile(code string) {
	tokens := parse_tokens(code)
	output := gen_c(tokens)

	os.WriteFile("a.out.c", []byte(output), 0666)
	exec.Command("gcc", "a.out.c").Run()
}
```

为便于测试，我们再包装一个`run`函数：

```go
func run(code string) int {
	compile(code)
	if err := exec.Command("./a.out").Run(); err != nil {
		return err.(*exec.ExitError).ExitCode()
	}
	return 0
}
```

`run`函数将输入的表达式字符串编译并运行、返回状态码。最后我们更新`main`函数如下：

```go
func main() {
	code, _ := io.ReadAll(os.Stdin)
	fmt.Println(run(string(code)))
}
```

与第一章略为不同，在这里我们使用`io.ReadAll`从命令行读入待编译的字符串，使程序可以不经重新编译即可分析执行不同的输入字符串。在Linux命令行下，通过以下命令执行：

```
$ echo "1+2+3" | go run main.go 
6
```

当然，我们可以基于`run`函数构造单元测试代码如下：

```go
func TestRun(t *testing.T) {
	for i, tt := range tests {
		if got := run(tt.code); got != tt.value {
			t.Fatalf("%d: expect = %v, got = %v", i, tt.value, got)
		}
	}
}

var tests = []struct {
	code  string
	value int
}{
	{code: "1", value: 1},
	{code: "1+1", value: 2},
	{code: "1 + 3 - 2  ", value: 2},
	{code: "1+2+3+4", value: 10},
}
```
