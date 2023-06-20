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

		err = chess.Move(from, to)
		if err != nil {
			println(err.Error(), "\ntry again")
		}

		exec.Command("clear")
	}
}
