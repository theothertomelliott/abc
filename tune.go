package abc

type Tune struct {
	Sequence   int
	Title      string
	Composer   string
	Comments   []string
	Rhythm     string
	Origin     string
	Meter      Meter
	NoteLength float64
	Key        Key

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
