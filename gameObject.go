package main

import (
	"math"

	"github.com/beefsack/go-astar"
)

const (
	BlockEmpty  = " "
	BlockPerson = "P"
	BlockTree   = "T"
	BlockStone  = "S"
	BlockBerry  = "B"
)

var (
	directions = map[string][]int{
		"N":  {-1, 0},
		"S":  {1, 0},
		"W":  {0, -1},
		"E":  {0, 1},
		"NW": {-1, -1},
		"NE": {-1, 1},
		"SW": {1, -1},
		"SE": {1, 1},
	}
)

type GameObject struct {
	Type   BlockType `json:"type"`
	X      int       `json:"x"`
	Y      int       `json:"y"`
	Health int       `json:"health"`
	Map    *Map      `json:"-"`
}

type BlockType string

func NewGameObject(x, y int, blockType BlockType, m *Map) *GameObject {
	return &GameObject{X: x, Y: y, Type: blockType, Map: m}
}

// For A* implementing Pather
func (gameObject *GameObject) PathNeighbors() []astar.Pather {
	neighbours := []astar.Pather{}
	potentialNeighbours := gameObject.Neighbours()
	for _, neighbour := range potentialNeighbours {
		if !neighbour.Blocker() {
			neighbours = append(neighbours, neighbour)
		}
	}
	return neighbours
}

func (gameObject *GameObject) PathNeighborCost(to astar.Pather) float64 {
	return 1
}

func (gameObject *GameObject) PathEstimatedCost(to astar.Pather) float64 {
	toGO := to.(*GameObject)
	absX := toGO.X - gameObject.X
	if absX < 0 {
		absX = -absX
	}
	absY := toGO.Y - gameObject.Y
	if absY < 0 {
		absY = -absY
	}
	return float64(absX + absY)
}

// Returns true if this gameobject blocks a path (except the targetBlock)
func (gameObject *GameObject) Blocker() bool {
	if gameObject.X == gameObject.Map.NavigationTargetX && gameObject.Y == gameObject.Map.NavigationTargetY {
		return false
	} else {
		return gameObject.Type != BlockEmpty
	}
}

// Find the nearest tile of the given type.
func (gameObject *GameObject) NearestTileOfType(blockType BlockType) *GameObject {
	mapWidth := gameObject.Map.Game.World.Width
	mapHeight := gameObject.Map.Game.World.Height

	maxDistanceToWorldEdge := math.Max(
		math.Max(float64(gameObject.X), float64(mapWidth-gameObject.X)),
		math.Max(float64(gameObject.Y), float64(mapHeight-gameObject.Y)),
	)

	for distance := 1; distance <= int(maxDistanceToWorldEdge); distance++ {
		for x := gameObject.X - distance; x <= gameObject.X+distance; x++ {
			if x == gameObject.X-distance || x == gameObject.X+distance {
				for y := gameObject.Y - distance; y <= gameObject.Y+distance; y++ {
					if gameObject.Map.CoordOnMap(x, y) {
						tile := gameObject.Map.Tiles[y][x]
						if tile != nil && tile.Type == blockType {
							return tile
						}
					}
				}
			} else {
				if gameObject.Map.CoordOnMap(x, gameObject.Y-distance) {
					topTile := gameObject.Map.Tiles[gameObject.Y-distance][x]
					if topTile != nil && topTile.Type == blockType {
						return topTile
					}
				}
				if gameObject.Map.CoordOnMap(x, gameObject.Y+distance) {
					bottomTile := gameObject.Map.Tiles[gameObject.Y+distance][x]

					if bottomTile != nil && bottomTile.Type == blockType {
						return bottomTile
					}
				}

			}
		}
	}

	return nil
}

// Looks for an adjacent block on the map of the type specified.
func (gameObject *GameObject) NeighbourOfType(blockType BlockType) *GameObject {
	neighbours := gameObject.Neighbours()
	for _, neighbour := range neighbours {
		if neighbour.Type == blockType {
			return neighbour
		}
	}
	return nil
}

// Gets a slice of the adjacent game objects
func (gameObject *GameObject) Neighbours() []*GameObject {
	potentialNeighbours := []*GameObject{}

	potentialNeighbours = append(potentialNeighbours,
		gameObject.Up(),
		gameObject.Down(),
		gameObject.Left(),
		gameObject.Right(),
		//gameObject.UpLeft(),
		//gameObject.UpRight(),
		//gameObject.DownLeft(),
		//gameObject.DownRight()
	)

	neighbours := []*GameObject{}
	for _, potentialNeighbour := range potentialNeighbours {
		if potentialNeighbour != nil {
			neighbours = append(neighbours, potentialNeighbour)
		}
	}
	return neighbours
}

// Gets a neighbouring tile in a cardinal direction e.g. N, NE, etc.
func (gameObject *GameObject) GetNeighbourInDirection(direction string) *GameObject {
	dirVector := directions[direction]
	neighbourY := gameObject.Y + dirVector[0]
	neighbourX := gameObject.X + dirVector[1]
	if gameObject.Map.CoordOnMap(neighbourX, neighbourY) {
		return gameObject.Map.Tiles[neighbourY][neighbourX]
	} else {
		return nil
	}
}

// Returns the gameobject above me on the gameMap.
func (gameObject *GameObject) Up() *GameObject {
	return gameObject.GetNeighbourInDirection("N")
}

// Returns the gameobject below me on the gameMap.
func (gameObject *GameObject) Down() *GameObject {
	return gameObject.GetNeighbourInDirection("S")
}

// Returns the gameobject to the left of me on the gameMap.
func (gameObject *GameObject) Left() *GameObject {
	return gameObject.GetNeighbourInDirection("W")
}

// Returns the gameobject to the right of me on the gameMap.
func (gameObject *GameObject) Right() *GameObject {
	return gameObject.GetNeighbourInDirection("E")
}

// Returns the gameobject to above left of me on the gameMap
func (gameObject *GameObject) UpLeft() *GameObject {
	return gameObject.GetNeighbourInDirection("NW")
}

// Returns the gameobject to above right of me on the gameMap
func (gameObject *GameObject) UpRight() *GameObject {
	return gameObject.GetNeighbourInDirection("NE")
}

// Returns the gameobject to below left of me on the gameMap
func (gameObject *GameObject) DownLeft() *GameObject {
	return gameObject.GetNeighbourInDirection("SW")
}

// Returns the gameobject to below right of me on the gameMap
func (gameObject *GameObject) DownRight() *GameObject {
	return gameObject.GetNeighbourInDirection("SE")
}
