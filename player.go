package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Player holds the chicken's state.
type Player struct {
	x, y        float64
	vy          float64
	squishY     float64 // 1.0 = normal; >1 squash (land); <1 stretch (jump)
	facingRight bool
	onGround    bool
	jumpsLeft   int // 2 = both jumps available; double jump on first air press
	runTick     int // drives the head-bob sine wave
}

func newPlayer() *Player {
	return &Player{
		x:         100,
		y:         float64(groundY - spriteSize),
		squishY:   1.0,
		jumpsLeft: 2,
	}
}

// update processes input and physics.
// Returns (jumped, landed): jumped=true the tick a jump starts,
// landed=true the tick the chicken touches down.
func (p *Player) update() (jumped, landed bool) {
	wasOnGround := p.onGround

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

	if p.onGround {
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyRight) {
			p.runTick++
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && p.jumpsLeft > 0 {
		p.vy = jumpSpeed
		p.onGround = false
		p.jumpsLeft--
		p.squishY = 0.65 // stretch upward on jump
		jumped = true
	}

	p.vy += gravity
	p.y += p.vy

	bottom := float64(groundY - spriteSize)
	if p.y >= bottom {
		p.y = bottom
		if p.vy > 1 && !wasOnGround {
			p.squishY = 1.40 // squash on hard landing
			landed = true
		}
		p.vy = 0
		p.onGround = true
		p.jumpsLeft = 2
	}

	// Ease squishY back to 1.0
	p.squishY += (1.0 - p.squishY) * 0.18
	if math.Abs(p.squishY-1.0) < 0.005 {
		p.squishY = 1.0
	}

	return
}

// bounds returns the collision box, inset for fairness.
func (p *Player) bounds() rect {
	const pad = 8
	return rect{p.x + pad, p.y + pad, p.x + spriteSize - pad, p.y + spriteSize - pad}
}

// draw renders the chicken with squish/stretch and a running head-bob.
func (p *Player) draw(screen *ebiten.Image, img *ebiten.Image, shakeX, shakeY float64) {
	// Head bob: slight vertical oscillation while running on ground.
	bob := 0.0
	if p.onGround {
		bob = math.Sin(float64(p.runTick)*0.45) * 1.8
	}

	// Pin the bottom of the sprite to the ground when squashing/stretching.
	// drawY is the top-left Y of the sprite in screen space.
	drawY := p.y + float64(spriteSize)*(1-p.squishY) + bob + shakeY

	op := &ebiten.DrawImageOptions{}
	if p.facingRight {
		// Scale -0.5 flips horizontally. Translate right by spriteSize to
		// re-anchor the left edge of the sprite to p.x.
		op.GeoM.Scale(-0.5, 0.5*p.squishY)
		op.GeoM.Translate(p.x+spriteSize+shakeX, drawY)
	} else {
		op.GeoM.Scale(0.5, 0.5*p.squishY)
		op.GeoM.Translate(p.x+shakeX, drawY)
	}
	screen.DrawImage(img, op)
}
