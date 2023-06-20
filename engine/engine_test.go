package engine

import (
	"fmt"
	"reflect"
	"testing"
)

//func TestEngine(t *testing.T) {
//
//}

func TestEngine_GetFen(t *testing.T) {
	inputs := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w Kq c6 5 23",
	}

	for _, input := range inputs {
		chess, _ := NewChessGameWithFen(input)
		output := chess.GetFEN()

		if input != output {
			t.Errorf("FAILED\n\tgot:      %+v\n\texpected: %+v", output, input)
		}
	}
}
func TestEngine_determineColor(t *testing.T) {
	inputs := []rune{
		'b', 'r', 'R', 'K', '-',
	}

	expectedOutputs := []rune{
		'b', 'b', 'w', 'w', '-',
	}

	for i, input := range inputs {
		expected := expectedOutputs[i]
		output := determineColor(input)

		if output != expected {
			t.Errorf("FAILED\n\tgot: %+v\n\texpected:%+v", output, expected)
		}
	}
}
func TestEngine_determinePieceWithCoords(t *testing.T) {
	inputs := []*Coords{
		{0, 2},
		{0, 5},
		{7, 0},
		{7, 6},
		//{10, 3},
	}

	expectedOutputs := []rune{
		'b',
		'b',
		'R',
		'N',
		'-',
	}

	chess := NewGameChess()

	for i, input := range inputs {
		expected := expectedOutputs[i]
		output := determinePieceWithCoords(input, chess.boardTable)

		if output != expected {
			t.Errorf("FAILED\n\tgot: %+v\n\texpected:%+v", output, expected)
		}
	}
}
func TestEngine_translateCBtoCoords(t *testing.T) {
	inputs := []string{
		"a8",
		"h0",
		"-",
	}

	expectedOutputs := []*Coords{
		{0, 0},
		nil,
		nil,
	}

	for i, input := range inputs {
		output := translateCBtoCoords(input)
		expected := expectedOutputs[i]

		if !reflect.DeepEqual(output, expected) {
			t.Errorf("FAILED\n\tgot: %+v\n\texpected:%+v", output, expected)
		}
	}
}
func TestEngine_translateCoordsToCB(t *testing.T) {
	inputs := []*Coords{
		{1, 1},
		{0, 0},
		{4, 5},
	}

	expectedOutputs := []string{
		"b7",
		"a8",
		"f4",
	}

	chess := NewGameChess()

	for i, input := range inputs {
		output := chess.translateCoordsToCB(input)
		expected := expectedOutputs[i]

		if !reflect.DeepEqual(output, expected) {
			t.Errorf("FAILED\n\tgot: %+v\n\texpected:%+v", output, expected)
		}
	}
}
func TestEngine_decodeFen(t *testing.T) {
	inputs := []string{
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w Kq c6 5 23",
	}

	expectedOutputs := []*Chess{
		{
			boardTable: Board{
				{'r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'},
				{'p', 'p', 'p', 'p', 'p', 'p', 'p', 'p'},
				{'-', '-', '-', '-', '-', '-', '-', '-'},
				{'-', '-', '-', '-', '-', '-', '-', '-'},
				{'-', '-', '-', '-', '-', '-', '-', '-'},
				{'-', '-', '-', '-', '-', '-', '-', '-'},
				{'P', 'P', 'P', 'P', 'P', 'P', 'P', 'P'},
				{'R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'},
			},
			turn: "w",
			castle: CastleAvailability{
				WhiteKing:  true,
				WhiteQueen: true,
				BlackKing:  true,
				BlackQueen: true,
			},
			pawnPassant: "-",
			halfmoves:   0,
			fullmoves:   1,
		},
		{
			boardTable: Board{
				{'r', 'n', 'b', 'q', 'k', 'b', 'n', 'r'},
				{'p', 'p', '-', 'p', 'p', 'p', 'p', 'p'},
				{'-', '-', '-', '-', '-', '-', '-', '-'},
				{'-', '-', 'p', '-', '-', '-', '-', '-'},
				{'-', '-', '-', '-', 'P', '-', '-', '-'},
				{'-', '-', '-', '-', '-', '-', '-', '-'},
				{'P', 'P', 'P', 'P', '-', 'P', 'P', 'P'},
				{'R', 'N', 'B', 'Q', 'K', 'B', 'N', 'R'},
			},
			turn: "w",
			castle: CastleAvailability{
				WhiteKing:  true,
				WhiteQueen: false,
				BlackKing:  false,
				BlackQueen: true,
			},
			pawnPassant: "c6",
			halfmoves:   5,
			fullmoves:   23,
		},
	}

	for i, input := range inputs {
		chess, err := NewChessGameWithFen(input)
		expected := expectedOutputs[i]

		if err != nil {
			fmt.Println(err.Error())
		}

		if !reflect.DeepEqual(chess, expected) {
			t.Errorf("FAILED\n\tgot:     %+v\n\texpected:%+v", chess, expected)
		}
	}
}
func TestEngine_calculateValidPaths(t *testing.T) {
	inputs := []string{
		"a6",
		"a2",
		"b5",
	}

	expectedOutputs := [][]*Coords{
		{{3, 0}, {3, 1}},
		{{5, 0}, {4, 0}},
		{{2, 1}, {2, 2}, {2, 0}},
	}

	chess, _ := NewChessGameWithFen("rnbqkbnr/1p1ppppp/p7/1Pp5/4P3/8/PPPP1PPP/RNBQKBNR w Kq c5 5 23")
	for i, input := range inputs {

		expectedOutput := expectedOutputs[i]
		coords := translateCBtoCoords(input)
		piece := determinePieceWithCoords(coords, chess.boardTable)
		output := chess.calculateValidMoves(piece, coords)

		var expectedOutputReadable []string
		for _, expectedRaw := range expectedOutput {
			expectedOutputReadable = append(expectedOutputReadable, chess.translateCoordsToCB(expectedRaw))
		}

		var outputReadable []string
		for _, outputRaw := range output {
			outputReadable = append(outputReadable, chess.translateCoordsToCB(outputRaw))
		}

		if !reflect.DeepEqual(output, expectedOutput) {
			chess.PrintBoard()
			t.Errorf("FAILED: %s\n\tgot:     %+v\n\texpected:%+v", input, outputReadable, expectedOutputReadable)
		}
	}
}
