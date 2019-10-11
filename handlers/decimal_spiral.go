package handlers

import (
	"errors"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/unswpcsoc/PCSocBot/commands"
	"github.com/unswpcsoc/PCSocBot/internal/utils"
)

const (
	lowerLimit = 5
	upperLimit = 43
)

var (
	// ErrdecimalSpiralRange means user gave invalid input outside of range
	ErrdecimalSpiralRange = errors.New("size is not an odd integer between " + strconv.Itoa(lowerLimit) + " and " + strconv.Itoa(upperLimit))
)

type decimalSpiral struct {
	nilCommand
	Size int `arg:"size"`
}

func newDecimalSpiral() *decimalSpiral { return &decimalSpiral{} }

func (d *decimalSpiral) Aliases() []string { return []string{"decimalspiral", "ds"} }

func (d *decimalSpiral) Desc() string {
	return "Generate a decimal spiral. Size must be an odd integer between " +
		strconv.Itoa(lowerLimit) + " and " + strconv.Itoa(upperLimit)
}

func (d *decimalSpiral) MsgHandle(ses *discordgo.Session, msg *discordgo.Message) (*commands.CommandSend, error) {
	if d.Size%2 == 0 || d.Size < lowerLimit || d.Size > upperLimit {
		return nil, ErrdecimalSpiralRange
	}

	return commands.NewSimpleSend(msg.ChannelID, utils.Block(genDecimalSpiral(d.Size))), nil
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// genDecimalSpiral consists of 2 steps:
// 1. Loop left to right, up to down, calling isDigit to determine
// whether to print a digit or a dash for each coordinate
// 2. If a digit should be printed, call getDigit to calculate
// the digit to be printed for that coordinate
// Returns the decimal spiral as a string
func genDecimalSpiral(size int) string {
	out := ""
	row := 0
	for row < size {
		col := 0
		for col < size {
			if isDigit(size, row, col) {
				out += strconv.Itoa(getDigit(size, row, col) % 10)
			} else {
				out += "-"
			}
			col++
		}
		out += "\n"
		row++
	}

	return out
}

// isDigit determines whether a coordinate on a square is part of
// the spiral of digits and returns true if it is, and false if not
func isDigit(size int, row int, col int) bool {
	// Absolute row distance from midpoint
	rowDist := abs(row - size/2)
	// Absolute column distance from midpoint
	colDist := abs(col - size/2)
	isDigit := false

	if size%4 == 1 {
		// Type 1 spiral (digit in centre)
		if row <= size/2 && col < size/2 && row == col+1 {
			// Special handling to turn boxes into spirals
			if rowDist%2 == 0 {
				isDigit = true
			}
		} else if colDist >= rowDist && colDist%2 == 0 {
			isDigit = true
		} else if colDist < rowDist && rowDist%2 == 0 {
			isDigit = true
		}
	} else {
		// Type 2 spiral (no digit in centre)
		if row <= size/2 && col < size/2 && row == col+1 {
			// Special handling to turn boxes into spirals
			if rowDist%2 == 1 {
				isDigit = true
			}
		} else if colDist >= rowDist && colDist%2 != 0 {
			isDigit = true
		} else if colDist < rowDist && rowDist%2 != 0 {
			isDigit = true
		}
	}

	return isDigit
}

// getDigit calculates the digit to print at a coordinate of a
// decimal spiral and returns that digit
//
// It works by splitting the spiral into triangular quadrants
//   *******
// *  *****  *
// **  ***  **
// ***  *  ***
// **  ***  **
// *  *****  *
//   *******
//
// For the top quadrant, observe the following digits
// 6**************
// --------------*
// **0**********-*
// *-----------*-*
// *-**0******-*-*
// *-*-------*-*-*
// *-*-**6**-*-*-*
// *-*-*---*-*-*-*
// *-*-*-***-*-*-*
// *-*-*-----*-*-*
// *-*-*******-*-*
// *-*---------*-*
// *-***********-*
// *-------------*
// ***************
//
// The actual numbers at the location of these digits form a
// quadratic sequence
// 6  30  70  126
//  24  40  56
//	16  16
// Where the differences between the differences is 16.
// Using this, you can make a quadratic equation for the top left
// corners of each box, and subtract the current column to find the
// digits for coordinates to the right
//
// You'll need to come up with a different quadratic equation for each
// quadrant. Since there are 2 types of spiral (last digit in the
// centre, last digit off-centre), that's a total of 8 different
// quadratic equations. Use the current column or current row
// accordingly to determine how much to add or subtract from the
// corner values
func getDigit(size int, row int, col int) int {
	// Special handling for box segments modified to be spirals
	if row <= size/2 && col < size/2 && row == col+1 {
		row = row - 1
		col = col - 1
	}

	// Row displacement from midpoint
	rowDist := row - size/2

	// Absolute row distance from midpoint
	absRowDist := abs(rowDist)

	// Column displacement from midpoint
	colDist := col - size/2

	// Absolute column distance from midpoint
	absColDist := abs(colDist)

	// Size of box current coordinate is on
	subSize := 0

	if absRowDist >= absColDist {
		subSize = 2*absRowDist + 1
	} else {
		subSize = 2*absColDist + 1
	}

	row = row - (size-subSize)/2
	col = col - (size-subSize)/2
	// Layer of current box. 0 is centre
	layer := (subSize + 1) / 4

	if rowDist <= 0 && absRowDist >= absColDist {
		// Top quadrant
		if subSize%4 == 1 {
			// Type 1 boxes
			return 8*layer*layer + 8*layer - col
		}
		return 8*layer*layer - 2 - col
	} else if colDist > 0 && absColDist > absRowDist {
		// Right quadrant
		if subSize%4 == 1 {
			// Type 1 boxes
			return 8*layer*layer + 4*layer - row
		}
		// Type 2 boxes
		return 8*layer*layer - 4*layer - row
	} else if rowDist > 0 && absRowDist >= absColDist {
		// Bottom quadrant
		if subSize%4 == 1 {
			// Type 1 boxes
			return 8*layer*layer - 4*layer + col
		}
		// Type 2 boxes
		return 8*layer*layer - 12*layer + 4 + col
	} else {
		// Left quadrant
		if subSize%4 == 1 {
			// Type 1 boxes
			return 8*layer*layer - 8*layer + row
		}
		// Type 2 boxes
		return 8*layer*layer - 16*layer + 6 + row
	}
}
