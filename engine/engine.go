package engine

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	DefaultFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	MaxStep    = 7
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

type Chess struct {
	boardTable  Board
	turn        rune
	castle      CastleAvailability
	pawnPassant string
	halfmoves   int
	fullmoves   int
}

// Move moves a piece
func (c *Chess) Move(from, to string) {

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
	fen += " " + string(c.turn) + " "

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
	if len(splitFen[1]) != 1 {
		return &FENError{err: "invalid turn parameter"}
	}
	c.turn = []rune(splitFen[1])[0]

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
	if translateCBtoCoords(c.pawnPassant) == nil && c.pawnPassant != "-" {
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

// calculateValidMoves calculates the valid paths in a given piece coordinate
func (c *Chess) calculateValidMoves(piece rune, coord *Coords) []*Coords {
	var validMoves []*Coords

	color := determineColor(piece)

	moves := c.calculateMoves(piece, coord, c.boardTable, MaxStep)

	for _, move := range moves {
		newBoard := c.boardTable
		movePiece(coord, move, &newBoard)

		if c.checkIfChecked(color, newBoard) {
			continue
		}

		validMoves = append(validMoves, move)
	}

	return validMoves
}

// TODO: optimize move calculations
// calculateMoves calculates the paths in a given piece and coordinate
func (c *Chess) calculateMoves(piece rune, coord *Coords, board Board, maxStep int) []*Coords {
	var moves []*Coords
	color := determineColor(piece)

	filter := func(move *Coords) {
		if !checkIfCoordsIsOutOfBounds(move) {
			moves = append(moves, move)
		}
	}

	switch unicode.ToLower(piece) {
	// Pawn
	case 'p':
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

			if checkIfThereIsPieceInCoords(newCoords, board) {
				break
			}

			filter(newCoords)
		}

		for _, sideDirection := range []int{1, -1} {
			newCoords := &Coords{coord.row + direction, coord.col + sideDirection}

			if !checkIfThereIsPieceInCoords(newCoords, board) {
				if c.pawnPassant == "-" {
					continue
				}

				pawnPassantCoord := translateCBtoCoords(c.pawnPassant)
				if checkIfAllyInCoords(pawnPassantCoord, color, board) {
					continue
				}

				if !(coord.row == pawnPassantCoord.row && newCoords.col == pawnPassantCoord.col) {
					continue
				}
			}

			if checkIfAllyInCoords(newCoords, color, board) {
				continue
			}

			filter(newCoords)
		}
	// Knight
	case 'n':
		for _, rowDirection := range []int{2, -2} {
			for _, colDirection := range []int{1, -1} {
				newCoords := &Coords{coord.row + rowDirection, coord.col + colDirection}

				if checkIfAllyInCoords(newCoords, color, board) {
					continue
				}

				filter(newCoords)
			}
		}
		for _, rowDirection := range []int{1, -1} {
			for _, colDirection := range []int{2, -2} {
				newCoords := &Coords{coord.row + rowDirection, coord.col + colDirection}

				if checkIfAllyInCoords(newCoords, color, board) {
					continue
				}

				filter(newCoords)
			}
		}
	// Bishop
	case 'b':
		for _, rowDirection := range []int{1, -1} {
			for _, colDirection := range []int{1, -1} {
				for i := 0; i < maxStep; i++ {
					newCoords := &Coords{coord.row + rowDirection*(i+1), coord.col + colDirection*(i+1)}

					if checkIfThereIsPieceInCoords(newCoords, board) {
						if !checkIfAllyInCoords(newCoords, color, board) {
							filter(newCoords)
						}
						break
					}

					filter(newCoords)
				}
			}
		}
	// Rook
	case 'r':
		for _, rowDirection := range []int{1, -1} {
			for i := 0; i < maxStep; i++ {
				newCoords := &Coords{coord.row + rowDirection*(i+1), coord.col}

				if checkIfThereIsPieceInCoords(newCoords, board) {
					if !checkIfAllyInCoords(newCoords, color, board) {
						filter(newCoords)
					}
					break
				}

				filter(newCoords)
			}
		}
		for _, colDirection := range []int{1, -1} {
			for i := 0; i < maxStep; i++ {
				newCoords := &Coords{coord.row, coord.col + colDirection*(i+1)}

				if checkIfThereIsPieceInCoords(newCoords, board) {
					if !checkIfAllyInCoords(newCoords, color, board) {
						filter(newCoords)
					}
					break
				}

				filter(newCoords)
			}
		}
	// Queen
	case 'q':
		moves = append(moves, c.calculateMoves(determineColorPiece(color, 'b'), coord, board, maxStep)...)
		moves = append(moves, c.calculateMoves(determineColorPiece(color, 'r'), coord, board, maxStep)...)
	// King
	case 'k':
		moves = append(moves, c.calculateMoves(determineColorPiece(color, 'q'), coord, board, 1)...)
	}

	return moves
}

// checkIfMate checks if the king is mated
func (c *Chess) checkIfMate(color rune) bool {
	for y, row := range c.boardTable {
		for x, piece := range row {
			if determineColor(piece) == color {
				if len(c.calculateValidMoves(piece, &Coords{y, x})) > 0 {
					return false
				}
			}
		}
	}

	return true
}

// checkIfChecked checks if the king is checked
func (c *Chess) checkIfChecked(color rune, board Board) bool {
	king := determineColorPiece(color, 'k')

	var kingCoord Coords

	for y, row := range board {
		for x, piece := range row {
			if piece == king {
				kingCoord.col = x
				kingCoord.row = y
			}
		}
	}

	for _, attackingPiece := range []rune{'p', 'n', 'b', 'r', 'q'} {
		attackingPiece = determineColorPiece(color, attackingPiece)
		enemyPiece := determineEnemyVersion(attackingPiece)
		moves := c.calculateMoves(attackingPiece, &kingCoord, board, MaxStep)

		for _, move := range moves {
			if determinePieceWithCoords(move, board) == enemyPiece {
				return true
			}
		}
	}

	return false
}

func movePiece(from *Coords, to *Coords, board *Board) {
	board[to.row][to.col] = board[from.row][from.col]
	board[from.row][from.col] = '-'
}
