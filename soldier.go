package main

type Soldier struct {
	BaseGameObject
}

func NewSoldier() *Soldier {
	person := &Soldier{BaseGameObject: BaseGameObject{Category: PersonCategory, Type: "Soldier", FoodCost: 30, WoodCost: 30, Emoji: "🪖", BuildAt: "Barracks"}}
	return person
}

func (p *Soldier) Update() {
}
