package main

type Woodcutter struct {
	BaseGameObject
}

func NewWoodCutter() *Woodcutter {
	person := &Woodcutter{BaseGameObject: BaseGameObject{Category: PersonCategory, Type: "Woodcutter", FoodCost: 60, BuildAt: "TownCentre"}}
	return person
}

func (p *Woodcutter) Update() {
	if p.Player.Game.Tick%60*3 == 0 {
		p.Player.Wood += 1
		//move = false
	}
}
