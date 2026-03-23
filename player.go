package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Player holds the chicken's state.
type Player struct {
	x, y        float64
	vy          float64
	rotation    float64 // positive = nose up, negative = nose down
	facingRight bool
	onGround    bool
	jumpsLeft   int
	walkTick    int
}

func newPlayer() *Player {
	return &Player{
		x:           100,
		y:           float64(groundY - spriteSize),
		jumpsLeft:   2,
		facingRight: true,
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

	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyUp)) && p.jumpsLeft > 0 {
		p.vy = jumpSpeed
		p.onGround = false
		p.jumpsLeft--
		p.rotation = 0.45 // instant nose-up on flap
		jumped = true
	}

	p.vy += gravity
	p.y += p.vy

	bottom := float64(groundY - spriteSize)
	if p.y >= bottom {
		p.y = bottom
		if p.vy > 1 && !wasOnGround {
			landed = true
		}
		p.vy = 0
		p.onGround = true
		p.jumpsLeft = 2
	}

	// Walk animation: cycle legs when moving on ground.
	if p.onGround && (ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyRight)) {
		p.walkTick++
	} else {
		p.walkTick = 0
	}

	// Flappy-bird rotation: nose-up on jump, gradually tilts nose-down.
	if p.onGround {
		p.rotation += (0 - p.rotation) * 0.3
	} else {
		p.rotation -= 0.025
		if p.rotation < -1.5 {
			p.rotation = -1.5
		}
	}

	return
}

// bounds returns the collision box, inset for fairness.
func (p *Player) bounds() rect {
	const pad = 8
	return rect{p.x + pad, p.y + pad, p.x + spriteSize - pad, p.y + spriteSize - pad}
}

// drawShadow draws a ground shadow that shrinks and fades as the chicken rises.
func (p *Player) drawShadow(screen *ebiten.Image, shakeX, shakeY float64) {
	height := float64(groundY-spriteSize) - p.y
	if height < 1 {
		return
	}
	maxH := 120.0
	t := 1.0 - math.Min(height/maxH, 1.0)
	if t < 0.05 {
		return
	}
	alpha := uint8(t * 60)
	r := float32(6 + t*10)
	cx := float32(p.x + spriteSize/2 + shakeX)
	cy := float32(float64(groundY) + shakeY + 2)
	vector.DrawFilledCircle(screen, cx, cy, r, color.RGBA{0, 0, 0, alpha}, true)
}

// draw renders the chicken with flappy-bird-style rotation.
func (p *Player) draw(screen *ebiten.Image, img *ebiten.Image, shakeX, shakeY float64) {
	const imgSize = 100.0
	scale := float64(spriteSize) / imgSize

	centerX := p.x + float64(spriteSize)/2 + shakeX
	centerY := p.y + float64(spriteSize)/2 + shakeY

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-imgSize/2, -imgSize/2)

	sx := scale
	if p.facingRight {
		sx = -scale // flip to face right (source image faces left)
	}
	op.GeoM.Scale(sx, scale)

	rot := p.rotation
	if p.facingRight {
		rot = -rot // mirror rotation for flipped sprite
	}
	op.GeoM.Rotate(rot)
	op.GeoM.Translate(centerX, centerY)

	screen.DrawImage(img, op)
}

// walkFrame returns the current walk animation frame (0 or 1).
func (p *Player) walkFrame() int {
	return (p.walkTick / 6) % 2
}

// genChickenGroundImage creates the standing/walking pose (facing left).
// frame 0 and 1 alternate leg positions for walk animation.
func genChickenGroundImage(frame int) *ebiten.Image {
	const size = 100
	img := ebiten.NewImage(size, size)

	white := color.RGBA{255, 248, 238, 255}
	cream := color.RGBA{248, 240, 225, 255}
	wingGray := color.RGBA{230, 224, 214, 255}
	red := color.RGBA{210, 45, 45, 255}
	darkRed := color.RGBA{185, 35, 35, 255}
	orange := color.RGBA{235, 165, 50, 255}
	darkOrange := color.RGBA{210, 140, 40, 255}
	black := color.RGBA{25, 25, 25, 255}
	eyeHL := color.RGBA{255, 255, 255, 255}

	ci := func(x, y, r float32, col color.RGBA) {
		vector.DrawFilledCircle(img, x, y, r, col, true)
	}
	ri := func(x, y, w, h float32, col color.RGBA) {
		vector.DrawFilledRect(img, x, y, w, h, col, false)
	}

	// Tail — small feather bump at back
	ci(74, 40, 8, cream)
	ci(76, 36, 6, white)

	// Legs — walk cycle: one leg grounded, other lifted
	switch frame {
	case 0:
		ri(40, 72, 3, 22, orange)        // front leg (grounded)
		ri(36, 92, 9, 3, darkOrange)     // front foot
		ri(58, 72, 3, 18, orange)        // back leg (lifted)
		ri(56, 88, 7, 3, darkOrange)     // back foot (tucked)
	case 1:
		ri(42, 72, 3, 18, orange)        // front leg (lifted)
		ri(40, 88, 7, 3, darkOrange)     // front foot (tucked)
		ri(56, 72, 3, 22, orange)        // back leg (grounded)
		ri(52, 92, 9, 3, darkOrange)     // back foot
	}

	// Body — one round shape, head blends in (like reference sprite)
	ci(50, 50, 23, white)

	// Head/breast area — overlaps body to form continuous shape
	ci(36, 44, 11, white)

	// Wing fold (subtle shading on body)
	ci(58, 52, 13, wingGray)
	ci(57, 50, 10, cream)

	// Comb — prominent red bumps on top
	ci(34, 24, 6, red)
	ci(40, 22, 5.5, red)
	ci(46, 25, 4.5, red)

	// Beak — small, simple
	ri(20, 42, 9, 4, orange)

	// Wattle
	ci(26, 50, 3.5, darkRed)

	// Eye — simple dot with highlight
	ci(30, 38, 3, black)
	ci(28.5, 36.5, 1.2, eyeHL)

	return img
}

// genChickenAirImage creates the flying pose: wings raised, legs tucked.
func genChickenAirImage() *ebiten.Image {
	const size = 100
	img := ebiten.NewImage(size, size)

	white := color.RGBA{255, 248, 238, 255}
	cream := color.RGBA{248, 240, 225, 255}
	wingGray := color.RGBA{230, 224, 214, 255}
	red := color.RGBA{210, 45, 45, 255}
	darkRed := color.RGBA{185, 35, 35, 255}
	orange := color.RGBA{235, 165, 50, 255}
	black := color.RGBA{25, 25, 25, 255}
	eyeHL := color.RGBA{255, 255, 255, 255}

	ci := func(x, y, r float32, col color.RGBA) {
		vector.DrawFilledCircle(img, x, y, r, col, true)
	}
	ri := func(x, y, w, h float32, col color.RGBA) {
		vector.DrawFilledRect(img, x, y, w, h, col, false)
	}

	// Tail — spread in air
	ci(76, 38, 9, cream)
	ci(80, 32, 7, cream)
	ci(78, 46, 7, cream)
	ci(74, 34, 5, white)

	// Legs tucked — small dots
	ci(44, 74, 3, orange)
	ci(54, 74, 3, orange)

	// Body — same round shape as ground
	ci(50, 50, 23, white)

	// Head/breast
	ci(36, 44, 11, white)

	// Wing RAISED and spread
	ci(56, 34, 15, wingGray)
	ci(54, 32, 12, cream)
	ci(64, 24, 10, wingGray)
	ci(70, 18, 7, cream)

	// Comb
	ci(34, 24, 6, red)
	ci(40, 22, 5.5, red)
	ci(46, 25, 4.5, red)

	// Beak — open, squawking
	ri(20, 40, 9, 3, orange)
	ri(21, 44, 8, 3, orange)

	// Wattle
	ci(26, 50, 3.5, darkRed)

	// Eye — wider, surprised
	ci(30, 38, 3.5, black)
	ci(28.5, 36.5, 1.2, eyeHL)

	return img
}
