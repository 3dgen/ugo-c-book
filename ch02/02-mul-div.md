# 乘除法表达式

本节我们在加减表达式基础之上增加乘除和小括弧的支持。加减表达式只有一种优先级，可以从左到右依次在当前结果的基础上计算出每一步的结果；而加入乘除法和小括号后，由于存在不同优先级，从左到右累加运算无法得到正确的结果。

为了简化分析，我们先假设输入的表达式已经根据优先级被转化为树形结构。比如 1+2*(3+4) 对应以下树形结构：

```
  +
 / \
1   *
   / \
  2   +
     / \
    3   4
```

在该树形结构中：叶子节点为数字；非叶子节点为运算符，其左右子树分别为该运算的两个参数；节点的深度隐式体现了它的运算优先级——越深的节点运算优先级越高。

> **tips** 经验丰富的读者一定已经看出来了，这个结构是表达式语法树的雏形。

我们可以用以下结构存储上述语法树：

```go
type ExprNode struct {
	Value string // +, -, *, /, 123
	Left  *ExprNode
	Right *ExprNode
}

var expr = &ExprNode{
	Value: "+",
	Left: &ExprNode{
		Value: "1",
	},
	Right: &ExprNode{
		Value: "*",
		Left: &ExprNode{
			Value: "2",
		},
		Right: &ExprNode{
			Value: "+",
			Left: &ExprNode{
				Value: "3",
			},
			Right: &ExprNode{
				Value: "4",
			},
		},
	},
}
```

现在我们来构造用于编译`ExprNode`的`Compiler`对象：

```go
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
```

`GenC`方法将`node`编译为C代码，它通过调用`p.genValue(&buf, node)`完成这一操作的主要部分。`genValue`是一个递归方法，它将输入的节点展开为C代码，并返回该节点表达式结果对应的局部变量的名称，实现如下：

```go
func (p *Compiler) genValue(w io.Writer, node *ExprNode) (id string) {
	if node == nil {
		return ""
	}
	id = p.genId()
	switch node.Value {
	case "+":
		fmt.Fprintf(w, "\tint %s = %s + %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	case "-":
		fmt.Fprintf(w, "\tint %s = %s - %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	case "*":
		fmt.Fprintf(w, "\tint %s = %s * %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	case "/":
		fmt.Fprintf(w, "\tint %s = %s / %s;\n",
			id, p.genValue(w, node.Left), p.genValue(w, node.Right),
		)
	default:
		fmt.Fprintf(w, "\tint %s = %s;\n",
			id, node.Value,
		)
	}
	return
}
```

如果输入的节点是加减乘除运算符，则递归处理左右子树并根据运算符转为对应的C代码，并将算式结果保存在新的局部变量中返回；否则输入节点为数字，则直接将数字赋值给新的局部变量返回。其中调用了`p.genId()`方法，它用于创建不重复的局部变量名称：

```go
func (p *Compiler) genId() string {
	id := fmt.Sprintf("t%d", p.nextId)
	p.nextId++
	return id
}
```

在`main`函数中调用`Compiler.Genc`处理`expr`：

```go
func main() {
    ...
	result := run(expr)
	fmt.Println(result)
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
```

我们将在a.out.c中得到以下C代码：

```c
int main() {
	int t1 = 1;
	int t3 = 2;
	int t5 = 3;
	int t6 = 4;
	int t4 = t5 + t6;
	int t2 = t3 * t4;
	int t0 = t1 + t2;
	return t0;
}
```

由此我们就完成了算术表达式语法树到C代码的翻译。