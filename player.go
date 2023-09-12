package main

import (
	"fmt"
	"reflect"
)

type Player struct {
	PlayerID          string       `json:"playerID"`
	Name              string       `json:"playerName"`
	Food              int          `json:"food"`
	Wood              int          `json:"wood"`
	Iron              int          `json:"iron"`
	Gold              int          `json:"gold"`
	GameObjects       []GameObject `json:"gameObjects"`
	Upgrades          []string     `json:"upgrades"`
	Game              *Game        `json:"-"`
	Messages          []string     `json:"messages"`
	HomeLocation      string       `json:"homeLocation"`
	ExploredLocations []string     `json:"exploredLocations"`
}

// Create a new Player.
func (game *Game) NewPlayer(playerID, playerName string) *Player {
	// Create a male and female person for new player.
	man := NewWoodCutter()
	woman := NewFoodCollector()
	miner := NewMiner()
	tc := NewTownCentre()
	gameObjects := []GameObject{}
	gameObjects = append(gameObjects, man, woman, miner, tc)

	// Find a home location that is unoccupied.
	homeLocation := ""
	for _, loc := range game.Locations {
		isEmpty := true
		for _, player := range game.Players {
			if player.HomeLocation == loc.Name {
				isEmpty = false
			}
		}
		if isEmpty {
			homeLocation = loc.Name
			break
		}
	}

	man.Location = homeLocation
	woman.Location = homeLocation
	miner.Location = homeLocation
	tc.Location = homeLocation

	exploredLocations := []string{}
	exploredLocations = append(exploredLocations, homeLocation)

	newPlayer := &Player{PlayerID: playerID, Name: playerName, GameObjects: gameObjects, Game: game, Food: 60, HomeLocation: homeLocation, ExploredLocations: exploredLocations}
	man.Player = newPlayer
	woman.Player = newPlayer
	miner.Player = newPlayer
	return newPlayer
}

// Called on each tick of the game.
func (player *Player) Update() {
	for _, gameObject := range player.GameObjects {
		gameObject.Update()
	}
}

func (player *Player) GetGameObjectTally() []*GameObjectTally {
	var gameObjectTally []*GameObjectTally

	for _, location := range player.ExploredLocations {
		tallyMap := make(map[string]int)
		categoryMap := make(map[string]string)
		for _, gameObject := range player.GameObjects {
			if gameObject.GetLocationName() == location {
				tallyMap[gameObject.GetType()] += 1
				categoryMap[gameObject.GetType()] = gameObject.GetCategory()
			}
		}

		for objectType, quantity := range tallyMap {
			tally := &GameObjectTally{Location: location, Type: objectType, Quantity: quantity, Category: categoryMap[objectType]}
			gameObjectTally = append(gameObjectTally, tally)
		}
	}

	return gameObjectTally
}

// Transfer object from one sourceLocation to targetLocation
func (player *Player) TransferObject(objectType, sourceLocation, targetLocation string) {
	// Find an object of the objectType to transfer, then update its location.
	for _, gameObject := range player.GameObjects {
		if gameObject.GetType() == objectType && gameObject.GetLocationName() == sourceLocation {
			gameObject.SetLocation(targetLocation)
			return
		}
	}
}

// Send a message to another player.
func (player *Player) SendMessage(recipientPlayer *Player, messageBody string) {
	senderName := player.Name
	recipientName := recipientPlayer.Name
	recipientPlayer.Messages = append(recipientPlayer.Messages, fmt.Sprintf("Message from %s: %s", senderName, messageBody))
	player.Messages = append(player.Messages, fmt.Sprintf("Sent message to %s: %s", recipientName, messageBody))
}

func (player *Player) Attack(target *Player, soldiersToCommit int) string {
	// For now, just take away one of the target's people.
	killCount := 0
	for i := 0; i < soldiersToCommit; i++ {
		if len(target.GameObjects) > 0 {
			target.GameObjects = target.GameObjects[1:]
			killCount++
		}

		if len(player.GameObjects) > 0 {
			player.GameObjects = player.GameObjects[1:]
		}
	}

	target.Messages = append(target.Messages, fmt.Sprintf("You were attacked by %d %s and lost %d people", killCount, player.Name, killCount))

	return fmt.Sprintf("You destroyed %d of %s's people and lost %d of your people", killCount, target.Name, killCount)
}

// Gets the number items of itemClass owned by player
func (player *Player) GetItemCount(itemClass string) int {
	count := 0
	for _, gameObject := range player.GameObjects {
		if gameObject.GetType() == itemClass {
			count++
		}
	}
	return count
}

func (player *Player) Buy(itemClass string, quantity int, location string) bool {
	server := player.Game.Server

	f := reflect.ValueOf(server.GameObjectTypes[itemClass])
	retVal := f.Call([]reflect.Value{})                // This calls the NewItemclass function.
	retInterface := retVal[0].Interface().(GameObject) // This gets the return value
	retInterface.SetPlayer(player)
	retInterface.SetLocation(location)

	// Check if maxItems reached.
	if retInterface.GetMaxItems() == 0 || player.GetItemCount(itemClass) < retInterface.GetMaxItems() {
		// Check if player can afford this.
		if player.Food >= retInterface.GetFoodCost()*quantity && player.Wood >= retInterface.GetWoodCost()*quantity && player.Iron >= retInterface.GetIronCost()*quantity && player.Gold >= retInterface.GetGoldCost()*quantity {
			player.Food -= retInterface.GetFoodCost() * quantity
			player.Wood -= retInterface.GetWoodCost() * quantity
			player.Iron -= retInterface.GetIronCost() * quantity
			player.Gold -= retInterface.GetGoldCost() * quantity
			player.GameObjects = append(player.GameObjects, retInterface)
			return true
		} else {
			return false
		}
	} else {
		return false
	}

}
