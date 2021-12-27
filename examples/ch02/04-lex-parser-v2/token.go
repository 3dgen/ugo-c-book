package main

import "fmt"

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

func (op TokenType) Precedence() int {
	switch op {
	case ADD, SUB:
		return 1
	case MUL, DIV:
		return 2
	}
	return 0
}