package main

import (
	"context"
	"log"
	"strconv"

	"google.golang.org/grpc"

	"bufio"
	"fmt"
	"os"
	"time"

	connect4 "github.com/danieljcksn/connect-four/proto"
)

const COLS = 7

func main() {
	scanner := bufio.NewScanner(os.Stdin)

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

	fmt.Print("Enter your nickname: ")
	scanner.Scan()

	nickname := scanner.Text()

	if err := stream.Send(&connect4.GameCommand{Command: "connect", Nickname: nickname}); err != nil {
		log.Fatalf("failed to send connection request: %v", err)
	}

	isMyTurn := false

	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Fatalf("Failed to receive a message: %v", err)
			}
			fmt.Println(in.Message)
			if in.Board != "" { // Check if the board data is included in the update
				fmt.Println("Current Board:")
				fmt.Println(in.Board) // Print the formatted board received from the server
			}

			// Check if it's this client's turn
			if in.Message == "It's your turn!" || in.Message == "Invalid move. Try again." {
				isMyTurn = true
			} else {
				isMyTurn = false
			}

			fmt.Println()
		}
	}()

	for {
		for !isMyTurn {
			// Wait for the turn flag to change
			<-time.After(time.Millisecond) // Add a small delay to avoid spinning and using 100% CPU
		}

		fmt.Println("Enter column number (1 to 7):")

		scanner.Scan()
		column, err := strconv.Atoi(scanner.Text())

		if err != nil || column < 1 || column > COLS {
			fmt.Println("Invalid column. Please enter a number between 1 and 7.")
			continue
		}

		if err := stream.Send(&connect4.GameCommand{Command: "move", Column: int32(column - 1)}); err != nil {
			log.Fatalf("failed to send move: %v", err)
		}

		isMyTurn = false // Reset the turn flag after making a move
	}

}
