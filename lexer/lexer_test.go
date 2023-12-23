package lexer

import (
	"fmt"
	"strings"
	"testing"
)

type testLex struct {
	input string
	want  []Item
}

func fmtItems(items []Item) string {
	var (
		builder strings.Builder
		item    Item
	)

	for _, item = range items {
		builder.WriteString(fmt.Sprintln(item))
	}

	return builder.String()
}

func getItems(input string) []Item {
	var (
		item  Item
		items []Item
	)

	for item = range Lex(input) {
		items = append(items, item)
	}

	return items
}

func cmp(got, want []Item) bool {
	var i int

	if len(got) != len(want) {
		return false
	}

	for i = range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func TestLex(t *testing.T) {
	var (
		tests []testLex
		test  testLex
		got   []Item
	)

	tests = []testLex{
		{"",
			[]Item{
				{FilePos{1, 1}, ItemEOF, ""},
			},
		},

		{";;;",
			[]Item{
				{FilePos{1, 1}, ItemEOL, ";"},
				{FilePos{2, 1}, ItemEOL, ";"},
				{FilePos{3, 1}, ItemEOL, ";"},
				{FilePos{4, 1}, ItemEOF, ""},
			},
		},

		{"\n;;\n;",
			[]Item{
				{FilePos{1, 1}, ItemEOL, "\n"},
				{FilePos{1, 2}, ItemEOL, ";"},
				{FilePos{2, 2}, ItemEOL, ";"},
				{FilePos{3, 2}, ItemEOL, "\n"},
				{FilePos{1, 3}, ItemEOL, ";"},
				{FilePos{2, 3}, ItemEOF, ""},
			},
		},

		{"\t \t \t",
			[]Item{
				{FilePos{6, 1}, ItemEOF, ""},
			},
		},
	}

	for _, test = range tests {
		got = getItems(test.input)

		if !cmp(got, test.want) {
			t.Errorf("%q: \ngot: \n%s \nwant: \n%s", test.input, fmtItems(got), fmtItems(test.want))
		}
	}
}
