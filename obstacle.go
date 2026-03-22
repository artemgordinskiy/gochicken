package main

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// Obstacle is a hazard that scrolls in from the right.
// Its y-position is fixed to the ground.
type Obstacle struct {
	x     float64
	tintR uint8
	tintG uint8
	tintB uint8
}

// Tint palettes give each obstacle a slightly different look.
var obstacleTints = [][3]uint8{
	{255, 255, 255}, // plain
	{255, 160, 80},  // warm orange
	{100, 200, 255}, // cool blue
	{180, 255, 140}, // lime green
	{255, 100, 120}, // pink-red
}

func newObstacle() *Obstacle {
	t := obstacleTints[rand.Intn(len(obstacleTints))]
	return &Obstacle{
		x:     screenWidth + 10,
		tintR: t[0], tintG: t[1], tintB: t[2],
	}
}

// bounds returns the collision box, inset slightly for fairness.
func (o *Obstacle) bounds() rect {
	const pad = 6
	y := float64(groundY - spriteSize)
	return rect{o.x + pad, y + pad, o.x + spriteSize - pad, y + spriteSize - pad}
}

// draw renders the obstacle sprite at ground level with its colour tint.
func (o *Obstacle) draw(screen *ebiten.Image, img *ebiten.Image, shakeX, shakeY float64) {
	y := float64(groundY - spriteSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(o.x+shakeX, y+shakeY)
	op.ColorScale.Scale(
		float32(o.tintR)/255,
		float32(o.tintG)/255,
		float32(o.tintB)/255,
		1,
	)
	screen.DrawImage(img, op)
}
