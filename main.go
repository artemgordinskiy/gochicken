package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Chicken Game")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
