package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type particle struct {
	x, y    float64
	vx, vy  float64
	life    float64 // 1 → 0
	decay   float64
	size    float64
	r, g, b uint8
}

func (p *particle) update() bool {
	p.x += p.vx
	p.y += p.vy
	p.vy += 0.18 // particle gravity
	p.life -= p.decay
	p.size *= 0.96
	return p.life > 0
}

// Particles manages a pool of active particles.
type Particles struct {
	pool []particle
}

func (ps *Particles) emit(x, y float64, count int, r, g, b uint8, speed, decayBase float64) {
	for range count {
		angle := rand.Float64() * 2 * math.Pi
		spd := speed * (0.5 + rand.Float64())
		ps.pool = append(ps.pool, particle{
			x: x, y: y,
			vx:    math.Cos(angle) * spd,
			vy:    math.Sin(angle)*spd - speed*0.6,
			life:  1,
			decay: decayBase + rand.Float64()*0.03,
			size:  2 + rand.Float64()*4,
			r:     r, g: g, b: b,
		})
	}
}

// EmitDust spawns sandy dust when the chicken lands.
func (ps *Particles) EmitDust(x, y float64) {
	ps.emit(x+spriteSize/2, y+spriteSize, 7, 200, 170, 110, 1.2, 0.04)
}

// EmitExplosion spawns a fiery burst on death.
func (ps *Particles) EmitExplosion(x, y float64) {
	cx, cy := x+spriteSize/2, y+spriteSize/2
	ps.emit(cx, cy, 16, 255, 80, 0, 2.5, 0.025)
	ps.emit(cx, cy, 10, 255, 215, 0, 2.0, 0.030)
	ps.emit(cx, cy, 8, 255, 40, 40, 3.0, 0.020)
}

// EmitScorePop spawns golden sparkles when an obstacle is cleared.
func (ps *Particles) EmitScorePop(x, y float64) {
	ps.emit(x, y, 6, 255, 220, 50, 1.5, 0.045)
}

func (ps *Particles) Update() {
	alive := ps.pool[:0]
	for i := range ps.pool {
		if ps.pool[i].update() {
			alive = append(alive, ps.pool[i])
		}
	}
	ps.pool = alive
}

func (ps *Particles) Draw(screen *ebiten.Image, shakeX, shakeY float64) {
	for _, p := range ps.pool {
		c := color.RGBA{p.r, p.g, p.b, uint8(p.life * 230)}
		vector.DrawFilledCircle(screen,
			float32(p.x+shakeX), float32(p.y+shakeY),
			float32(p.size*p.life+0.5), c, true)
	}
}
