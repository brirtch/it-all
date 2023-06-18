package main

import (
	"math/rand"
)

type Miner struct {
	BaseGameObject
}

func NewMiner() *Miner {
	person := &Miner{BaseGameObject: BaseGameObject{Category: PersonCategory, Type: "Miner", FoodCost: 60, BuildAt: "TownCentre"}}
	return person
}

func (p *Miner) Update() {
	if p.Player.Game.Tick%60*3 == 0 {
		//90 chance of iron, 10 of gold.
		number := 1 + rand.Intn(10-1)
		if number <= 7 {
			p.Player.Iron += 1
		} else {
			p.Player.Gold += 1
		}

		//move = false
	}
}
