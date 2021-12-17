# 解析表达式，生成语法树

2.2节中我们定义了算术运算的语法树并对其进行了翻译处理，本节将解决如何从表达式中构建语法树的问题，为此，需要开发一个简单的词法解析器和与之配套的语法解析器。

## 2.3.1 词法解析

2.1节解析加减法表达式时，已经实现了一个词法解析器，故以它为基础添加对`*/()`的支持。传统上词法解析器一般被称为lexer，因此我们将词法解析函数更名为`Lex`：

```go
func Lex(code string) (tokens []string) {
	for code != "" {
		if idx := strings.IndexAny(code, "+-*/()"); idx >= 0 {
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

其中`strings.IndexAny`增加了乘除法和小括弧的支持。目前我们暂时忽略错误的输入。开发调试的同时添加测试代码如下：

```go
func TestLex(t *testing.T) {
	var tests = []struct {
		input  string
		tokens []string
	}{
		{"1", []string{"1"}},
		{"1+22*333", []string{"1", "+", "22", "*", "333"}},
		{"1+2*(3+4)", []string{"1", "+", "2", "*", "(", "3", "+", "4", ")"}},
	}
	for i, tt := range tests {
		if got := Lex(tt.input); !reflect.DeepEqual(got, tt.tokens) {
			t.Fatalf("%d: expect = %v, got = %v", i, tt.tokens, got)
		}
	}
}
```

目前的词法解析器虽然简陋，但有了单元测试后，即可以放心重构和优化。词法解析可以参考Rob Pike的报告：https://talks.golang.org/2011/lex.slide

## 2.3.2 语法定义

词法解析输入字符串，输出词法单元（token)序列；语法解析紧随其后，输入词法单元序列，输出结构化的语法树。

在本节中，我们将沿用2.2节的语法树结构：

```go
type ExprNode struct {
	Value string    // +, -, *, /, 123
	Left  *ExprNode `json:",omitempty"`
	Right *ExprNode `json:",omitempty"`
}

func NewExprNode(value string, left, right *ExprNode) *ExprNode {
	return &ExprNode{
		Value: value,
		Left:  left,
		Right: right,
	}
}
```

> **tips** 为便于以JSON格式打印后检查语法树，忽略了空指针，同时增加`NewExprNode`构造函数。

在进行解析语法之前需要明确定义语法规则，下面是四则运算的EBNF规则：

```
expr    = mul ("+" mul | "-" mul)*
mul     = primary ("*" primary | "/" primary)*
primary = num | "(" expr ")"
```

EBNF可以视为正则表达式的增强版本，其中`|`表示或、`()`表示组合、`*`表示0或多个。比如上述规则中：
- `primary`表示数字或由小括号括起来的表达式；
- `mul`表示`primary`间的乘除法。由于`primary`定义，这一规则隐含了“括号的优先级高于乘除法”；
- `expr`表示`mul`间的加减法。由于`mul`的定义，这一规则隐含了“乘除法的优先级高于加减法”。

## 2.3.3 递归下降解析

定义好EBNF规则后，我们就可以它为参考手写一个递归下降的解析程序。首先定义一个`parser`对象，其中包含了词法单元序列和当前处理到的词法单元位置`pos`：

```go
type parser struct {
	tokens []string
	pos    int
}

func (p *parser) peekToken() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *parser) nextToken() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}
```

同时我们定义了2个辅助方法：`peekToken`用于预取下个元素；`nextToken`用于移动到下个元素。

接下来按照2.3.2中的3条EBNF规则，定义3个递归方法，每个方法返回的就是该规则对应的语法树节点：

```go
func (p *parser) build_expr() *ExprNode {
	node := p.build_mul()
	for {
		switch p.peekToken() {
		case "+":
			p.nextToken()
			node = NewExprNode("+", node, p.build_mul())
		case "-":
			p.nextToken()
			node = NewExprNode("-", node, p.build_mul())
		default:
			return node
		}
	}
}
func (p *parser) build_mul() *ExprNode {
	node := p.build_primary()
	for {
		switch p.peekToken() {
		case "*":
			p.nextToken()
			node = NewExprNode("*", node, p.build_primary())
		case "/":
			p.nextToken()
			node = NewExprNode("/", node, p.build_primary())
		default:
			return node
		}
	}
}
func (p *parser) build_primary() *ExprNode {
	if tok := p.peekToken(); tok == "(" {
		p.nextToken()
		node := p.build_expr()
		p.nextToken() // skip ')'
		return node
	} else {
		p.nextToken()
		return NewExprNode(tok, nil, nil)
	}
}
```

再封装一个公开的ParseExpr方法，输入词法单元序列，输出其语法树：

```go
func ParseExpr(tokens []string) *ExprNode {
	p := &parser{tokens: tokens}
	return p.build_expr()
}
```

至此，我们可以将词法解析和语法解析连接起来，并使用2.2节的方法编译执行语法树：

```go
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
```

输出如下：

```
{
    "Value": "+",
    "Left": {
        "Value": "1"
    },
    "Right": {
        "Value": "*",
        "Left": {
            "Value": "2"
        },
        "Right": {
            "Value": "+",
            "Left": {
                "Value": "3"
            },
            "Right": {
                "Value": "4"
            }
        }
    }
}
15
```

可见语法树与我们设想的一致。

## 2.3.4 goyacc等工具
Go再1.5版之前都是基于goyacc工具来产生编译器。但是对于新手来说，并不推荐goyacc和AntLR等自动生成解析器代码的工具，因此删除了这部分内容。首先是手写解析器对于Go这种语法比较规则的语言并不困难，手写代码不仅仅可以熟悉解析器的工作模式，也可以为错误处理带来更大的灵活性。正如Rob Pike所言，我们也不建议通过goyacc自动生成代码的迂回战术、而是要手写解析器的方式迎头而上解决问题。