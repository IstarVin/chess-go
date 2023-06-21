package main

import (
	"chess-go/engine"
	"fmt"
	"os/exec"
)

func main() {
	chess := engine.NewGameChess()

	var move string
	var from, to string

	for {
		chess.PrintBoard()

		fmt.Print("\nMake a move: ")
		_, err := fmt.Scan(&from, &to)
		println(move)
		if err != nil {
			panic(err)
		}

		if move == "exit" {
			break
		}

		_, err1 := chess.Move(from, to)
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
