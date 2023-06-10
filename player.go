package main

type Player struct {
	PlayerID string    `json:"playerID"`
	Name     string    `json:"playerName"`
	Food     int       `json:"food"`
	Wood     int       `json:"wood"`
	Gold     int       `json:"gold"`
	People   []*Person `json:"people"`
	Game     *Game     `json:"-"`
}

// Create a new Player.
func (game *Game) NewPlayer(playerID, playerName string) *Player {
	// Create a male and female person for new player.
	man := &Person{Role: Woodcutter, Health: 100, Gender: Male, GameObject: GameObject{X: 7, Y: 7, Type: "P", Map: game.Map}, Map: game.Map}
	woman := &Person{Role: FoodCollector, Health: 100, Gender: Female, GameObject: GameObject{X: 20, Y: 20, Type: "P", Map: game.Map}, Map: game.Map}
	people := []*Person{}
	people = append(people, man, woman)

	newPlayer := &Player{PlayerID: playerID, Name: playerName, People: people, Game: game}
	man.Player = newPlayer
	woman.Player = newPlayer
	return newPlayer
}

// Called on each tick of the game.
func (player *Player) Update() {
	// Move people north.
	for _, person := range player.People {
		person.Update()
	}
}

func (player *Player) Buy(itemName string, quantity int) {
	if itemName == "person" {
		// Add person.
	}
}
