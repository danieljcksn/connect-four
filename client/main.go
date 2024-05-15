package main

import (
	"context"
	"log"
	"strconv"

	"google.golang.org/grpc"

	"bufio"
	"fmt"
	"os"

	connect4 "github.com/danieljcksn/connect-four/proto"
)

const COLS = 7

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	client := connect4.NewConnect4GameClient(conn)

	stream, err := client.GameSession(context.Background())
	if err != nil {
		log.Fatalf("error creating game session: %v", err)
	}

	// Send connect command
	if err := stream.Send(&connect4.GameCommand{Command: "connect", Nickname: "Daniel"}); err != nil {
		log.Fatalf("failed to send connection request: %v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Fatalf("Failed to receive a message: %v", err)
			}
			if in.Board != "" { // Checks if the board data is included in the update
				fmt.Println("Current Board:")
				fmt.Println(in.Board) // Prints the formatted board received from the server
			}
			fmt.Println(in.Message) // This prints any text message received from the server
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Choose an option: \n1) Show board\n2) Make a move\n3) Check turn")
		scanner.Scan()
		input := scanner.Text()

		switch input {
		case "1":
			clearScreen()

			if err := stream.Send(&connect4.GameCommand{Command: "show_board"}); err != nil {
				log.Fatalf("failed to send request for board: %v", err)
			}
		case "2":
			clearScreen()

			fmt.Println("Enter column number:")
			scanner.Scan()
			column, err := strconv.Atoi(scanner.Text())

			if err != nil || column < 1 || column > COLS {
				fmt.Println("Invalid column. Please enter a number between 1 and 7.")
				continue
			}
			if err := stream.Send(&connect4.GameCommand{Command: "move", Column: int32(column - 1)}); err != nil {
				log.Fatalf("failed to send move: %v", err)
			}
		case "3":
			clearScreen()

			if err := stream.Send(&connect4.GameCommand{Command: "check_turn"}); err != nil {
				log.Fatalf("failed to send check turn request: %v", err)
			}
		}
	}

}
