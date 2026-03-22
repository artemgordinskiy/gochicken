package main

import "github.com/hajimehoshi/ebiten/v2"

// Obstacle is a hazard that scrolls in from the right.
// Its y-position is fixed to the ground.
type Obstacle struct {
	x float64
}

func newObstacle() *Obstacle {
	return &Obstacle{x: screenWidth + 10}
}

// bounds returns the collision box, inset slightly for fairness.
func (o *Obstacle) bounds() rect {
	const pad = 6
	y := float64(groundY - spriteSize)
	return rect{o.x + pad, y + pad, o.x + spriteSize - pad, y + spriteSize - pad}
}

// draw renders the obstacle sprite at ground level.
func (o *Obstacle) draw(screen *ebiten.Image, img *ebiten.Image, shakeX, shakeY float64) {
	y := float64(groundY - spriteSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(o.x+shakeX, y+shakeY)
	screen.DrawImage(img, op)
}
