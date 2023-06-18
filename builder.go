package main

type Builder struct {
	BaseGameObject
}

func NewBuilder() *Builder {
	b := &Builder{BaseGameObject: BaseGameObject{Category: PersonCategory, Type: "Builder", FoodCost: 30, WoodCost: 30, Emoji: "🪖", BuildAt: "TownCentre"}}
	return b
}

func (p *Builder) Update() {
}
