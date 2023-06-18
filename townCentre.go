package main

type TownCentre struct {
	BaseGameObject
}

func NewTownCentre() *TownCentre {
	tc := &TownCentre{BaseGameObject: BaseGameObject{Category: BuildingCategory, Type: "TownCentre", WoodCost: 400, BuildAt: "Builder"}}
	return tc
}

func (p *TownCentre) Update() {
}
