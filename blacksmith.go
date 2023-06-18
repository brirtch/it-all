package main

type Blacksmith struct {
	BaseGameObject
}

func NewBlacksmith() *Blacksmith {
	blacksmith := &Blacksmith{BaseGameObject: BaseGameObject{Category: BuildingCategory, Type: "Blacksmith", WoodCost: 200, BuildAt: "Builder"}}
	return blacksmith
}

func (p *Blacksmith) Update() {
}
