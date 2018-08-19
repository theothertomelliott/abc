package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemFieldName
	itemColon
	itemString
	itemURL
	itemUnit
	itemKey
	itemMeter
	itemMacro

	itemVoice
	itemOpenParen
	itemCloseParen
	itemLetter
	itemNumber
	itemDivide
	itemPlus
	itemMinus
	itemEquals

	itemSharp
	itemNatural
	itemFlat

	itemMinor
	itemExclamation
	itemStar

	itemPercent

	itemDottedBarline
	itemBarline
	itemThinThickDoubleBarLine
	itemThinThinDoubleBarLine
	itemThickThinDoubleBarLine

	itemStartRepeat
	itemEndRepeat
	itemStartEndRepeats

	itemQuote

	itemNewline
	itemSpace
	itemBackslash
	itemGreaterThan
	itemLessThan

	itemComment

	itemInvisibleRest
	itemRest
	itemMultiMeasureRest

	itemChord

	itemAnnotationPosition
	itemAnnotation

	itemVariantNumber
	itemVariantComma
	itemVariantRange

	itemEOF
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string     // the name of the input; used only for error reports
	input      string     // the string being scanned
	pos        Pos        // current position in the input
	start      Pos        // start position of this item
	width      Pos        // width of last rune read from input
	items      chan *item // channel of scanned items
	parenDepth int        // nesting depth of ( ) exprs
	line       int        // 1+number of newlines seen
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- &item{t, l.start, l.input[l.start:l.pos], l.line}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- &item{itemError, l.start, fmt.Sprintf(format, args...), l.line}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() *item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan *item),
		line:  1,
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexLine; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// state functions

const (
	headerX = 'X'
	headerT = 'T'
	headerN = 'N'
	headerM = 'M'
	headerL = 'L'
	headerK = 'K'
	headerC = 'C'
	headerO = 'O'
	headerR = 'R'

	colon = ":"
)

func (l *lexer) ignoreWhitespace() {
	for true {
		r := l.peek()
		if isSpace(r) {
			l.pos++
			l.ignore()
		} else {
			return
		}
	}
}

func (l *lexer) consumeToEndOfLine() {
	for true {
		r := l.peek()
		if !isEndOfLine(r) && !(r == eof) {
			l.pos++
		} else {
			return
		}
	}
}

func (l *lexer) acceptNewline() {
	r := l.next()
	if r == eof {
		return
	}
	if !isEndOfLine(r) {
		l.errorf("expected newline, got %v", r)
		return
	}
	l.emit(itemNewline)
}

func (l *lexer) ignoreNewline() {
	r := l.next()
	if r == eof {
		return
	}
	if !isEndOfLine(r) {
		l.errorf("expected newline, got %v", r)
		return
	}
	l.ignore()
}

func lexLine(l *lexer) stateFn {
	if l.peek() == eof {
		l.emit(itemEOF)
		return nil
	}
	if strings.HasPrefix(l.input[l.pos+1:], colon) {
		return lexHeaderLine
	}
	return lexBodyLine
}

func lexBodyLine(l *lexer) stateFn {
	if isEndOfLine(l.peek()) {
		l.acceptNewline()
		return lexLine
	}

	if unicode.IsDigit(l.peek()) {
		l.acceptDecimalRun()
		l.emit(itemNumber)
		return lexLine
	}

	if len(l.input)-2 > int(l.pos) {
		if strings.HasPrefix(l.input[l.pos+1:], "|]") {
			l.pos += 2
			l.emit(itemThinThickDoubleBarLine)
			return lexBodyLine
		}
		if strings.HasPrefix(l.input[l.pos+1:], "||") {
			l.pos += 2
			l.emit(itemThinThinDoubleBarLine)
			return lexBodyLine
		}
		if strings.HasPrefix(l.input[l.pos+1:], "[|") {
			l.pos += 2
			l.emit(itemThickThinDoubleBarLine)
			return lexBodyLine
		}
		if strings.HasPrefix(l.input[l.pos+1:], "|:") {
			l.pos += 2
			l.emit(itemStartRepeat)
			return lexBodyLine
		}
		if strings.HasPrefix(l.input[l.pos+1:], ":|") {
			l.pos += 2
			l.emit(itemEndRepeat)
			return lexBodyLine
		}
		if strings.HasPrefix(l.input[l.pos+1:], "::") ||
			strings.HasPrefix(l.input[l.pos+1:], ":|:") ||
			strings.HasPrefix(l.input[l.pos+1:], ":||:") {
			l.pos += 2
			l.emit(itemStartEndRepeats)
			return lexBodyLine
		}
		if strings.HasPrefix(l.input[l.pos+1:], ".|") {
			l.pos += 2
			l.emit(itemDottedBarline)
			return lexBodyLine
		}
		// TODO: Support multiple repeats

	}

	c := l.next()
	switch {
	case c == '%':
		l.emit(itemPercent)
		return lexComment
	case isSpace(c):
		l.emit(itemSpace)
	case c == 'x':
		l.emit(itemInvisibleRest)
	case c == 'z':
		l.emit(itemRest)
	case c == 'Z':
		l.emit(itemMultiMeasureRest)
	case unicode.IsLetter(c):
		l.emit(itemLetter)
	case c == '^':
		l.emit(itemSharp)
	case c == '=':
		l.emit(itemNatural)
	case c == '_':
		l.emit(itemFlat)
	case c == '/':
		l.emit(itemDivide)
	case c == '\\':
		l.ignore()
		l.ignoreNewline()
	case c == '"':
		l.ignore()
		return lexChord
	case c == '|':
		l.emit(itemBarline)

	case c == '+':
		l.emit(itemPlus)
	case c == '<':
		l.emit(itemLessThan)
	case c == '>':
		l.emit(itemGreaterThan)
	case c == '-':
		l.emit(itemMinus)
	case c == '[':
		if unicode.IsDigit(l.peek()) {
			l.ignore()
			return lexVariant
		}
		l.errorf("unexpected character after [")
	case c == ':':
		l.emit(itemColon)
	case c == eof:
		l.emit(itemEOF)
		return nil
	default:
		l.errorf("unknown character: %c", c)
		return nil
	}
	return lexBodyLine
}

func lexVariant(l *lexer) stateFn {
	l.acceptDecimalRun()
	l.emit(itemVariantNumber)

	if l.peek() == ',' {
		l.pos++
		l.emit(itemVariantComma)
		return lexVariant
	}

	if l.peek() == '-' {
		l.pos++
		l.emit(itemVariantRange)
		return lexVariant
	}

	return lexLine
}

func lexChord(l *lexer) stateFn {
	a := l.peek()
	if a == '^' ||
		a == '_' ||
		a == '<' ||
		a == '>' ||
		a == '@' {
		l.pos++
		l.emit(itemAnnotationPosition)
		return lexAnnotation
	}

	for true {
		r := l.peek()
		if r == '"' {
			break
		}
		l.pos++
	}
	l.emit(itemChord)
	l.pos++
	l.ignore()
	return lexBodyLine
}

func lexAnnotation(l *lexer) stateFn {
	for true {
		r := l.peek()
		if r == '"' {
			break
		}
		l.pos++
	}
	l.emit(itemAnnotation)
	l.pos++
	l.ignore()
	return lexBodyLine
}

func lexComment(l *lexer) stateFn {
	l.consumeToEndOfLine()
	l.emit(itemComment)
	l.acceptNewline()
	return lexLine
}

func lexHeaderLine(l *lexer) stateFn {
	fieldName := l.next()

	l.emit(itemFieldName)
	// Skip the colon
	l.pos++
	l.ignore()

	switch fieldName {
	case headerX:
		return lexHeaderInt
	case headerT, headerN:
		return lexHeaderString
	case headerM:
		return lexHeaderMeter
	case headerL:
		return lexHeaderNoteLength
	case headerK:
		// TODO: Implement something for the key
		return lexHeaderString
	case headerC:
		return lexHeaderString
	case headerO:
		return lexHeaderString
	case headerR:
		return lexHeaderString
	default:
		l.errorf("unknown header field %s", string(fieldName))
		return nil
	}
}

func lexNextLine(l *lexer) stateFn {
	l.ignoreWhitespace()
	l.acceptNewline()
	return lexLine
}

func lexHeaderInt(l *lexer) stateFn {
	l.ignoreWhitespace()
	l.acceptDecimalRun()
	l.emit(itemNumber)
	return lexNextLine
}

func lexHeaderString(l *lexer) stateFn {
	l.ignoreWhitespace()
	l.consumeToEndOfLine()
	l.emit(itemString)
	return lexNextLine
}

func lexHeaderMeter(l *lexer) stateFn {
	l.ignoreWhitespace()
	c := l.peek()
	for !isEndOfLine(c) {
		l.pos++
		switch {
		case unicode.IsDigit(c):
			l.acceptDecimalRun()
			l.emit(itemNumber)
		case c == '(':
			l.emit(itemOpenParen)
		case c == ')':
			l.emit(itemCloseParen)
		case c == '+':
			l.emit(itemPlus)
		case c == '/':
			l.emit(itemDivide)
		}
		c = l.peek()
	}
	return lexNextLine
}

func lexHeaderNoteLength(l *lexer) stateFn {
	l.ignoreWhitespace()
	c := l.peek()
	for !isEndOfLine(c) {
		l.pos++
		switch {
		case unicode.IsLetter(c):
			return lexHeaderString
		case unicode.IsDigit(c):
			l.acceptDecimalRun()
			l.emit(itemNumber)
		case c == '/':
			l.emit(itemDivide)
		default:
			l.errorf("unknown character: %c", c)
			return nil
		}
		c = l.peek()
	}
	return lexNextLine
}

func (l *lexer) acceptDecimalRun() {
	l.acceptRun("0123456789")
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
