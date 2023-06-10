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
	World          *World
	Map            *Map
	Tick           int64
	HostPlayerName string
}

// World
type World struct {
	Width  int
	Height int
}

// Create a new Game.
func NewGame(hostPlayerName string) *Game {
	gameID := uuid.New()

	world := &World{Width: 40, Height: 40}
	g := &Game{GameID: gameID.String(), GameName: hostPlayerName + "'s game " + time.Now().Format("2006-1-2 15:4:5"), World: world, HostPlayerName: hostPlayerName}

	newMap := CreateMap(g, world.Width, world.Height)
	g.Map = newMap

	newMap.Game = g

	return g
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
