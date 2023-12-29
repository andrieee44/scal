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

		{"\t \v \f \r",
			[]Item{
				{FilePos{8, 1}, ItemEOF, ""},
			},
		},

		{"0",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "0"},
				{FilePos{2, 1}, ItemEOF, ""},
			},
		},

		{"01234567",
			[]Item{
				{FilePos{1, 1}, ItemOctal, "01234567"},
				{FilePos{9, 1}, ItemEOF, ""},
			},
		},

		{"0123456789",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "0123456789"},
				{FilePos{11, 1}, ItemEOF, ""},
			},
		},

		{"+123456",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "+123456"},
				{FilePos{8, 1}, ItemEOF, ""},
			},
		},

		{"123;+456\n-789",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "123"},
				{FilePos{4, 1}, ItemEOL, ";"},
				{FilePos{5, 1}, ItemDecimal, "+456"},
				{FilePos{9, 1}, ItemEOL, "\n"},
				{FilePos{1, 2}, ItemDecimal, "-789"},
				{FilePos{5, 2}, ItemEOF, ""},
			},
		},

		{"+",
			[]Item{
				{FilePos{1, 1}, ItemError, ErrorExpectedNumber},
			},
		},

		{"+-123456",
			[]Item{
				{FilePos{1, 1}, ItemError, ErrorExpectedNumber},
			},
		},

		{"1+2",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "1"},
				{FilePos{2, 1}, ItemOperator, "+"},
				{FilePos{3, 1}, ItemDecimal, "2"},
				{FilePos{4, 1}, ItemEOF, ""},
			},
		},

		{"12+34-56*78/9",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "12"},
				{FilePos{3, 1}, ItemOperator, "+"},
				{FilePos{4, 1}, ItemDecimal, "34"},
				{FilePos{6, 1}, ItemOperator, "-"},
				{FilePos{7, 1}, ItemDecimal, "56"},
				{FilePos{9, 1}, ItemOperator, "*"},
				{FilePos{10, 1}, ItemDecimal, "78"},
				{FilePos{12, 1}, ItemOperator, "/"},
				{FilePos{13, 1}, ItemDecimal, "9"},
				{FilePos{14, 1}, ItemEOF, ""},
			},
		},

		{"12 + 34 - 56 * 78 / 9",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "12"},
				{FilePos{4, 1}, ItemOperator, "+"},
				{FilePos{6, 1}, ItemDecimal, "34"},
				{FilePos{9, 1}, ItemOperator, "-"},
				{FilePos{11, 1}, ItemDecimal, "56"},
				{FilePos{14, 1}, ItemOperator, "*"},
				{FilePos{16, 1}, ItemDecimal, "78"},
				{FilePos{19, 1}, ItemOperator, "/"},
				{FilePos{21, 1}, ItemDecimal, "9"},
				{FilePos{22, 1}, ItemEOF, ""},
			},
		},

		{"+12+-34-+56*-78/+9",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "+12"},
				{FilePos{4, 1}, ItemOperator, "+"},
				{FilePos{5, 1}, ItemDecimal, "-34"},
				{FilePos{8, 1}, ItemOperator, "-"},
				{FilePos{9, 1}, ItemDecimal, "+56"},
				{FilePos{12, 1}, ItemOperator, "*"},
				{FilePos{13, 1}, ItemDecimal, "-78"},
				{FilePos{16, 1}, ItemOperator, "/"},
				{FilePos{17, 1}, ItemDecimal, "+9"},
				{FilePos{19, 1}, ItemEOF, ""},
			},
		},

		{"+12 + -34 - +56 * -78 / +9",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "+12"},
				{FilePos{5, 1}, ItemOperator, "+"},
				{FilePos{7, 1}, ItemDecimal, "-34"},
				{FilePos{11, 1}, ItemOperator, "-"},
				{FilePos{13, 1}, ItemDecimal, "+56"},
				{FilePos{17, 1}, ItemOperator, "*"},
				{FilePos{19, 1}, ItemDecimal, "-78"},
				{FilePos{23, 1}, ItemOperator, "/"},
				{FilePos{25, 1}, ItemDecimal, "+9"},
				{FilePos{27, 1}, ItemEOF, ""},
			},
		},

		{" 1+\n123",
			[]Item{
				{FilePos{2, 1}, ItemDecimal, "1"},
				{FilePos{3, 1}, ItemOperator, "+"},
				{FilePos{4, 1}, ItemError, ErrorExpectedNumber},
			},
		},

		{"-0X123",
			[]Item{
				{FilePos{1, 1}, ItemHexadecimal, "-0X123"},
				{FilePos{7, 1}, ItemEOF, ""},
			},
		},

		{"0x",
			[]Item{
				{FilePos{1, 1}, ItemError, ErrorHexadecimalNoDigits},
			},
		},

		{"+0x12 + -0X34 - +0x56 * -0X78 / +0x9",
			[]Item{
				{FilePos{1, 1}, ItemHexadecimal, "+0x12"},
				{FilePos{7, 1}, ItemOperator, "+"},
				{FilePos{9, 1}, ItemHexadecimal, "-0X34"},
				{FilePos{15, 1}, ItemOperator, "-"},
				{FilePos{17, 1}, ItemHexadecimal, "+0x56"},
				{FilePos{23, 1}, ItemOperator, "*"},
				{FilePos{25, 1}, ItemHexadecimal, "-0X78"},
				{FilePos{31, 1}, ItemOperator, "/"},
				{FilePos{33, 1}, ItemHexadecimal, "+0x9"},
				{FilePos{37, 1}, ItemEOF, ""},
			},
		},

		{"0.",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "0."},
				{FilePos{3, 1}, ItemEOF, ""},
			},
		},

		{"0.123",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "0.123"},
				{FilePos{6, 1}, ItemEOF, ""},
			},
		},

		{"1.2; 0. / 0XA * 0023",
			[]Item{
				{FilePos{1, 1}, ItemDecimal, "1.2"},
				{FilePos{4, 1}, ItemEOL, ";"},
				{FilePos{6, 1}, ItemDecimal, "0."},
				{FilePos{9, 1}, ItemOperator, "/"},
				{FilePos{11, 1}, ItemHexadecimal, "0XA"},
				{FilePos{15, 1}, ItemOperator, "*"},
				{FilePos{17, 1}, ItemOctal, "0023"},
				{FilePos{21, 1}, ItemEOF, ""},
			},
		},
	}

	for _, test = range tests {
		got = getItems(test.input)

		if !cmp(got, test.want) {
			t.Errorf("%q: \ngot: \n%s \nwant: \n%s\n", test.input, fmtItems(got), fmtItems(test.want))
		}
	}
}
