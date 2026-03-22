package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Player holds the chicken's state.
type Player struct {
	x, y        float64
	vy          float64
	facingRight bool
	onGround    bool
}

func newPlayer() *Player {
	return &Player{
		x: 100,
		y: float64(groundY - spriteSize),
	}
}

// update processes input and applies physics.
// Returns true if the player just jumped this tick.
func (p *Player) update() (jumped bool) {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.x -= 2
		if p.x < 0 {
			p.x = 0
		}
		p.facingRight = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.x += 2
		if p.x > screenWidth-spriteSize {
			p.x = screenWidth - spriteSize
		}
		p.facingRight = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && p.onGround {
		p.vy = jumpSpeed
		p.onGround = false
		jumped = true
	}

	p.vy += gravity
	p.y += p.vy
	bottom := float64(groundY - spriteSize)
	if p.y >= bottom {
		p.y = bottom
		p.vy = 0
		p.onGround = true
	}
	return
}

// bounds returns the collision box, inset by a few pixels for fairness.
func (p *Player) bounds() rect {
	const pad = 8
	return rect{p.x + pad, p.y + pad, p.x + spriteSize - pad, p.y + spriteSize - pad}
}

// draw renders the chicken sprite, flipping it based on movement direction.
func (p *Player) draw(screen *ebiten.Image, img *ebiten.Image, shakeX, shakeY float64) {
	op := &ebiten.DrawImageOptions{}
	if p.facingRight {
		// Scale -0.5 flips horizontally; translate by +spriteSize to re-anchor to the left edge.
		op.GeoM.Scale(-0.5, 0.5)
		op.GeoM.Translate(p.x+spriteSize+shakeX, p.y+shakeY)
	} else {
		op.GeoM.Scale(0.5, 0.5)
		op.GeoM.Translate(p.x+shakeX, p.y+shakeY)
	}
	screen.DrawImage(img, op)
}
