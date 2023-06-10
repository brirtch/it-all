package main

type Map struct {
	Tiles             [][]*GameObject
	Game              *Game `json:"-"`
	Width             int
	Height            int
	NavigationTargetX int
	NavigationTargetY int
}

// Create a new map.
func CreateMap(g *Game, width, height int) *Map {
	newMap := &Map{Game: g, Width: width, Height: height}

	newMap.Tiles = make([][]*GameObject, height)
	for y := 0; y < height; y++ {
		newMap.Tiles[y] = make([]*GameObject, width)
		for x := 0; x < width; x++ {
			newMap.Tiles[y][x] = NewGameObject(x, y, BlockEmpty, newMap)
		}
	}

	// Add barrier
	for y := 2; y < 10; y++ {
		newMap.Tiles[y][10-y] = NewGameObject(10-y, y, BlockStone, newMap)
	}
	newMap.Tiles[0][13] = NewGameObject(13, 0, BlockStone, newMap)
	newMap.Tiles[1][13] = NewGameObject(13, 1, BlockStone, newMap)
	newMap.Tiles[2][13] = NewGameObject(13, 2, BlockStone, newMap)
	newMap.Tiles[3][12] = NewGameObject(12, 3, BlockStone, newMap)
	newMap.Tiles[4][11] = NewGameObject(11, 4, BlockStone, newMap)

	newMap.Tiles[1][1] = NewGameObject(1, 1, BlockTree, newMap)
	newMap.Tiles[height-5][width-5] = NewGameObject(width-5, height-5, BlockBerry, newMap)
	//newMap.Tiles[1][5] = NewGameObject(5, 1, BlockTree, newMap)
	//newMap.Tiles[42][50] = NewGameObject(50, 42, BlockTree, newMap)

	return newMap
}

// Returns true if x,y is on the map
func (m *Map) CoordOnMap(x, y int) bool {
	return y >= 0 && y < m.Height && x >= 0 && x < m.Width
}

// Gets an array of non-blank tiles (for sending map to client and saving space)
func (m *Map) GetNonBlankTiles() []*GameObject {
	nonBlankTiles := []*GameObject{}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Tiles[y][x].Type != BlockEmpty {
				nonBlankTiles = append(nonBlankTiles, m.Tiles[y][x])
			}
		}
	}
	return nonBlankTiles
}
