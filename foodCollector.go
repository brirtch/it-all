package main

type FoodCollector struct {
	BaseGameObject
}

func NewFoodCollector() *FoodCollector {
	person := &FoodCollector{BaseGameObject: BaseGameObject{Category: PersonCategory, Type: "FoodCollector", FoodCost: 60, BuildAt: "TownCentre"}}
	return person
}

func (p *FoodCollector) Update() {
	if p.Player.Game.Tick%60*3 == 0 {
		p.Player.Food += 1
		//move = false
	}
}
