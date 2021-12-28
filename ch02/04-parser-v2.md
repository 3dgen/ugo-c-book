# 2.4 重构词法解析器

在之前的章节中，我们使用字符串来记录词法单元（token），这一方式虽然简单，但有诸多缺陷：比如当某个词法单元是标识符时，难以区分它究竟是关键字，还是变量名或函数名等自定义标识符。本节将对次进行改进，使其更接近常用的词法解析器。

## 2.4.1 更新词法单元定义

新的词法单元结构如下：

```go
// 词法单元类型
type TokenType int

const (
	EOF TokenType = iota
	ILLEGAL

	NUMBER

	ADD // +
	SUB // -
	MUL // *
	DIV // /

	LPAREN // (
	RPAREN // )
)

// 词法单元
type Token struct {
	Type TokenType // 词法单元类型
	Val  string    // 词法单元原始字面值
	Pos  int       // 开始位置
}
```

在该结构中，`Type`成员用于标识该词法单元的类型。除表示文件结尾的`EOF`和表示无效值的`ILLEGAL`外，其余的词法单元类型在之前的小节中均已接触过，它们分别是表示数字的`NUMBER`、表示加减乘除运算的`ADD`/`SUB`/`MUL`/`DIV`以及左右小括号`(`/`)`。除此之外，`Token.Val`为词法单元的原始字面值，`Token.Pos`为该词法单元位于源代码字符串中的开始位置。

为便于对二元表达式进行优先级处理，增加`Precedence`方法用于获取每类词法单元的优先级，如果该类词法单元不是二元运算符则返回0：

```go
func (op TokenType) Precedence() int {
	switch op {
	case ADD, SUB:
		return 1
	case MUL, DIV:
		return 2
	}
	return 0
}
```

为便于词法单元打印输出，增加字符串化辅助方法：

```go
func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case ILLEGAL:
		return "ILLEGAL"
    //...
	default:
		return "UNKNWON"
	}
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%v:%v)", t.Type, t.Val)
}
```

## 2.4.2 调整语法树

由于词法单元由`string`变为`Token`，语法树节点结构需要进行相应调整如下：

```go
type ExprNode struct {
	Token           // +, -, *, /, 123
	Left  *ExprNode `json:",omitempty"`
	Right *ExprNode `json:",omitempty"`
}

func NewExprNode(token Token, left, right *ExprNode) *ExprNode {
	return &ExprNode{
		Token: token,
		Left:  left,
		Right: right,
	}
}
```

## 2.4.3 更新词法解析函数

词法解析的主要作用是将源代码字符串拆分为词法单元，在之前的章节中我们使用`string.IndexAny`来执行分词操作，随着词法单元定义的更新，现在我们改用`text/scanner`包来辅助解析：

```go
func Lex(code string) (tokens []Token) {
	var s scanner.Scanner
	s.Init(strings.NewReader(code))
	for x := s.Scan(); x != scanner.EOF; x = s.Scan() {
		var tok = Token{
			Val: s.TokenText(),
			Pos: s.Pos().Offset,
		}
		switch x {
		case scanner.Int:
			tok.Type = NUMBER
		default:
			switch s.TokenText() {
			case "+":
				tok.Type = ADD
			case "-":
				tok.Type = SUB
			case "*":
				tok.Type = MUL
			case "/":
				tok.Type = DIV
			case "(":
				tok.Type = LPAREN
			case ")":
				tok.Type = RPAREN
			default:
				tok.Type = ILLEGAL
				tokens = append(tokens, tok)
				return
			}
		}

		tokens = append(tokens, tok)
	}

	tokens = append(tokens, Token{Type: EOF})
	return
}
```

`scanner.Scanner`可以提取我们需要的整数、四则运算符和小括号，遇到这些词法单元时我们保存了它们的原始字面值和在源代码中出现的位置，并赋予了对应的词法单元类型值。其他类型的词法单元我们暂时不支持，以`ILLEGAL`类型标记。

> **tips:** 目前我们的语法结构很简单，`scanner.Scanner`即可满足要求。更复杂的词法解析功能的实现，大家可参考`go/token`包。

## 2.4.4 增加词法单元读取器

`Lex`函数将源代码转换为词法单元序列，为方便后续语法解析器构建语法树，我们将词法单元读取、预读取、回退等功能拆分出来，单独封装为词法单元读取器：

```go
type TokenReader struct {
	tokens []Token
	pos    int
	width  int
}

func NewTokenReader(input []Token) *TokenReader {
	return &TokenReader{tokens: input}
}

func (p *TokenReader) PeekToken() Token {
	tok := p.ReadToken()
	p.UnreadToken()
	return tok
}

func (p *TokenReader) ReadToken() Token {
	if p.pos >= len(p.tokens) {
		p.width = 0
		return Token{Type: EOF}
	}
	tok := p.tokens[p.pos]
	p.width = 1
	p.pos += p.width
	return tok
}

func (p *TokenReader) UnreadToken() {
	p.pos -= p.width
	return
}
```

除此之外，我们再为词法单元读取器`TokenReader`增加`AcceptToken`、`MustAcceptToken`两个方法：

```go
func (p *TokenReader) AcceptToken(expectTypes ...TokenType) (tok Token, ok bool) {
	tok = p.ReadToken()
	for _, x := range expectTypes {
		if tok.Type == x {
			return tok, true
		}
	}
	p.UnreadToken()
	return tok, false
}

func (p *TokenReader) MustAcceptToken(expectTypes ...TokenType) (tok Token) {
	tok, ok := p.AcceptToken(expectTypes...)
	if !ok {
		panic(fmt.Errorf("token.Reader.MustAcceptToken(%v) failed", expectTypes))
	}
	return tok
}
```

`AcceptToken`用于判断接下来的词法单元的类型是否为`expectTypes`之一;`MustAcceptToken`则强制断言接下来的词法单元必须为`expectTypes`之一，该方法用于处理左右小括号必须成对出现等语法规则。

## 2.4.5 二元表达式解析简化

BNF语法可以实现表达式的多优先级支持，比如2.3节中我们使用过加减乘除四则运算的EBNF规则：

```
expr    = mul ("+" mul | "-" mul)*
mul     = primary ("*" primary | "/" primary)*
primary = num | "(" expr ")"
```

其中通过多级规则定义的方法实现了**乘除法优先级高于加减法**、**小括号优先级高于乘除法**。而Go语言的二元表达式有`||`、`&&`、`==`、`+`和`*`等5钟不同的优先级，如果完全通过EBNF来表示优先级则需要构造更为复杂的规则：

```
expr       = logic_or
logic_or   = logic_and ("||" logic_and)*
logic_and  = equality ("&&" relational)*
equality   = relational ("==" relational | "!=" relational)*
add        = mul ("+" mul | "-" mul)*
mul        = unary ("*" unary | "/" unary)*
unary      = ("+" | "-")? primary
primary    = num | "(" expr ")"
```

这种复杂性有悖于Go语言推崇的**少即是多**原则。Go语言在设计表达式时有意无意地忽略了对右结合二元表达式的支持而只存在左结合表达式，如果将运算符优先级的判定后置，则ENBF规则可以定义得很简单，比如四则运算的简化ENBF规则如下：

```
expr  = unary ("+" | "-" | "*" | "/") unary)*
unary = ("+" | "-")? primary
primary    = num | "(" expr ")"
```

既只有二元和一元表达式之分，而不再区分二元表达式的优先级。为正确控制优先级（既判断表达式左结合时机），并配合新的词法单元结构，语法解析部分需作相应调整。

`ParseExpr`函数调整如下：

```go
func ParseExpr(input []Token) *ExprNode {
	r := NewTokenReader(input)
	return parseExpr(r)
}

func parseExpr(r *TokenReader) *ExprNode {
	return parseExpr_binary(r, 1)
}
```

`ParseExpr`函数为传入的词法单元序列创建`TokenReader`，然后以优先级1为参数调用`parseExpr_binary`函数解析二元表达式。`parseExpr_binary`实现如下：

```go
func parseExpr_binary(r *TokenReader, prec int) *ExprNode {
	x := parseExpr_unary(r)
	for {
		op := r.PeekToken()
		if op.Type.Precedence() < prec {
			return x
		}

		r.MustAcceptToken(op.Type)
		y := parseExpr_binary(r, op.Type.Precedence()+1)
		x = &ExprNode{Token: op, Left: x, Right: y}
	}
	return nil
}
```

首先调用`parseExpr_unary`函数获取一个一元表达式`x`，然后预取下一个运算符`op`，若`op`的优先级低于当前所处表达式的优先级则结束左结合，否则递归调用`parseExpr_binary`解析表达式获得`op`的右部`y`，与`op`左部`x`结合为新子树后赋给`x`。

一元表达式解析函数`parseExpr_unary`实现如下：

```go
func parseExpr_unary(r *TokenReader) *ExprNode {
	if _, ok := r.AcceptToken(ADD); ok {
		return parseExpr_primary(r)
	}
	if _, ok := r.AcceptToken(SUB); ok {
		return &ExprNode{
			Token: Token{Type: SUB},
			Left:  &ExprNode{Token: Token{Type: NUMBER, Val: "0"}},
			Right: parseExpr_primary(r),
		}
	}
	return parseExpr_primary(r)
}
```

对形如`+x`的表达式，则返回`x`节点；对形如`-x`的表达式，则返回`0-x`节点。`parseExpr_primary`函数解析数值或以小括弧括起来的表达式，如下：

```go
func parseExpr_primary(r *TokenReader) *ExprNode {
	if _, ok := r.AcceptToken(LPAREN); ok {
		expr := parseExpr(r)
		r.MustAcceptToken(RPAREN)
		return expr
	}
	return &ExprNode{
		Token: r.MustAcceptToken(NUMBER),
	}
}
```

至此我们得到了一个更简洁的、支持多优先级、只有左结合二元表达式的解析器。以后如果引入了更多不同优先级的运算符，只需要更新词法单元类型`TokenType`和它的优先级方法`Precedence`即可。