package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	connect4 "github.com/danieljcksn/connect-four/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type ClientInfo struct {
	IP       string
	Nickname string
	Symbol   string

	Stream connect4.Connect4Game_GameSessionServer // Store the stream reference, so we can broadcast messages to both clients later
}

type server struct {
	connect4.UnimplementedConnect4GameServer
	clients     map[string]ClientInfo // Track clients by IP and include their nickname
	clientsLock sync.Mutex            // Ensure thread-safe access to clients
	gameBoard   [ROWS][COLS]int32     // 6 rows, 7 columns

	currentPlayer int // Index of current player, 0 or 1
	players       [2]string
}

const (
	ROWS = 6
	COLS = 7
)

func (s *server) GameSession(stream connect4.Connect4Game_GameSessionServer) error {
	// Get the network information of the client
	p, ok := peer.FromContext(stream.Context())

	if !ok {
		return fmt.Errorf("error retrieving peer information")
	}

	ipAddr := p.Addr.String()
	s.clientsLock.Lock()

	if len(s.clients) >= 2 {
		s.clientsLock.Unlock()
		return fmt.Errorf("maximum number of clients reached")
	}

	client, exists := s.clients[ipAddr]

	if !exists {
		client = ClientInfo{
			IP:     ipAddr,
			Stream: stream,
			Symbol: chooseSymbol(len(s.clients)),
		}
		s.players[len(s.clients)] = ipAddr
		s.clients[ipAddr] = client
	}

	s.clientsLock.Unlock()

	return s.handleClientCommands(stream, ipAddr)
}

// handleClientCommands processes commands from the client's stream.
func (s *server) handleClientCommands(stream connect4.Connect4Game_GameSessionServer, ipAddr string) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive a message: %v", err)
			break
		}

		switch in.Command {
		case "connect":
			s.handleConnectCommand(ipAddr, in.Nickname, stream)
		case "move":
			s.handleMoveCommand(ipAddr, in.Column, stream)
		}
	}

	s.clientsLock.Lock()
	delete(s.clients, ipAddr)
	s.clientsLock.Unlock()

	return nil
}

func chooseSymbol(numClients int) string {
	if numClients == 0 {
		return "x"
	}
	return "o"
}

// handleConnectCommand processes the connect command from the client.
func (s *server) handleConnectCommand(ipAddr, nickname string, stream connect4.Connect4Game_GameSessionServer) {
	s.clientsLock.Lock()
	client := s.clients[ipAddr]
	client.Nickname = nickname
	s.clients[ipAddr] = client
	s.clientsLock.Unlock()

	fmt.Println(nickname, "connected from IP:", ipAddr, ". The player's symbol is:", client.Symbol)
	stream.Send(&connect4.GameUpdate{Message: "Welcome to Connect Four, " + nickname + "!"})

	if len(s.clients) == 2 {
		s.broadcast((nickname + " has joined the game!"), "You are now connected.", "", false)
		s.broadcast("It's your turn!", "Waiting for "+s.clients[s.players[s.currentPlayer]].Nickname+" to make a move.", s.formatBoard(), false)
	} else {
		stream.Send(&connect4.GameUpdate{Message: "Just a second! Waiting for another player to connect"})
	}
}

// handleMoveCommand processes the move command from the client.
func (s *server) handleMoveCommand(ipAddr string, column int32, stream connect4.Connect4Game_GameSessionServer) {
	if len(s.clients) < 2 {
		stream.Send(&connect4.GameUpdate{Message: "Just a second! Waiting for another player to connect."})
		return
	}

	if s.players[s.currentPlayer] != ipAddr {
		stream.Send(&connect4.GameUpdate{Message: "It's not your turn yet."})
		return
	}

	if !s.isValidMove(column) {
		stream.Send(&connect4.GameUpdate{Message: "Invalid move. Try again."})
		return
	}

	s.applyMove(column, s.clients[ipAddr].Symbol)
	winner := s.checkForWinner()
	switch winner {
	case "Tie":
		s.broadcast("The game is a tie.", "The game is a tie.", s.formatBoard(), true)
		return // End game session after a tie
	case "":
		s.switchPlayerTurn()
		s.broadcast("It's your turn!", "Move accepted. "+s.clients[s.players[s.currentPlayer]].Nickname+"'s turn", s.formatBoard(), false)
	default:
		s.broadcast("Congratulations, "+s.clients[s.players[s.currentPlayer]].Nickname+"! You won!", "You lost. Better luck next time.", s.formatBoard(), true)
		return // End game session after a win
	}
}

func (s *server) broadcast(activeMessage, otherMessage, board string, closeConnections bool) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	for _, client := range s.clients {
		if client.Stream != nil {
			message := otherMessage
			if client.IP == s.players[s.currentPlayer] {
				message = activeMessage
			}

			client.Stream.Send(&connect4.GameUpdate{Message: message, Board: board})

			if closeConnections {
				client.Stream.Context().Done() // Close the stream
			}
		}
	}
}

// Game logic functions and helpers
func (s *server) isValidMove(column int32) bool {
	return column >= 0 && column < COLS && s.gameBoard[0][column] == 0
}

func (s *server) applyMove(column int32, symbol string) {
	for i := ROWS - 1; i >= 0; i-- {
		if s.gameBoard[i][column] == 0 {
			playerMark := int32(1)
			if symbol == "o" {
				playerMark = 2
			}
			s.gameBoard[i][column] = playerMark
			break
		}
	}
}

func (s *server) checkForWinner() string {
	// Check rows for a 4-in-a-row
	for i := 0; i < ROWS; i++ {
		for j := 0; j < COLS-3; j++ {
			if s.gameBoard[i][j] != 0 && s.gameBoard[i][j] == s.gameBoard[i][j+1] && s.gameBoard[i][j] == s.gameBoard[i][j+2] && s.gameBoard[i][j] == s.gameBoard[i][j+3] {
				return s.clients[s.players[s.currentPlayer]].Nickname
			}
		}
	}

	// Check columns for a 4-in-a-row
	for j := 0; j < COLS; j++ {
		for i := 0; i < ROWS-3; i++ {
			if s.gameBoard[i][j] != 0 && s.gameBoard[i][j] == s.gameBoard[i+1][j] && s.gameBoard[i][j] == s.gameBoard[i+2][j] && s.gameBoard[i][j] == s.gameBoard[i+3][j] {
				return s.clients[s.players[s.currentPlayer]].Nickname
			}
		}
	}

	// Check positive diagonal for a 4-in-a-row
	for i := 0; i < ROWS-3; i++ {
		for j := 0; j < COLS-3; j++ {
			if s.gameBoard[i][j] != 0 && s.gameBoard[i][j] == s.gameBoard[i+1][j+1] && s.gameBoard[i][j] == s.gameBoard[i+2][j+2] && s.gameBoard[i][j] == s.gameBoard[i+3][j+3] {
				return s.clients[s.players[s.currentPlayer]].Nickname
			}
		}
	}

	// Check negative diagonal for a 4-in-a-row
	for i := 0; i < ROWS-3; i++ {
		for j := 3; j < COLS; j++ {
			if s.gameBoard[i][j] != 0 && s.gameBoard[i][j] == s.gameBoard[i+1][j-1] && s.gameBoard[i][j] == s.gameBoard[i+2][j-2] && s.gameBoard[i][j] == s.gameBoard[i+3][j-3] {
				return s.clients[s.players[s.currentPlayer]].Nickname
			}
		}
	}

	// Check for a tie (no empty spaces left)
	for i := 0; i < ROWS; i++ {
		for j := 0; j < COLS; j++ {
			if s.gameBoard[i][j] == 0 {
				return ""
			}
		}
	}

	return "Tie" // If no empty cells and no winner, it's a tie
}

func (s *server) switchPlayerTurn() {
	s.currentPlayer ^= 1 // Toggle between 0 and 1 using XOR
}

// Formatting functions
func (s *server) formatBoard() string {
	var boardStr string
	for i := 0; i < ROWS; i++ {
		for j := 0; j < COLS; j++ {
			cellSymbol := GetCellSymbol(s.gameBoard[i][j])
			boardStr += fmt.Sprintf("[%s]", cellSymbol)
		}
		boardStr += "\n"
	}
	return boardStr
}

func GetCellSymbol(value int32) string {
	switch value {
	case 0:
		return " "
	case 1:
		return "x"
	case 2:
		return "o"
	default:
		return "?"
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	connect4.RegisterConnect4GameServer(s, &server{clients: make(map[string]ClientInfo)})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	print("Server started on port 50051")
}
