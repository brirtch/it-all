// I referenced this article in creating this: https://eli.thegreenplace.net/2019/on-concurrency-in-go-http-servers/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GameServer struct {
	Games []*Game

	createGameCommandChannel           chan CreateGameRequest
	joinGameCommandChannel             chan JoinGameRequest
	getGamesRequestChannel             chan GetGamesRequest
	getGameStatusChannel               chan GetGameStatusRequest
	buyRequestChannel                  chan BuyRequest
	attackRequestChannel               chan AttackRequest
	getGameObjectLibraryRequestChannel chan GameObjectLibraryRequest
	sendMessageChannel                 chan SendMessageRequest
	transferObjectRequestChannel       chan TransferObjectRequest

	gameObjectLibrary []GameObject
	GameObjectTypes   map[string]interface{}
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
	GameID    string `json:"gameID"`
	PlayerID  string `json:"playerID"`
	Location  string `json:"location"`
	ItemName  string `json:"itemName"`
	Quantity  int    `json:"quantity"`
	ReplyChan chan BuyResponse
}

type BuyResponse struct {
	Success bool `json:"success"`
}

type SendMessageRequest struct {
	GameID      string `json:"gameID"`
	SenderID    string `json:"senderID"`
	RecipientID string `json:"recipientID"`
	MessageBody string `json:"messageBody"`
	ReplyChan   chan SendMessageResponse
}

type SendMessageResponse struct {
	Success bool `json:"success"`
}

type AttackRequest struct {
	GameID           string `json:"gameID"`
	AttackerID       string `json:"attackerID"`
	PlayerToAttackID string `json:"playerToAttackID"`
	SoldiersToCommit int    `json:"soldiersToCommit"`
	ReplyChan        chan AttackResponse
}

type AttackResponse struct {
	Outcome string `json:"outcome"`
}

type GameListItemResponse struct {
	GameID         string `json:"gameID"`
	GameName       string `json:"gameName"`
	HostPlayerName string `json:"hostPlayerName"`
}

type GameObjectLibraryRequest struct {
	ReplyChan chan GameObjectLibraryResponse
}

type GameObjectLibraryResponse struct {
	GameObjectLibrary []GameObject `json:"gameObjectLibrary"`
}

type GetGameStatusRequest struct {
	GameID    string
	PlayerID  string
	ReplyChan chan GameStatusResponse
}

type GameStatusPlayer struct {
	PlayerID   string `json:"playerID"`
	PlayerName string `json:"playerName"`
}

type GameObjectTally struct {
	Location string `json:"location"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type GameStatusResponse struct {
	Player       *Player             `json:"player"`
	OtherPlayers []*GameStatusPlayer `json:"otherPlayers"`

	GameObjects     []GameObject       `json:"gameObjects"`
	GameObjectTally []*GameObjectTally `json:"gameObjectTallies"`
}

type GetGamesResponse struct {
	Games []GameListItemResponse `json:"games"`
}

type TransferObjectRequest struct {
	GameID         string `json:"gameID"`
	PlayerID       string `json:"playerID"`
	ObjectType     string `json:"type"`
	SourceLocation string `json:"sourceLocation"`
	TargetLocation string `json:"targetLocation"`
	ReplyChan      chan TransferObjectResponse
}

type TransferObjectResponse struct {
	Success bool `json:"success"`
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
	server.attackRequestChannel = make(chan AttackRequest)
	server.getGameObjectLibraryRequestChannel = make(chan GameObjectLibraryRequest)
	server.sendMessageChannel = make(chan SendMessageRequest)
	server.transferObjectRequestChannel = make(chan TransferObjectRequest)

	server.GameObjectTypes = make(map[string]interface{})
	server.GameObjectTypes["Woodcutter"] = NewWoodCutter
	server.GameObjectTypes["FoodCollector"] = NewFoodCollector
	server.GameObjectTypes["Soldier"] = NewSoldier
	server.GameObjectTypes["Miner"] = NewMiner
	server.GameObjectTypes["Barracks"] = NewBarracks
	server.GameObjectTypes["Secateurs"] = NewSecateurs
	server.GameObjectTypes["Blacksmith"] = NewBlacksmith
	server.GameObjectTypes["TownCentre"] = NewTownCentre
	server.GameObjectTypes["Builder"] = NewBuilder
	server.GameObjectTypes["Scout"] = NewScout

	for _, v := range server.GameObjectTypes {
		f := reflect.ValueOf(v)
		retVal := f.Call([]reflect.Value{})
		retInterface := retVal[0].Interface().(GameObject)
		server.gameObjectLibrary = append(server.gameObjectLibrary, GameObject(retInterface))
	}

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
			g := NewGame(server, createGameRequest.PlayerName)
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

				var otherPlayers []*GameStatusPlayer
				g := server.GetGameByID(getGameStatusRequest.GameID)
				for _, player := range g.Players {
					if player.PlayerID != thisPlayer.PlayerID {
						playerToAdd := &GameStatusPlayer{PlayerID: player.PlayerID, PlayerName: player.Name}
						otherPlayers = append(otherPlayers, playerToAdd)
					}
				}

				gameStatusResponse = GameStatusResponse{Player: thisPlayer, OtherPlayers: otherPlayers, GameObjects: thisPlayer.GameObjects, GameObjectTally: thisPlayer.GetGameObjectTally()}
			} else {
				gameStatusResponse = GameStatusResponse{Player: nil, OtherPlayers: nil, GameObjects: nil, GameObjectTally: nil}

			}
			getGameStatusRequest.ReplyChan <- gameStatusResponse
		case sendMessageRequest := <-server.sendMessageChannel:
			game := server.GetGameByID(sendMessageRequest.GameID)
			senderPlayer := game.GetPlayerByID(sendMessageRequest.SenderID)
			recipientPlayer := game.GetPlayerByID(sendMessageRequest.RecipientID)
			senderPlayer.SendMessage(recipientPlayer, sendMessageRequest.MessageBody)
			sendMessageResponse := SendMessageResponse{Success: true}
			sendMessageRequest.ReplyChan <- sendMessageResponse
		case buyRequest := <-server.buyRequestChannel:
			// Purchase item.
			game := server.GetGameByID(buyRequest.GameID)
			thisPlayer := game.GetPlayerByID(buyRequest.PlayerID)
			success := thisPlayer.Buy(buyRequest.ItemName, buyRequest.Quantity, buyRequest.Location)
			buyResponse := BuyResponse{Success: success}
			buyRequest.ReplyChan <- buyResponse
		case attackRequest := <-server.attackRequestChannel:
			// Attack.
			game := server.GetGameByID(attackRequest.GameID)
			attacker := game.GetPlayerByID(attackRequest.AttackerID)
			target := game.GetPlayerByID(attackRequest.PlayerToAttackID)
			outcome := attacker.Attack(target, attackRequest.SoldiersToCommit)
			attackResponse := AttackResponse{Outcome: outcome}
			attackRequest.ReplyChan <- attackResponse
		case gameObjectLibraryRequest := <-server.getGameObjectLibraryRequestChannel:
			response := GameObjectLibraryResponse{GameObjectLibrary: server.gameObjectLibrary}
			gameObjectLibraryRequest.ReplyChan <- response
		case transferObjectRequest := <-server.transferObjectRequestChannel:
			game := server.GetGameByID(transferObjectRequest.GameID)
			thisPlayer := game.GetPlayerByID(transferObjectRequest.PlayerID)
			thisPlayer.TransferObject(transferObjectRequest.ObjectType, transferObjectRequest.SourceLocation, transferObjectRequest.TargetLocation)
			transferResponse := TransferObjectResponse{Success: true}
			transferObjectRequest.ReplyChan <- transferResponse
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
	//fmt.Println(string(responseJSON))
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
	buyRequest.ReplyChan = make(chan BuyResponse)
	server.buyRequestChannel <- buyRequest
	buyResponse := <-buyRequest.ReplyChan
	responseJSON, err := json.Marshal(buyResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}

func (server *GameServer) AttackHandler(w http.ResponseWriter, r *http.Request) {
	var attackRequest AttackRequest
	err := json.NewDecoder(r.Body).Decode(&attackRequest)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	attackRequest.ReplyChan = make(chan AttackResponse)
	server.attackRequestChannel <- attackRequest
	buyResponse := <-attackRequest.ReplyChan
	responseJSON, err := json.Marshal(buyResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}

func (server *GameServer) GetGameObjectLibraryHandler(w http.ResponseWriter, r *http.Request) {
	var request GameObjectLibraryRequest
	request.ReplyChan = make(chan GameObjectLibraryResponse)
	server.getGameObjectLibraryRequestChannel <- request
	response := <-request.ReplyChan
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}

func (server *GameServer) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	var request SendMessageRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	request.ReplyChan = make(chan SendMessageResponse)
	server.sendMessageChannel <- request
	sendMessageResponse := <-request.ReplyChan
	responseJSON, err := json.Marshal(sendMessageResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}

func (server *GameServer) TransferObjectHandler(w http.ResponseWriter, r *http.Request) {
	var request TransferObjectRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	request.ReplyChan = make(chan TransferObjectResponse)
	server.transferObjectRequestChannel <- request
	sendMessageResponse := <-request.ReplyChan
	responseJSON, err := json.Marshal(sendMessageResponse)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseJSON))
	w.Write(responseJSON)
}
