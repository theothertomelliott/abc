package parse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/theothertomelliott/abc"
)

func i(t itemType, val string) *item {
	return &item{
		typ: t,
		val: val,
	}
}

func newline() *item {
	return &item{
		typ: itemNewline,
		val: "\n",
	}
}

func TestParse(t *testing.T) {

	var tests = []struct {
		name     string
		input    []*item
		expected []abc.Tune
	}{
		{
			name: "empty file",
		},
		{
			name: "empty file",
			input: []*item{
				i(itemFieldName, "X"), i(itemNumber, "1"), newline(),
				i(itemFieldName, "M"), i(itemNumber, "4"), i(itemDivide, "/"), i(itemNumber, "4"), newline(),
				i(itemFieldName, "O"), i(itemString, "Irish"), newline(),
				i(itemFieldName, "R"), i(itemString, "Reel"), newline(),
			},
			expected: []abc.Tune{
				abc.Tune{
					Sequence: 1,
					Meter: abc.Meter{
						Numerator:   []int{4},
						Denominator: 4,
					},
					Rhythm: "Reel",
					Origin: "Irish",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lexer := &itemSource{
				items: test.input,
			}
			p := parser{
				lexer: lexer,
			}
			got, err := p.parse()
			if !cmp.Equal(test.expected, got) {
				t.Errorf("tunes did not match: %v", cmp.Diff(test.expected, got))
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

type itemSource struct {
	items []*item
	pos   int
}

func (i *itemSource) nextItem() *item {
	if i.pos >= len(i.items) {
		return nil
	}
	index := i.pos
	i.pos++
	return i.items[index]
}

func (i *itemSource) drain() {}
