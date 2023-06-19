package engine

type FENError struct {
	err string
}

func (F *FENError) Error() string {
	return "Invalid FEN: " + F.err
}
