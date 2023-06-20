package engine

type FENError struct {
	err string
}

func (F *FENError) Error() string {
	return "Invalid FEN: " + F.err
}

type MoveError struct {
	err string
}

func (m *MoveError) Error() string {
	return "Invalid Move: " + m.err
}
