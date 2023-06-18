package main

type Barracks struct {
	BaseGameObject
}

func NewBarracks() *Barracks {
	barracks := &Barracks{BaseGameObject: BaseGameObject{Category: BuildingCategory, Type: "Barracks", WoodCost: 200, BuildAt: "Builder"}}
	return barracks
}

func (p *Barracks) Update() {
}
