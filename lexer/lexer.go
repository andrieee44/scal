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

	ErrorNumber      string = "expected a number"
	ErrorHexNoDigits string = "hexadecmial has no numbers"
	ErrorUnexpected  string = "unexpected character"

	eof       rune   = 0
	digits    string = "0123456789"
	signs     string = "+-"
	operators string = "+-*/"
	hexPrefix string = "Xx"
	hex       string = digits + "ABCDEFabcdef"
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

func consume(l *lexer, match string) bool {
	if strings.IndexRune(match, next(l)) != -1 {
		return true
	}

	prev(l)
	return false
}

func consumeAll(l *lexer, match string) int {
	var n int

	for n = 0; consume(l, match); n++ {
	}

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

func space(l *lexer) {
	var r rune

	for r = next(l); r != '\n' && unicode.IsSpace(r); r = next(l) {
		skip(l)
	}

	prev(l)
}

func operator(l *lexer) stateFn {
	emit(l, ItemOperator)
	space(l)
	return number
}

func number(l *lexer) stateFn {
	if consumeAll(l, signs) > 1 {
		return die(l, ErrorNumber)
	}

	switch {
	case consume(l, "0") && consume(l, hexPrefix):
		if consumeAll(l, hex) == 0 {
			return die(l, ErrorHexNoDigits)
		}
	case consumeAll(l, digits) == 0:
		return die(l, ErrorNumber)
	}

	emit(l, ItemNumber)

	space(l)
	if consume(l, operators) {
		return operator
	}

	return start
}

func start(l *lexer) stateFn {
	space(l)

	switch {
	case consume(l, signs), consume(l, digits):
		prev(l)
		return number(l)
	case consume(l, "\n"):
		l.indexPos.X = 1
		l.indexPos.Y++
		emit(l, ItemEOL)
	case consume(l, ";"):
		emit(l, ItemEOL)
	case consume(l, string(eof)):
		emit(l, ItemEOF)
		return nil
	default:
		return die(l, ErrorUnexpected)
	}

	return start
}

func run(l *lexer) {
	var state stateFn

	for state = start; state != nil; state = state(l) {
	}

	close(l.items)
}
