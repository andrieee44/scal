package lexer

import (
	"fmt"
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

	eof rune = 0
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

func emit(l *lexer, typ ItemType) {
	l.items <- Item{
		Pos:   l.startPos,
		Type:  typ,
		Value: l.input[l.start:l.index],
	}

	l.startPos = l.indexPos
	l.start = l.index
}

func start(l *lexer) stateFn {
	var r rune

	r = next(l)

	switch r {
	case eof:
		emit(l, ItemEOF)
		return nil
	case '\n':
		l.indexPos.X = 1
		l.indexPos.Y++
		emit(l, ItemEOL)
	case ';':
		emit(l, ItemEOL)
	}

	if unicode.IsSpace(r) {
		skip(l)
	}

	return start
}

func run(l *lexer) {
	var state stateFn

	for state = start; state != nil; state = state(l) {
	}

	close(l.items)
}
