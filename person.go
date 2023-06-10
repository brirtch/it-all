package main

import (
	"math"

	"github.com/beefsack/go-astar"
)

const (
	Male   Gender = "M"
	Female Gender = "F"
)

type PersonRole string
type Gender string

const (
	Peasant       PersonRole = "PEASANT"
	Soldier       PersonRole = "SOLDIER"
	Woodcutter    PersonRole = "WOODCUTTER"
	FoodCollector PersonRole = "FOODCOLLECTOR"
)

type Person struct {
	GameObject
	Role   PersonRole `json:"role"`
	Health int        `json:"health"`
	Gender Gender     `json:"gender"`
	Map    *Map
	Player *Player `json:"-"`
}

// Updates this person's map based on the person's current position.
func (p *Person) UpdateMap() {
	// 5 blocks around the player in all directions.
	for x := int(math.Max(0, float64(p.X-5))); x < p.X+5; x++ {
		for y := int(math.Max(0, float64(p.Y-5))); y < p.Y+5; y++ {
			//p.Map.Set(x, y, p.Player.Game.Map.GetItemAt(x, y).Type)
		}
	}
}

// Called each tick of the game.
func (p *Person) Update() {
	//move := true
	if p.Role == Woodcutter {
		// if tree adjacent, chop.
		adjacentTree := p.GameObject.NeighbourOfType(BlockTree)
		if adjacentTree != nil {
			if p.Player.Game.Tick%60*3 == 0 {
				p.Player.Wood += 1
				//move = false
			}
		} else {
			// Find the nearest tree.

			nearestTree := p.NearestTileOfType(BlockTree)

			if p.Player.Game.Tick%60*3 == 0 {
				if nearestTree != nil {
					p.Map.NavigationTargetX = nearestTree.X
					p.Map.NavigationTargetY = nearestTree.Y
					path, _, found := astar.Path(&p.GameObject, nearestTree)
					if found {
						nextNode := path[len(path)-2].(*GameObject)
						p.GameObject.X = nextNode.X
						p.GameObject.Y = nextNode.Y
					}
				}
			}
		}

	}

	if p.Role == FoodCollector {
		// if tree adjacent, chop.
		adjacentBerry := p.GameObject.NeighbourOfType(BlockBerry)
		if adjacentBerry != nil {
			if p.Player.Game.Tick%60*3 == 0 {
				p.Player.Food += 1
				//move = false
			}
		} else {
			// Find the nearest tree.

			nearestTree := p.NearestTileOfType(BlockBerry)

			if p.Player.Game.Tick%60*3 == 0 {
				if nearestTree != nil {
					p.Map.NavigationTargetX = nearestTree.X
					p.Map.NavigationTargetY = nearestTree.Y
					path, _, found := astar.Path(&p.GameObject, nearestTree)
					if found {
						nextNode := path[len(path)-2].(*GameObject)
						p.GameObject.X = nextNode.X
						p.GameObject.Y = nextNode.Y
					}
				}
			}
		}

	}
	/*if move && p.Player.Game.Tick%60*3 == 0 {
		p.GameObject.Y = int(math.Max(0, float64(p.Y-1)))
	}*/
}
