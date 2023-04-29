package main

type Player struct {
	PlayerID string    `json:"playerID"`
	Name     string    `json:"playerName"`
	Food     int       `json:"food"`
	Wood     int       `json:"wood"`
	People   []*Person `json:"people"`
	Game     *Game
}

// Create a new Player.
func (game *Game) NewPlayer(playerID, playerName string) *Player {
	// Create a male and female person for new player.
	man := &Person{Role: Peasant, Health: 100, Gender: Male, X: 50, Y: 50}
	woman := &Person{Role: Peasant, Health: 100, Gender: Female, X: 52, Y: 52}
	people := []*Person{}
	people = append(people, man, woman)

	newPlayer := &Player{PlayerID: playerID, Name: playerName, People: people}
	man.Player = newPlayer
	woman.Player = newPlayer
	return newPlayer
}
