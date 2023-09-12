package main

import (
	"encoding/json"
	"fmt"
)

type GameObjectCategory string
type GameObjectType string
type Gender string

const (
	PersonCategory   GameObjectCategory = "PERSON"
	BuildingCategory GameObjectCategory = "BUILDING"
	UpgradeCategory  GameObjectCategory = "UPGRADE"
)

type BaseGameObject struct {
	Category GameObjectCategory `json:"category"`
	Type     GameObjectType     `json:"type"`
	Emoji    string             `json:"emoji"`
	Player   *Player            `json:"-"`
	BuildAt  string             `json:"buildAt"`
	Location string             `json:"location"`
	FoodCost int
	WoodCost int
	IronCost int
	GoldCost int
	MaxItems int // How many of these a player can have
}

type GameObject interface {
	Update()
	SetPlayer(p *Player)
	GetFoodCost() int
	GetWoodCost() int
	GetIronCost() int
	GetGoldCost() int
	GetType() string
	GetCategory() string
	GetMaxItems() int
	GetLocationName() string
	SetLocation(locationName string)
}

func (gameObject *BaseGameObject) SetPlayer(p *Player) {
	gameObject.Player = p
}

func (gameObject *BaseGameObject) GetFoodCost() int {
	return gameObject.FoodCost
}

func (gameObject *BaseGameObject) GetWoodCost() int {
	return gameObject.WoodCost
}

func (gameObject *BaseGameObject) GetIronCost() int {
	return gameObject.IronCost
}

func (gameObject *BaseGameObject) GetGoldCost() int {
	return gameObject.GoldCost
}

func (gameObject *BaseGameObject) GetEmoji() string {
	return gameObject.Emoji
}

func (gameObject *BaseGameObject) GetType() string {
	return string(gameObject.Type)
}

func (gameObject *BaseGameObject) GetCategory() string {
	return string(gameObject.Category)
}

func (gameObject *BaseGameObject) GetMaxItems() int {
	return gameObject.MaxItems
}

func (gameObject *BaseGameObject) GetLocationName() string {
	return gameObject.Location
}

func (gameObject *BaseGameObject) SetLocation(locationName string) {
	gameObject.Location = locationName
}

func (gameObject *BaseGameObject) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	m["category"] = string(gameObject.Category)
	m["type"] = string(gameObject.Type)
	m["foodCost"] = fmt.Sprint(gameObject.GetFoodCost())
	m["woodCost"] = fmt.Sprint(gameObject.GetWoodCost())
	m["ironCost"] = fmt.Sprint(gameObject.GetIronCost())
	m["buildAt"] = gameObject.BuildAt
	m["maxItems"] = fmt.Sprint(gameObject.MaxItems)
	m["location"] = gameObject.Location

	return json.Marshal(m)
}
