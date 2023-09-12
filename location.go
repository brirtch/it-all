package main

type Location struct {
	Name string `json:"name"`
	Food int    `json:"food"`
	Wood int    `json:"wood"`
	Iron int    `json:"iron"`
	Gold int    `json:"gold"`
	Game *Game
}

// Return true if this location has been discovered/explored by a player.
func (loc *Location) Discovered() bool {

	for _, player := range loc.Game.Players {
		for _, playerLocation := range player.ExploredLocations {
			if playerLocation == loc.Name {
				return true
			}
		}
	}

	return false
}
