package main

import "fmt"

const ROWS = 6
const COLS = 7

type Board [ROWS][COLS]string

type Game struct {
	board Board

	insertedByColumn [COLS]int

	round  int
	state  string
	winner string
}

func newBoard() (board Board) {
	for i := 0; i < ROWS; i++ {
		for j := 0; j < COLS; j++ {
			board[i][j] = " "
		}
	}

	return
}

func newGame() (game Game) {
	game.board = newBoard()
	game.round = 1
	game.state = "running"
	game.winner = ""

	return
}

func printBoard(
	board Board,
) {
	for i := 0; i < ROWS; i++ {
		if i == 0 {
			println(" 1  2  3  4  5  6  7")
		}
		for j := 0; j < COLS; j++ {
			print("[" + board[i][j] + "]")
		}
		println()
	}

	println()
}

func isColumnFull(
	board Board,
	col int,
) bool {
	return board[0][col] != " "
}

func game() {
	game := newGame()

	for game.state == "running" {
		var col int

		for {
			println("\n\nRound #", game.round)
			printBoard(game.board)

			fmt.Print("Enter column: ")
			fmt.Scan(&col)

			if col < 1 || col > 7 {
				println("Invalid column. Please enter a number between 1 and 7.")
				continue
			}

			if isColumnFull(game.board, col-1) {
				println("Column is full. Please enter another column.")
				continue
			}

			break
		}

		symbol := "o"

		if game.round&1 == 0 {
			symbol = "x"
		}

		game.board[ROWS-game.insertedByColumn[col-1]-1][col-1] = symbol
		game.insertedByColumn[col-1]++
		game.round++
	}

}
