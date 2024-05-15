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

const (
	ROWS = 6
	COLS = 7
)

type ClientInfo struct {
	IP       string
	Nickname string
	Symbol   string
}

type server struct {
	connect4.UnimplementedConnect4GameServer
	clients     map[string]ClientInfo // Track clients by IP and include their nickname
	clientsLock sync.Mutex            // Ensure thread-safe access to clients
	gameBoard   [ROWS][COLS]int32     // 6 rows, 7 columns

	currentPlayer int // Index of current player, 0 or 1
	players       [2]string
}

func (s *server) GameSession(stream connect4.Connect4Game_GameSessionServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return fmt.Errorf("error retrieving peer information")
	}

	ipAddr := p.Addr.String()
	s.clientsLock.Lock()
	if len(s.clients) >= 2 {
		s.clientsLock.Unlock()
		return fmt.Errorf("cannot connect more than two clients")
	}

	client, exists := s.clients[ipAddr]
	if !exists {
		client = ClientInfo{IP: ipAddr}
		if len(s.clients) == 0 {
			s.currentPlayer = 0 // First player starts
			client.Symbol = "x"
		} else {
			client.Symbol = "o"
		}
		s.players[len(s.clients)] = ipAddr // Store IP in the array
		s.clients[ipAddr] = client
	}
	s.clientsLock.Unlock()

	for {
		in, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive a message: %v", err)
			break
		}

		switch in.Command {
		case "connect":
			nickname := in.Nickname

			s.clientsLock.Lock()
			client.Nickname = nickname
			s.clients[ipAddr] = client
			s.clientsLock.Unlock()

			fmt.Println("%s connected from IP: %s with symbol %s", nickname, ipAddr, client.Symbol)
			stream.Send(&connect4.GameUpdate{Message: "Welcome to Connect Four, " + nickname + "!", Board: s.formatBoard()})

			if len(s.clients) == 2 {
				stream.Send(&connect4.GameUpdate{Message: "Game is ready to start", Board: s.formatBoard()})
			} else {
				stream.Send(&connect4.GameUpdate{Message: "Waiting for another player to connect"})
			}

		case "move":
			if len(s.clients) < 2 {
				stream.Send(&connect4.GameUpdate{Message: "Waiting for another player to connect"})
				continue
			}

			if s.players[s.currentPlayer] != ipAddr {
				stream.Send(&connect4.GameUpdate{Message: "Not your turn"})
				continue
			}

			if s.isValidMove(in.Column) {
				s.applyMove(in.Column, client.Symbol)

				winner := s.checkForWinner()

				if winner != "" {
					if winner == "Tie" {
						stream.Send(&connect4.GameUpdate{Message: "It's a tie!"})
					} else {
						stream.Send(&connect4.GameUpdate{Message: fmt.Sprintf("%s wins!", winner)})
					}

					return nil // End game session after a win
				}

				s.switchPlayerTurn()

				stream.Send(&connect4.GameUpdate{Message: "Move accepted", Board: s.formatBoard()})
			} else {
				stream.Send(&connect4.GameUpdate{Message: "Invalid move"})
			}

		case "show_board":
			boardString := s.formatBoard()
			stream.Send(&connect4.GameUpdate{Message: "Current Board", Board: boardString})

		case "check_turn":
			if len(s.clients) < 2 {
				stream.Send(&connect4.GameUpdate{Message: "Waiting for another player to connect"})
				continue
			}

			if s.players[s.currentPlayer] == ipAddr {
				stream.Send(&connect4.GameUpdate{Message: "It's your turn"})
			} else {
				stream.Send(&connect4.GameUpdate{Message: "It's the other player's turn"})
			}

		}

	}

	s.clientsLock.Lock()
	delete(s.clients, ipAddr)
	s.clientsLock.Unlock()
	return nil
}

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
}
