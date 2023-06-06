package reference

import "fmt"

// Reference contains information about a file and a position within that file
type Reference struct {
	FileName string // The path to the file
	Position Pos    // The position represented
}

// Pos represents the position of a token in the input
type Pos struct {
	Start   int
	End     int
	Line    int
	Endline int
}

func NewReference(fileName string, pos *Pos) *Reference {
	return &Reference{fileName, *pos}
}

func NewPos(values ...int) *Pos {
	if len(values) > 4 {
		panic("must provide less than 4 values")
	} else if len(values) < 2 {
		values = append(values, 0)
		values = append(values, 0)
	}

	var line, endline int
	if len(values) == 3 {
		line = values[2]
	} else if len(values) == 4 {
		line = values[2]
		endline = values[3]
	}

	return &Pos{values[0], values[1], line, endline}
}

func (self Pos) String() string {
	return fmt.Sprintf("Pos(start [%d] end [%d] line [%d] endline [%d])", self.Start, self.End, self.Line, self.Endline)
}

func (self Reference) String() string {
	return fmt.Sprintf("%s at %d:%d", self.FileName, self.Position.Line, self.Position.Start)
}