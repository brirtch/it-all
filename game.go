// Main game code.
package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	GameID         string //uuid.
	GameName       string
	Players        []*Player
	Locations      []*Location
	Tick           int64
	HostPlayerName string
	Server         *GameServer
}

// Create a new Game.
func NewGame(server *GameServer, hostPlayerName string) *Game {
	gameID := uuid.New()

	locationNames := []string{"London", "Birmingham", "Chester", "Edinburgh", "Glasgow", "York", "Manchester", "St. Ives", "Leeds", "Inverness", "Bristol", "Northampton", "Leicester", "Derby", "Nottingham", "Coventry", "Cambridge", "Oxford", "Canterbury", "Colchester", "Ipswitch", "Newcastle", "Norwich", "Bath", "Exeter", "Hull", "Sheffield"}
	var locations []*Location
	for _, name := range locationNames {
		loc := &Location{Name: name, Food: 10000, Wood: 10000, Iron: 5000, Gold: 2500}
		locations = append(locations, loc)
	}

	g := &Game{GameID: gameID.String(), GameName: hostPlayerName + "'s game " + time.Now().Format("2006-1-2 15:4:5"), HostPlayerName: hostPlayerName, Server: server, Locations: locations}
	for _, location := range g.Locations {
		location.Game = g
	}

	return g
}

// Get an unexplored location.
func (g *Game) GetUnexploredLocation() *Location {
	for _, location := range g.Locations {
		if !location.Discovered() {
			return location
		}
	}

	return nil
}

// Save the game to file.
func (g *Game) Save(filename string) {

}

// Update the game status on a tick of the game. The loop is managed by the server.
func (g *Game) Update(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, player := range g.Players {
		player.Update()
	}

	g.Tick += 1
}

// Find a player using player Name.
func (g *Game) GetPlayerByName(playerName string) *Player {
	for _, player := range g.Players {
		if player.Name == playerName {
			return player
		}
	}

	return nil
}

// Find a player using the player ID.
func (g *Game) GetPlayerByID(playerID string) *Player {
	for _, player := range g.Players {
		if player.PlayerID == playerID {
			return player
		}
	}

	return nil
}
