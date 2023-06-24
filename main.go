package main

import (
	"chess-go/engine"
	"fmt"
	"os/exec"
)

func main() {
	chess := engine.NewGameChess()

	//var from, to string
	var pgn string

	for {
		chess.PrintBoard()

		fmt.Printf("\nMake a move (%s): ", string(chess.Turn()))
		_, err := fmt.Scan(&pgn)
		if err != nil {
			panic(err)
		}

		//_, err1 := chess.Move(from, to)
		_, err1 := chess.MovePGN(pgn)
		if err1 != nil {
			println(err1.Error(), "\ntry again")
			var scan string
			_, err1 = fmt.Scan(&scan)
			if err1 != nil {
				panic(err1)
			}
		}

		exec.Command("clear")
	}
}
