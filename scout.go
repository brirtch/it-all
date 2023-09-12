package main

import (
	"math/rand"
)

type Scout struct {
	BaseGameObject
}

func NewScout() *Scout {
	person := &Scout{BaseGameObject: BaseGameObject{Category: PersonCategory, Type: "Scout", FoodCost: 60, BuildAt: "TownCentre"}}
	return person
}

func (p *Scout) Update() {
	if p.Player.Game.Tick%60*3 == 0 {
		//1 in 200 chance of discovering a new location.

		number := 1 + rand.Intn(20-1)
		if number == 10 {
			// Find an undiscovered location.
			newLoc := p.Player.Game.GetUnexploredLocation()
			if newLoc != nil {
				p.Player.ExploredLocations = append(p.Player.ExploredLocations, newLoc.Name)
			}
		}

		//move = false
	}
}
