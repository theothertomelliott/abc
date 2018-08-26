package abc

type Tune struct {
	Area           string // deprecated
	Book           string
	Composer       string
	Discography    string
	FileURL        string
	History        []string
	Group          string
	Sequence       int
	Title          string
	Comments       []string
	Rhythm         string
	Origin         string
	Meter          Meter
	NoteLength     NoteLength
	Key            Key
	Source         string
	Transcription  string
	WordsAfterTune []string

	Bars []Bar
}

type Bar struct {
	Left     BarLine
	Notation []Notation
	Right    BarLine
}

type BarLine struct {
	Repeat int
}

type Notation interface {
	Length() int
}

type Key string

type Meter struct {
	Numerator   []int
	Denominator int
}

type NoteLength struct {
	Numerator   int
	Denominator int
}
