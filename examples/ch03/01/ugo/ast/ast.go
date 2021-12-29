package ast

import (
	"github.com/chai2010/ugo/token"
)

// File 表示 µGo 源文件对应的语法树.
type File struct {
	Pkg   *Package // 包信息
	Funcs []*Func  // 函数列表
}

// 包信息
type Package struct {
	PkgPos  int    // package 关键字位置
	NamePos int    // 包名位置
	Name    string // 包名
}

// 函数信息
type Func struct {
	FuncPos int
	NamePos int
	Name    string
	Body    *BlockStmt
}

// 语句块
type BlockStmt struct {
	Lbrace int // '{'
	List   []Stmt
	Rbrace int // '}'
}

// 语句
type Stmt interface {
	Pos() int
	End() int
	stmt_type()
}

// 表达式语句
type ExprStmt struct {
	X Expr
}

// 表达式
type Expr interface {
	Pos() int
	End() int
	expr_type()
}

// 基本整数
type Number struct {
	ValuePos int
	ValueEnd int
	Value    int
}

// 二元表达式
type BinaryExpr struct {
	Op token.Token // 运算符
	X  Expr        // 左部
	Y  Expr        // 右部
}

// 一元表达式
type UnaryExpr struct {
	Op token.Token // 运算符
	X  Expr        // 运算对象
}

// 小括号表达式
type ParenExpr struct {
	X Expr // 小括号内的表达式对象
}

// 函数调用表达式
type CallExpr struct {
	FuncPos  int    // func 关键字位置
	FuncName string // 函数名字
	Lparen   int    // '(' 位置
	Args     []Expr // 调用参数列表
	Rparen   int    // ')' 位置
}
