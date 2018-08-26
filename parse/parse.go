package parse

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/theothertomelliott/abc"
)

// Read parses an input stream into a sequence of abc.Tune objects.
// An error is returned in the event the stream cannot be parsed.
func Read(in io.Reader) ([]abc.Tune, error) {
	file, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}
	l := lex("filename", string(file))
	parser := &parser{
		lexer: l,
	}
	return parser.parse()
}

type parser struct {
	lexer       lexingResult
	tunes       []abc.Tune
	currentTune *abc.Tune
}

type lexingResult interface {
	nextItem() *item
	drain()
}

func (p *parser) parse() ([]abc.Tune, error) {
	for true {
		item := p.lexer.nextItem()
		if item == nil {
			break
		}
		err := p.handleItem(item)
		if err != nil {
			return nil, err
		}
	}
	if p.currentTune != nil {
		p.tunes = append(p.tunes, *p.currentTune)
	}
	return p.tunes, nil
}

func (p *parser) handleItem(item *item) error {
	if item.typ == itemFieldName {
		err := p.handleFieldName(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) handleFieldName(item *item) error {
	var err error
	switch item.val {
	case string(headerA):
		p.currentTune.Area, err = p.expectString()
		return err
	case string(headerB):
		p.currentTune.Book, err = p.expectString()
		return err
	case string(headerC):
		p.currentTune.Composer, err = p.expectString()
		return err
	case string(headerD):
		p.currentTune.Discography, err = p.expectString()
		return err
	case string(headerF):
		p.currentTune.FileURL, err = p.expectString()
		return err
	case string(headerG):
		p.currentTune.Group, err = p.expectString()
		return err
	case string(headerH):
		return p.addHistory()
	case string(headerI):
		// TODO
		return p.consumeToNewline()
	case string(headerK):
		// TODO
		return p.consumeToNewline()
	case string(headerL):
		return p.setNoteLength()
	case string(headerM):
		return p.setMeter()
	case string(headerm):
		// TODO
		return p.consumeToNewline()
	case string(headerN):
		return p.addNotes()
	case string(headerO):
		p.currentTune.Origin, err = p.expectString()
		return err
	case string(headerP):
		// TODO
		return p.consumeToNewline()
	case string(headerQ):
		// TODO
		return p.consumeToNewline()
	case string(headerR):
		p.currentTune.Rhythm, err = p.expectString()
		return err
	case string(headerr):
		// Ignore remarks
		return p.consumeToNewline()
	case string(headerS):
		p.currentTune.Source, err = p.expectString()
		return err
	case string(headers):
		// TODO
		return p.consumeToNewline()
	case string(headerT):
		p.currentTune.Title, err = p.expectString()
		return err
	case string(headerU):
		// TODO
		return p.consumeToNewline()
	case string(headerV):
		// TODO
		return p.consumeToNewline()
	case string(headerW):
		return p.addWordsAfterTune()
	case string(headerw):
		// TODO
		return p.consumeToNewline()
	case string(headerX):
		return p.setSequence()
	case string(headerZ):
		p.currentTune.Transcription, err = p.expectString()
		return err
	}
	return errors.New("unhandled")
}

func (p *parser) consumeToNewline() error {
	for item := p.lexer.nextItem(); item != nil && item.typ != itemNewline; item = p.lexer.nextItem() {
	}
	return nil
}

func (p *parser) expectNewline() error {
	_, err := p.expect(itemNewline)
	return err
}

func (p *parser) expect(types ...itemType) (*item, error) {
	item := p.lexer.nextItem()
	if item == nil {
		return nil, fmt.Errorf("expected one of %v, got nil token", types)
	}
	for _, t := range types {
		if item.typ == t {
			return item, nil
		}
	}
	return nil, fmt.Errorf("expected one of %v, got nil token", types)
}

func (p *parser) setMeter() error {
	meter := abc.Meter{}
	// Numerator
	for item := p.lexer.nextItem(); item != nil && item.typ != itemDivide; item = p.lexer.nextItem() {
		switch item.typ {
		case itemNumber:
			numeratorValue, _ := strconv.Atoi(item.val)
			meter.Numerator = append(meter.Numerator, numeratorValue)
		case itemPlus:
		default:
			return fmt.Errorf("expected number or plus, got %v", item)
		}
	}

	// Denominator
	item, err := p.expect(itemNumber)
	if err != nil {
		return err
	}
	meter.Denominator, _ = strconv.Atoi(item.val)

	p.currentTune.Meter = meter
	return p.consumeToNewline()
}

func (p *parser) setNoteLength() error {
	noteLength := abc.NoteLength{}

	// Numerator
	item, err := p.expect(itemNumber)
	if err != nil {
		return err
	}
	noteLength.Numerator, _ = strconv.Atoi(item.val)

	// Divide separator
	item, err = p.expect(itemDivide)
	if err != nil {
		return err
	}

	// Denominator
	item, err = p.expect(itemNumber)
	if err != nil {
		return err
	}
	noteLength.Denominator, _ = strconv.Atoi(item.val)

	p.currentTune.NoteLength = noteLength
	return p.consumeToNewline()
}

func (p *parser) expectString() (string, error) {
	item, err := p.expect(itemString)
	if err != nil {
		return "", err
	}
	return item.val, p.expectNewline()
}

func (p *parser) addNotes() error {
	item, err := p.expect(itemString)
	if err != nil {
		return err
	}
	p.currentTune.Comments = append(p.currentTune.Comments, item.val)
	return p.expectNewline()
}

func (p *parser) addWordsAfterTune() error {
	item, err := p.expect(itemString)
	if err != nil {
		return err
	}
	p.currentTune.WordsAfterTune = append(p.currentTune.WordsAfterTune, item.val)
	return p.expectNewline()
}

func (p *parser) addHistory() error {
	item, err := p.expect(itemString)
	if err != nil {
		return err
	}
	p.currentTune.History = append(p.currentTune.History, item.val)
	return p.expectNewline()
}

func (p *parser) setSequence() error {
	item, err := p.expect(itemNumber)
	if err != nil {
		return err
	}
	if p.currentTune != nil {
		p.tunes = append(p.tunes, *p.currentTune)
	}
	sequenceNum, _ := strconv.Atoi(item.val)
	p.currentTune = &abc.Tune{
		Sequence: sequenceNum,
	}
	return p.expectNewline()
}
