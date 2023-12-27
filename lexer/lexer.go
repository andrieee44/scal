package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type FilePos struct {
	X, Y int
}

type ItemType int

type Item struct {
	Pos   FilePos
	Type  ItemType
	Value string
}

type lexer struct {
	startPos, indexPos  FilePos
	start, index, width int
	input               string
	items               chan Item
}

type stateFn func(*lexer) stateFn

//go:generate stringer -type=ItemType
const (
	ItemEOF ItemType = iota
	ItemEOL
	ItemError
	ItemNumber
	ItemOperator

	ErrorNumber string = "expected a number"
	ErrorIllegal string = "illegal character"

	eof       rune   = 0
	digits    string = "0123456789"
	signs     string = "+-"
	operators string = "+-*/"
)

func (item Item) String() string {
	if item.Type == ItemEOF {
		item.Value = "EOF"
	}

	return fmt.Sprintf("%d:%d (%s) %q", item.Pos.X, item.Pos.Y, item.Type.String(), item.Value)
}

func Lex(input string) chan Item {
	var l *lexer

	l = &lexer{
		indexPos: FilePos{X: 1, Y: 1},
		input:    input,
		items:    make(chan Item),
	}

	l.startPos = l.indexPos
	go run(l)
	return l.items
}

func next(l *lexer) rune {
	var r rune

	if l.index >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.index:])
	l.index += l.width
	l.indexPos.X += l.width
	return r
}

func prev(l *lexer) {
	l.index -= l.width
	l.indexPos.X -= l.width
}

func peek(l *lexer) rune {
	var r rune

	r = next(l)
	prev(l)
	return r
}

func skip(l *lexer) {
	l.start = l.index
	l.startPos = l.indexPos
}

func consume(l *lexer, match string) int {
	var n int

	for n = 0; strings.IndexRune(match, next(l)) != -1; n++ {
	}

	prev(l)
	return n
}

func emit(l *lexer, typ ItemType) {
	l.items <- Item{
		Pos:   l.startPos,
		Type:  typ,
		Value: l.input[l.start:l.index],
	}

	l.startPos = l.indexPos
	l.start = l.index
}

func die(l *lexer, errMsg string) stateFn {
	l.items <- Item{
		Pos:   l.startPos,
		Type:  ItemError,
		Value: errMsg,
	}

	return nil
}

func operator(l *lexer) stateFn {
	emit(l, ItemOperator)
	return number
}

func number(l *lexer) stateFn {
	if consume(l, signs) > 1 || consume(l, digits) == 0 {
		return die(l, ErrorNumber)
	}

	emit(l, ItemNumber)

	if strings.IndexRune(operators, next(l)) != -1 {
		return operator
	}

	prev(l)
	return start
}

func start(l *lexer) stateFn {
	var r rune

	r = next(l)

	switch {
	case r == eof:
		emit(l, ItemEOF)
		return nil
	case r == '\n':
		l.indexPos.X = 1
		l.indexPos.Y++
		emit(l, ItemEOL)
	case r == ';':
		emit(l, ItemEOL)
	case strings.IndexRune(signs, r) != -1, strings.IndexRune(digits, r) != -1:
		prev(l)
		return number(l)
	case unicode.IsSpace(r):
		skip(l)
	default:
		return die(l, ErrorIllegal)
	}

	return start
}

func run(l *lexer) {
	var state stateFn

	for state = start; state != nil; state = state(l) {
	}

	close(l.items)
}
