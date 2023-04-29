// Main game code.
package main

import (
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	GameID         string //uuid.
	GameName       string
	Players        []*Player
	Map            *World
	GameObjects    []*GameObject
	Tick           int64
	HostPlayerName string
}

type GameObject struct {
	Type byte `json:"type"`
	X    int  `json:"x"`
	Y    int  `json:"y"`
}

// World
type World struct {
	Width  int
	Height int
}

type Map struct {
	MapBlocks []*MapBlock
}

// MapBlock represents a single block on a map grid.
type MapBlock struct {
	X    int  `json:"y"`
	Y    int  `json:"x"`
	Type byte `json:"type"`
}

type PersonRole string
type Gender string

const (
	Peasant PersonRole = "PEASANT"
	Soldier PersonRole = "SOLDIER"
)

const (
	BlockEmpty = iota
	BlockTree
)

const (
	Male   Gender = "M"
	Female Gender = "F"
)

// Create a new Game.
func NewGame(hostPlayerName string) *Game {
	gameID := uuid.New()
	m := &World{Width: 100, Height: 100}

	tree1 := &GameObject{Type: BlockTree, X: 1, Y: 1}
	tree2 := &GameObject{Type: BlockTree, X: 5, Y: 1}
	tree3 := &GameObject{Type: BlockTree, X: 2, Y: 2}
	var gameObjects []*GameObject
	gameObjects = append(gameObjects, tree1, tree2, tree3)
	g := Game{GameID: gameID.String(), GameName: hostPlayerName + "'s game " + time.Now().Format("2006-1-2 15:4:5"), Map: m, GameObjects: gameObjects, HostPlayerName: hostPlayerName}

	return &g
}

type Person struct {
	Role   PersonRole `json:"role"`
	Health int        `json:"health"`
	Gender Gender     `json:"gender"`
	X      int        `json:"x"`
	Y      int        `json:"y"`
	Map    *Map
	Player *Player
}

// Updates this person's map based on the person's current position.
func (p *Person) UpdateMap() {
	// 5 blocks around the player in all directions.
	for x := int(math.Max(0, float64(p.X-5))); x < p.X+5; x++ {
		for y := int(math.Max(0, float64(p.Y-5))); y < p.Y+5; y++ {
			p.Map.Set(x, y, p.Player.Game.GetItemAt(x, y))
		}
	}
}

// Sets the item at coordinates x, y on the map.
func (m *Map) Set(x, y int, itemType byte) {
	found := false
	for _, mapBlock := range m.MapBlocks {
		if mapBlock.X == x && mapBlock.Y == y {
			mapBlock.Type = itemType
			found = true
		}
	}

	if !found {
		mapBlock := MapBlock{X: x, Y: y, Type: itemType}
		m.MapBlocks = append(m.MapBlocks, &mapBlock)
	}
}

// Get the item at coordinates x,y, including gameobjects and people.
func (g *Game) GetItemAt(x, y int) byte {
	// Check game objects.
	for _, gameObject := range g.GameObjects {
		if gameObject.X == x && gameObject.Y == y {
			return gameObject.Type
		}
	}

	// Check for person.

	return 0 // blank
}

// Save the game to file.
func (g *Game) Save(filename string) {

}

// Update the game status on a tick of the game. The loop is managed by the server.
func (g *Game) Update(wg *sync.WaitGroup) {
	defer wg.Done()

	for _, player := range g.Players {
		// Move people north.
		for _, person := range player.People {
			if g.Tick%60*3 == 0 {
				person.Y = int(math.Max(0, float64(person.Y-1)))
			}
		}
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
