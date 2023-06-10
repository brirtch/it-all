// I referenced this article in creating this: https://eli.thegreenplace.net/2019/on-concurrency-in-go-http-servers/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GameServer struct {
	Games []*Game

	createGameCommandChannel chan CreateGameRequest
	joinGameCommandChannel   chan JoinGameRequest
	getGamesRequestChannel   chan GetGamesRequest
	getGameStatusChannel     chan GetGameStatusRequest
	buyRequestChannel        chan BuyRequest
}

type CreateGameRequest struct {
	PlayerName string `json:"playerName"`
	ReplyChan  chan GameListItemResponse
}

type GetGamesRequest struct {
	ReplyChan chan GetGamesResponse
}

type JoinGameRequest struct {
	PlayerName string `json:"playerName"`
	GameID     string `json:"gameID"`
	ReplyChan  chan JoinGameResponse
}

type JoinGameResponse struct {
	PlayerID string `json:"playerID"`
}

type BuyRequest struct {
	ItemName  string `json:"itemName"`
	Quantity  int    `json:"quantity"`
	ReplyChan chan BuyResponse
}

type BuyResponse struct {
	Success bool `json:"success"`
}

type GameListItemResponse struct {
	GameID         string `json:"gameID"`
	GameName       string `json:"gameName"`
	HostPlayerName string `json:"hostPlayerName"`
}

type GetGameStatusRequest struct {
	GameID    string
	PlayerID  string
	ReplyChan chan GameStatusResponse
}

type GameStatusResponse struct {
	Player       *Player       `json:"player"`
	OtherPlayers []string      `json:"otherPlayers"`
	People       []*Person     `json:"people"`
	GameObjects  []*GameObject `json:"gameObjects"`
}

type GetGamesResponse struct {
	Games []GameListItemResponse `json:"games"`
}

// Create a new GameServer.
func NewGameServer() *GameServer {
	server := &GameServer{getGamesRequestChannel: make(chan GetGamesRequest)}
	server.Games = []*Game{}
	//server.
	server.createGameCommandChannel = make(chan CreateGameRequest)
	server.joinGameCommandChannel = make(chan JoinGameRequest)
	server.getGamesRequestChannel = make(chan GetGamesRequest)
	server.getGameStatusChannel = make(chan GetGameStatusRequest)
	server.buyRequestChannel = make(chan BuyRequest)

	return server
}

// Game server loop.
func (server *GameServer) Run() {
	fmt.Println("In GameServer.Run()")

	for {

		start := time.Now()
		//fmt.Printf("Game Server heartbeat %d", len(server.Games))

		select {
		case createGameRequest := <-server.createGameCommandChannel:
			g := NewGame(createGameRequest.PlayerName)
			server.Games = append(server.Games, g)
			createGameResponse := GameListItemResponse{GameID: g.GameID, GameName: g.GameName, HostPlayerName: g.HostPlayerName}
			createGameRequest.ReplyChan <- createGameResponse

		case joinGameRequest := <-server.joinGameCommandChannel:
			fmt.Println("Received join game request.")
			// Find the game.
			g := server.GetGameByID(joinGameRequest.GameID)
			var p *Player
			if g.GetPlayerByName(joinGameRequest.PlayerName) != nil {
				p = g.GetPlayerByName(joinGameRequest.PlayerName)
			} else {
				p = g.NewPlayer(uuid.New().String(), joinGameRequest.PlayerName)
				g.Players = append(g.Players, p)
			}
			joinGameResponse := JoinGameResponse{PlayerID: p.PlayerID}
			joinGameRequest.ReplyChan <- joinGameResponse
		case getGameRequest := <-server.getGamesRequestChannel:
			getGamesResponse := GetGamesResponse{}
			for _, game := range server.Games {
				game := GameListItemResponse{GameID: game.GameID, GameName: game.GameName, HostPlayerName: game.HostPlayerName}
				getGamesResponse.Games = append(getGamesResponse.Games, game)
			}
			getGameRequest.ReplyChan <- getGamesResponse
		case getGameStatusRequest := <-server.getGameStatusChannel:
			//Get player details
			game := server.GetGameByID(getGameStatusRequest.GameID)
			var gameStatusResponse GameStatusResponse
			if game != nil {
				thisPlayer := game.GetPlayerByID(getGameStatusRequest.PlayerID)

				var playerNames []string
				g := server.GetGameByID(getGameStatusRequest.GameID)
				for _, player := range g.Players {
					playerNames = append(playerNames, player.Name)
				}

				gameStatusResponse = GameStatusResponse{Player: thisPlayer, OtherPlayers: playerNames, People: thisPlayer.People, GameObjects: g.Map.GetNonBlankTiles()}
			} else {
				gameStatusResponse = GameStatusResponse{Player: nil, OtherPlayers: nil, People: nil, GameObjects: nil}

			}
			getGameStatusRequest.ReplyChan <- gameStatusResponse
		case buyRequest := <-server.buyRequestChannel:
			// Purchase item.
			//player.Buy(buyRequest.ItemName, buyRequest.Quantity)
			buyResponse := BuyResponse{Success: true}
			buyRequest.ReplyChan <- buyResponse
		default: // do nothing
		}

		// Update all games.
		var wg sync.WaitGroup
		wg.Add(len(server.Games))

		for _, game := range server.Games {
			go game.Update(&wg)
		}
		wg.Wait()
		executionDuration := time.Since(start)

		msToWait := time.Duration((17 - executionDuration.Milliseconds()) * int64(time.Millisecond))

		time.Sleep(msToWait)
	}
}

// Returns a game given its ID.
func (server *GameServer) GetGameByID(gameID string) *Game {
	for _, game := range server.Games {
		if game.GameID == gameID {
			return game
		}
	}
	return nil
}

// POST /game/create
func (server *GameServer) CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	var createGameRequest CreateGameRequest
	err := json.NewDecoder(r.Body).Decode(&createGameRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	createGameRequest.ReplyChan = make(chan GameListItemResponse)
	server.createGameCommandChannel <- createGameRequest
	createGameResponse := <-createGameRequest.ReplyChan

	responseJSON, _ := json.Marshal(createGameResponse)
	w.Write(responseJSON)
}

// POST /game/join
func (server *GameServer) JoinGameHandler(w http.ResponseWriter, r *http.Request) {
	var joinGameRequest JoinGameRequest
	err := json.NewDecoder(r.Body).Decode(&joinGameRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	joinGameRequest.ReplyChan = make(chan JoinGameResponse)
	server.joinGameCommandChannel <- joinGameRequest
	joinGameResponse := <-joinGameRequest.ReplyChan

	responseJSON, _ := json.Marshal(joinGameResponse)
	w.Write(responseJSON)
}

// POST /game/games
func (server *GameServer) GetGamesHandler(w http.ResponseWriter, r *http.Request) {
	getGamesRequest := GetGamesRequest{ReplyChan: make(chan GetGamesResponse)}
	server.getGamesRequestChannel <- getGamesRequest
	getGamesResponse := <-getGamesRequest.ReplyChan
	responseJSON, _ := json.Marshal(getGamesResponse)
	w.Write(responseJSON)
}

// GET /game/{gameID}/{playerID}/status
func (server *GameServer) GetGameStatusHandler(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	playerID := chi.URLParam(r, "playerID")

	getGameStatusRequest := GetGameStatusRequest{GameID: gameID, PlayerID: playerID, ReplyChan: make(chan GameStatusResponse)}
	server.getGameStatusChannel <- getGameStatusRequest
	gameStatusResponse := <-getGameStatusRequest.ReplyChan

	responseJSON, err := json.Marshal(gameStatusResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}

func (server *GameServer) BuyHandler(w http.ResponseWriter, r *http.Request) {
	var buyRequest BuyRequest
	err := json.NewDecoder(r.Body).Decode(&buyRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
}
