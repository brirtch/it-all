package main

type Secateurs struct {
	BaseGameObject
}

func NewSecateurs() *Secateurs {
	sec := &Secateurs{BaseGameObject: BaseGameObject{Category: UpgradeCategory, Type: "Secateurs", BuildAt: "Blacksmith", WoodCost: 60, IronCost: 100, MaxItems: 1}}
	return sec
}

func (p *Secateurs) Update() {
}
