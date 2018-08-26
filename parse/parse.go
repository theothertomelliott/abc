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
	switch item.val {
	case string(headerX):
		return p.setSequence()
	case string(headerT):
		// TODO
	case string(headerN):
		// TODO
	case string(headerM):
		return p.setMeter()
	case string(headerL):
		// TODO
	case string(headerK):
		// TODO
	case string(headerC):
		// TODO
	case string(headerO):
		return p.setOrigin()
	case string(headerR):
		return p.setRhythm()
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

func (p *parser) setRhythm() error {
	item, err := p.expect(itemString)
	if err != nil {
		return err
	}
	p.currentTune.Rhythm = item.val
	return p.expectNewline()
}

func (p *parser) setOrigin() error {
	item, err := p.expect(itemString)
	if err != nil {
		return err
	}
	p.currentTune.Origin = item.val
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
