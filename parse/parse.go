package parse

import (
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
		// TODO
	case string(headerL):
		// TODO
	case string(headerK):
		// TODO
	case string(headerC):
		// TODO
	case string(headerO):
		// TODO
	case string(headerR):
		// TODO
	}
	return nil
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

func (p *parser) setRhythm() error {
	item, err := p.expect(itemString)
	if err != nil {
		return err
	}
	sequenceNum, _ := strconv.Atoi(item.val)
	p.currentTune = &abc.Tune{
		Sequence: sequenceNum,
	}
	return nil
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
	return nil
}
