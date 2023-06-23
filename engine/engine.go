package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	DefaultFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	maxStep    = 7
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

// Move moves a piece and returns statusCode (int) and error
//
// -1 = error
// 0 = normal
// 1 = check
// 2 = checkmate
func (c *Chess) Move(from, to string) (int, error) {
	if from == to {
		return -1, &MoveError{err: "no move happened"}
	}

	fromCoords := translateCBtoCoords(from)
	toCoords := translateCBtoCoords(to)

	piece := determinePieceWithCoords(fromCoords, c.boardTable)

	color := determineColor(piece)
	enemy := determineEnemy(color)

	// Check if current turn
	if color != c.turn {
		return -1, &MoveError{err: "Not current turn"}
	}

	// Check if valid move
	validMoves := c.calculateValidMoves(fromCoords)
	if !checkIfMovesContains(&validMoves, toCoords) {
		return -1, &MoveError{err: fmt.Sprintf("not a valid move from %s to %s", from, to)}
	}

	statusCode := 0

	c.halfmoves++

	// Check if capture then reset halfmoves
	if determinePieceWithCoords(toCoords, c.boardTable) != '-' {
		c.halfmoves = 0
	}

	movePiece(fromCoords, toCoords, &c.boardTable)

	// Check if checked
	if c.checkIfChecked(enemy, c.boardTable) {
		statusCode = 1

		if c.checkIfMate(enemy) {
			statusCode = 2
			c.winner = c.turn
		}
	}

	// Post process
	switch unicode.ToLower(piece) {
	case 'p':
		// Check if pawn moves 2 times
		if math.Abs(float64(fromCoords.row-toCoords.row)) == 2 {
			c.pawnPassant = from
		}
		c.halfmoves = 0

	case 'r':
		if color == 'w' {
			if fromCoords.col == 7 {
				c.castle.WhiteKing = false
			} else if fromCoords.col == 0 {
				c.castle.WhiteQueen = false
			}
		} else {
			if fromCoords.col == 7 {
				c.castle.BlackKing = false
			} else if fromCoords.col == 0 {
				c.castle.BlackQueen = false
			}
		}

	case 'k':
		// Determine the row of the king based on color
		kingRow := 0
		if color == 'w' {
			kingRow = 7
		}

		// Move rook if king castled
		if fromCoords.row-toCoords.row == 2 {
			movePiece(&Coords{row: kingRow, col: 0}, &Coords{row: kingRow, col: 3}, &c.boardTable)
		} else if fromCoords.row-toCoords.row == -2 {
			movePiece(&Coords{row: kingRow, col: 7}, &Coords{row: kingRow, col: 5}, &c.boardTable)
		}

		// Make castle availability false if king moved
		if color == 'w' {
			c.castle.WhiteKing = false
			c.castle.WhiteQueen = false
		} else {
			c.castle.BlackKing = false
			c.castle.BlackQueen = false
		}
	}

	// Increment fullmoves after the turn of black
	if c.turn == 'b' {
		c.fullmoves++
	}

	// Switch the turn
	c.turn = enemy

	return statusCode, nil
}

// PrintBoard TODO: Improve this
// PrintBoard prints the board
func (c *Chess) PrintBoard() {
	for i, row := range c.boardTable {
		for _, content := range row {
			print(string(content), " ")
		}
		println(8 - i)
	}

	println("a b c d e f g h")
}

// CalculateValidMoves calculates the valid paths in a given cb notation
func (c *Chess) CalculateValidMoves(cb string) []string {
	coords := translateCBtoCoords(cb)

	var validMoves []string

	for _, move := range c.calculateValidMoves(coords) {
		validMoves = append(validMoves, translateCoordsToCB(move))
	}

	return validMoves
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
func (c *Chess) calculateValidMoves(coord *Coords) []*Coords {
	var validMoves []*Coords

	piece := determinePieceWithCoords(coord, c.boardTable)

	moves := c.calculateMoves(piece, coord, c.boardTable, maxStep)

	for _, move := range moves {
		if c.checkIfMoveIsCheck(coord, move, c.boardTable) {
			continue
		}

		validMoves = append(validMoves, move)
	}

	return validMoves
}

// calculateMoves TODO: optimize move calculations
// calculateMoves calculates the paths in a given piece and coordinate
func (c *Chess) calculateMoves(piece rune, coord *Coords, board Board, maxStep int) []*Coords {
	var moves []*Coords
	color := determineColor(piece)

	addMove := func(move *Coords) {
		// Check if the move is out of bounds
		if checkIfCoordsIsOutOfBounds(move) {
			return
		}

		moves = append(moves, move)
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

			addMove(newCoords)
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

			addMove(newCoords)
		}
	// Knight
	case 'n':
		for _, rowDirection := range []int{2, -2} {
			for _, colDirection := range []int{1, -1} {
				newCoords := &Coords{coord.row + rowDirection, coord.col + colDirection}

				if checkIfAllyInCoords(newCoords, color, board) {
					continue
				}

				addMove(newCoords)
			}
		}
		for _, rowDirection := range []int{1, -1} {
			for _, colDirection := range []int{2, -2} {
				newCoords := &Coords{coord.row + rowDirection, coord.col + colDirection}

				if checkIfAllyInCoords(newCoords, color, board) {
					continue
				}

				addMove(newCoords)
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
							addMove(newCoords)
						}
						break
					}

					addMove(newCoords)
				}
			}
		}
	// Rook
	case 'r':
		for _, switcher := range []int{0, 1} {
			for _, direction := range []int{1, -1} {
				for i := 0; i < maxStep; i++ {
					var switcher1 int
					if switcher == 0 {
						switcher1 = 1
					} else {
						switcher1 = 0
					}

					newCoords := &Coords{
						coord.row + direction*(i+1)*switcher,
						coord.col + direction*(i+1)*switcher1,
					}

					if checkIfThereIsPieceInCoords(newCoords, board) {
						if !checkIfAllyInCoords(newCoords, color, board) {
							addMove(newCoords)
						}
						break
					}

					addMove(newCoords)
				}
			}
		}
	// Queen
	case 'q':
		moves = append(moves, c.calculateMoves(determineColorPiece(color, 'b'), coord, board, maxStep)...)
		moves = append(moves, c.calculateMoves(determineColorPiece(color, 'r'), coord, board, maxStep)...)
	// King
	case 'k':
		moves = append(moves, c.calculateMoves(determineColorPiece(color, 'q'), coord, board, 1)...)

		// Break if the king is checked
		if c.checkIfChecked(color, board) {
			break
		}

		// Define row based on color
		whiteRow := 7
		blackRow := 0

		// Define col
		kingSideCol := 5
		queenSideCol := 3

		// Define coords of castle
		whiteKing := &Coords{row: whiteRow, col: kingSideCol}
		whiteQueen := &Coords{row: whiteRow, col: queenSideCol}
		blackKing := &Coords{row: blackRow, col: kingSideCol}
		blackQueen := &Coords{row: blackRow, col: queenSideCol}

		// Define coords of side of castle
		whiteKing1 := &Coords{row: whiteRow, col: kingSideCol + 1}
		whiteQueen1 := &Coords{row: whiteRow, col: queenSideCol - 1}
		blackKing1 := &Coords{row: blackRow, col: kingSideCol + 1}
		blackQueen1 := &Coords{row: blackRow, col: queenSideCol - 1}

		if color == 'w' {
			if c.castle.WhiteKing && !c.checkIfMoveIsCheck(coord, whiteKing1, board) {
				moves = append(moves, whiteKing)
			}
			if c.castle.WhiteQueen && !c.checkIfMoveIsCheck(coord, whiteQueen1, board) {
				moves = append(moves, whiteQueen)
			}
		} else {
			if c.castle.BlackKing && !c.checkIfMoveIsCheck(coord, blackKing1, board) {
				moves = append(moves, blackKing)
			}
			if c.castle.BlackQueen && !c.checkIfMoveIsCheck(coord, blackQueen1, board) {
				moves = append(moves, blackQueen)
			}
		}
	}

	return moves
}

// checkIfMate checks if the king is mated
func (c *Chess) checkIfMate(color rune) bool {
	for y, row := range c.boardTable {
		for x, piece := range row {
			if determineColor(piece) == color {
				if len(c.calculateValidMoves(&Coords{y, x})) > 0 {
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
		moves := c.calculateMoves(attackingPiece, &kingCoord, board, maxStep)

		for _, move := range moves {
			if determinePieceWithCoords(move, board) == enemyPiece {
				return true
			}
		}
	}

	return false
}

// checkIfMoveIsCheck Check if the move leads to a check
func (c *Chess) checkIfMoveIsCheck(from *Coords, to *Coords, board Board) bool {
	color := determineColor(determinePieceWithCoords(from, board))
	movePiece(from, to, &board)

	if c.checkIfChecked(color, board) {
		return true
	}
	return false
}
