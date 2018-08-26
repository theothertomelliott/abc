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
			name: "parses headers appropriately",
			input: []*item{
				i(itemFieldName, "X"), i(itemNumber, "1"), newline(),
				i(itemFieldName, "M"), i(itemNumber, "4"), i(itemPlus, "+"), i(itemNumber, "1"), i(itemDivide, "/"), i(itemNumber, "4"), newline(),
				i(itemFieldName, "O"), i(itemString, "Irish"), newline(),
				i(itemFieldName, "N"), i(itemString, "First comment line"), newline(),
				i(itemFieldName, "R"), i(itemString, "Reel"), newline(),
				i(itemFieldName, "T"), i(itemString, "Wild Irish Rose"), newline(),
				i(itemFieldName, "N"), i(itemString, "Second comment line"), newline(),
				i(itemFieldName, "r"), i(itemString, "this is remark"), newline(),
				i(itemFieldName, "L"), i(itemNumber, "1"), i(itemDivide, "/"), i(itemNumber, "4"), newline(),
				i(itemFieldName, "C"), i(itemString, "Composer name"), newline(),
				i(itemFieldName, "A"), i(itemString, "Area line"), newline(),
				i(itemFieldName, "B"), i(itemString, "O'Neills"), newline(),
				i(itemFieldName, "D"), i(itemString, "Chieftains IV"), newline(),
				i(itemFieldName, "F"), i(itemString, "http://a.b.c/file.abc"), newline(),
				i(itemFieldName, "G"), i(itemString, "flute"), newline(),
				i(itemFieldName, "H"), i(itemString, "this usage is considered as two entries"), newline(),
				i(itemFieldName, "H"), i(itemString, "rather than one"), newline(),
				i(itemFieldName, "S"), i(itemString, "collected in Brittany"), newline(),
				i(itemFieldName, "W"), i(itemString, "lyrics after the tune body"), newline(),
				i(itemFieldName, "Z"), i(itemString, "John Smith, <j.s@mail.com>"), newline(),
			},
			expected: []abc.Tune{
				abc.Tune{
					Sequence: 1,
					Meter: abc.Meter{
						Numerator:   []int{4, 1},
						Denominator: 4,
					},
					NoteLength: abc.NoteLength{
						Numerator:   1,
						Denominator: 4,
					},
					Rhythm:         "Reel",
					Origin:         "Irish",
					Title:          "Wild Irish Rose",
					Comments:       []string{"First comment line", "Second comment line"},
					Composer:       "Composer name",
					Area:           "Area line",
					Book:           "O'Neills",
					Discography:    "Chieftains IV",
					FileURL:        "http://a.b.c/file.abc",
					Group:          "flute",
					History:        []string{"this usage is considered as two entries", "rather than one"},
					Source:         "collected in Brittany",
					Transcription:  "John Smith, <j.s@mail.com>",
					WordsAfterTune: []string{"lyrics after the tune body"},
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
