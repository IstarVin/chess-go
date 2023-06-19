package engine

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	DefaultFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

func NewGameChess() *Chess {
	chess, _ := NewChessGameWithFen(DefaultFen)
	return chess
}

func NewChessGameWithFen(fen string) (*Chess, error) {
	var chess Chess
	err := chess.decodeFen(fen)
	if err != nil {
		return nil, err
	}
	return &chess, nil
}

type CastleAvailability struct {
	WhiteKing  bool
	WhiteQueen bool
	BlackKing  bool
	BlackQueen bool
}

type Coords struct {
	row int
	col int
}

type Chess struct {
	boardTable  [8][8]rune
	turn        string
	castle      CastleAvailability
	pawnPassant string
	halfmoves   int
	fullmoves   int
}

// PrintBoard prints the board
func (c *Chess) PrintBoard() {
	for _, row := range c.boardTable {
		for _, content := range row {
			print(string(content), " ")
		}
		println()
	}
}

// GetFEN returns the FEN string of the current chess game
func (c *Chess) GetFEN() string {
	var fen string

	// Board
	for i, row := range c.boardTable {
		var spaceCount int
		for _, content := range row {
			if content == '-' {
				spaceCount++
				continue
			}

			if spaceCount > 0 {
				fen += strconv.Itoa(spaceCount)
				spaceCount = 0
			}

			fen += string(content)
		}

		if spaceCount > 0 {
			fen += strconv.Itoa(spaceCount)
		}

		if i < len(c.boardTable)-1 {
			fen += "/"
		}
	}

	// Turn
	fen += " " + c.turn + " "

	// Castle
	if c.castle.WhiteKing {
		fen += "K"
	}
	if c.castle.WhiteQueen {
		fen += "Q"
	}
	if c.castle.BlackKing {
		fen += "k"
	}
	if c.castle.BlackQueen {
		fen += "q"
	}

	// Pawn Passant
	fen += " " + c.pawnPassant

	// Half Moves
	fen += " " + strconv.Itoa(c.halfmoves)

	// Full Moves
	fen += " " + strconv.Itoa(c.fullmoves)

	return fen
}

// determineColor returns the color of the entered piece
func (c *Chess) determineColor(piece rune) rune {
	if piece == '-' {
		return piece
	}

	if piece == unicode.ToUpper(piece) {
		return 'w'
	} else {
		return 'b'
	}
}

func (c *Chess) determinePieceWithCoords(coord *Coords) rune {
	if coord.row < 0 || coord.row > 7 || coord.col < 0 || coord.col > 7 {
		return '-'
	}

	return c.boardTable[coord.row][coord.col]
}

// translateCBtoCoords translates chessboard notation to coordinates
func (c *Chess) translateCBtoCoords(cb string) *Coords {
	column := rune(cb[0])
	// Check if column is within range
	if column > 104 || column < 97 {
		return nil
	}

	row := rune(cb[1])
	// Check if row is within range
	if row > 56 || row < 49 {
		return nil
	}

	return &Coords{int('8' - row), int(column - 'a')}
}

// translateCoordsToCB translates coordinates to chessboard notation
func (c *Chess) translateCoordsToCB(coord *Coords) string {
	if coord.row < 0 || coord.row > 7 || coord.col < 0 || coord.col > 7 {
		return ""
	}

	return string('a'+rune(coord.col)) + string('8'-rune(coord.row))
}

// decodeFen decodes a FEN string into the chess struct
func (c *Chess) decodeFen(fen string) error {
	var err error

	splitFen := strings.Split(fen, " ")
	if len(splitFen) != 6 {
		return &FENError{err: "Lacks parameters"}
	}

	// Board Table
	rows := strings.Split(splitFen[0], "/")
	if len(rows) != 8 {
		return &FENError{err: "lacks row in board parameter"}
	}

	for row, rowContent := range rows {
		currentColumn := 0
		for _, columnContent := range rowContent {
			column := int(columnContent - '0')
			if column < 1 || column > 8 {
				column = 1
			} else {
				columnContent = '-'
			}

			for i := 0; i < column; i++ {
				col := currentColumn + i
				if col > 7 {
					break
				}

				c.boardTable[row][col] = columnContent
			}

			if column > 1 {
				currentColumn += column
			} else {
				currentColumn++
			}
		}
		if currentColumn != 8 {
			return &FENError{err: "invalid board parameter"}
		}
	}

	// Turn
	c.turn = splitFen[1]

	// Castle
	for _, castleAble := range splitFen[2] {
		if castleAble == 'K' {
			c.castle.WhiteKing = true
		} else if castleAble == 'Q' {
			c.castle.WhiteQueen = true
		} else if castleAble == 'k' {
			c.castle.BlackKing = true
		} else if castleAble == 'q' {
			c.castle.BlackQueen = true
		}
	}

	// Pawn Passant
	c.pawnPassant = splitFen[3]
	if c.translateCBtoCoords(c.pawnPassant) == nil && c.pawnPassant != "-" {
		c.pawnPassant = "-"
		return &FENError{err: "invalid pawn passant"}
	}

	// Half Moves
	c.halfmoves, err = strconv.Atoi(splitFen[4])
	if err != nil {
		return &FENError{err: "invalid half moves"}
	}

	// Full Moves
	c.fullmoves, err = strconv.Atoi(splitFen[5])
	if err != nil {
		return &FENError{err: "invalid full moves"}
	}

	return nil
}

// calculateValidPaths calculates the valid paths in a given piece coordinate
func (c *Chess) calculateValidPaths(coord *Coords) []*Coords {
	piece := c.determinePieceWithCoords(coord)
	color := c.determineColor(piece)

	var validMoves []*Coords

	switch unicode.ToLower(piece) {
	// Pawn
	case 'p':
		// Determine the startingRow, enpassantRow, numOfMoves, and direction
		var startingRow, direction int
		if color == 'w' {
			startingRow = 6
			direction = -1
		} else {
			startingRow = 1
			direction = 1
		}

		numOfMoves := 1
		if coord.row == startingRow {
			numOfMoves = 2
		}

		for i := 0; i < numOfMoves; i++ {
			newCoords := &Coords{coord.row + direction*(i+1), coord.col}

			if c.checkIfCoordsIsNotValid(newCoords) {
				continue
			}

			if c.checkIfThereIsPieceInCoords(newCoords) {
				break
			}

			validMoves = append(validMoves, newCoords)
		}

		for _, sideDirection := range []int{1, -1} {
			newCoords := &Coords{coord.row + direction, coord.col + sideDirection}

			if c.checkIfCoordsIsNotValid(newCoords) {
				continue
			}

			if !c.checkIfThereIsPieceInCoords(newCoords) {
				if c.pawnPassant == "-" {
					continue
				}

				pawnPassantCoord := c.translateCBtoCoords(c.pawnPassant)
				if c.checkIfAllyInCoords(pawnPassantCoord, color) {
					continue
				}

				if !(coord.row == pawnPassantCoord.row && newCoords.col == pawnPassantCoord.col) {
					continue
				}
			}

			if c.checkIfAllyInCoords(newCoords, color) {
				continue
			}

			validMoves = append(validMoves, newCoords)
		}

	// Knight
	case 'n':

	case 'b':
	case 'r':
	case 'q':
	case 'k':

	}

	return validMoves
}

// checkIfCoordsIsNotValid
func (c *Chess) checkIfCoordsIsNotValid(coord *Coords) bool {
	if coord.row > 7 || coord.row < 0 || coord.col > 7 || coord.col < 0 {
		return true
	}
	return false
}

// checkIfThereIsPieceInCoords
func (c *Chess) checkIfThereIsPieceInCoords(coord *Coords) bool {
	return c.determinePieceWithCoords(coord) != '-'
}

// checkIfAllyInCoords
func (c *Chess) checkIfAllyInCoords(coord *Coords, color rune) bool {
	colorInCoord := c.determineColor(c.determinePieceWithCoords(coord))

	return colorInCoord == color
}
