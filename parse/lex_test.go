package parse

import (
	"testing"
	"time"
)

func TestLexItems(t *testing.T) {
	var tests = []struct {
		name     string
		file     string
		expected []itemType
	}{
		{
			name: "handles comment",
			file: "%abc-2.1",
			expected: []itemType{
				itemPercent,
				itemComment,
				itemEOF,
			},
		},
		{
			name: "handles basic headers",
			file: `X:1
M:4/4
O:Irish
R:Reel`,
			expected: []itemType{
				itemFieldName, itemNumber, itemNewline,
				itemFieldName, itemNumber, itemDivide, itemNumber, itemNewline,
				itemFieldName, itemString, itemNewline,
				itemFieldName, itemString,
				itemEOF,
			},
		},
		{
			name: "handles body line",
			file: `eg|a2ab ageg|`,
			expected: []itemType{
				itemLetter, itemLetter, itemBarline,
				itemLetter, itemNumber, itemLetter, itemLetter, itemSpace,
				itemLetter, itemLetter, itemLetter, itemLetter, itemBarline,
				itemEOF,
			},
		},
		{
			name: "handles headers -> body",
			file: `M:4/4
O:Irish
R:Reel
eg|a2ab ageg|`,
			expected: []itemType{
				itemFieldName, itemNumber, itemDivide, itemNumber, itemNewline,
				itemFieldName, itemString, itemNewline,
				itemFieldName, itemString, itemNewline,
				itemLetter, itemLetter, itemBarline,
				itemLetter, itemNumber, itemLetter, itemLetter, itemSpace,
				itemLetter, itemLetter, itemLetter, itemLetter, itemBarline,
				itemEOF,
			},
		},
		{
			name: "ignores escaped newlines",
			file: `eg|a21ab\
ageg|`,
			expected: []itemType{
				itemLetter, itemLetter, itemBarline,
				itemLetter, itemNumber, itemLetter, itemLetter,
				itemLetter, itemLetter, itemLetter, itemLetter, itemBarline,
				itemEOF,
			},
		},
		{
			name: "distinguishes between chords and annotations",
			file: `"Am" ">Annotation"`,
			expected: []itemType{
				itemChord, itemSpace, itemAnnotationPosition, itemAnnotation,
				itemEOF,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := lex(test.name, test.file)
			for _, expected := range test.expected {
				select {
				case got := <-l.items:
					if got.typ != expected {
						t.Errorf("expected type %d, got %d", expected, got.typ)
					} else {
						t.Logf("got %d", got.typ)
					}
				case <-time.After(time.Second):
					t.Errorf("Ran out of items, expecting %d", expected)
				}
			}
			for got := range l.items {
				t.Errorf("unexpected token '%v'", got)
			}
		})
	}
}
