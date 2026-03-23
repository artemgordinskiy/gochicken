package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Obstacle is a hazard that scrolls in from the right.
type Obstacle struct {
	x          float64
	variant    int
	brightness float32
}

func newObstacle(numVariants int) *Obstacle {
	return &Obstacle{
		x:          screenWidth + 10,
		variant:    rand.Intn(numVariants),
		brightness: 0.85 + rand.Float32()*0.30,
	}
}

// Per-variant hitboxes (in display coords relative to sprite origin).
// Each is {x1, y1, x2, y2} tightly fitted to the visible shape.
var obsHitboxes = [][4]float64{
	{9, 14, 41, 49},  // round boulder
	{14, 10, 36, 49}, // pointed rock (narrow!)
	{8, 14, 42, 49},  // rough boulder
	{10, 8, 40, 50},  // cactus
}

// bounds returns the collision box fitted to this obstacle variant.
func (o *Obstacle) bounds() rect {
	hb := obsHitboxes[o.variant]
	y := float64(groundY - spriteSize)
	return rect{o.x + hb[0], y + hb[1], o.x + hb[2], y + hb[3]}
}

// draw renders the obstacle sprite at ground level.
func (o *Obstacle) draw(screen *ebiten.Image, img *ebiten.Image, shakeX, shakeY float64) {
	y := float64(groundY - spriteSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(o.x+shakeX, y+shakeY)
	op.ColorScale.Scale(o.brightness, o.brightness, o.brightness, 1)
	screen.DrawImage(img, op)
}

// genObstacleImages creates the set of procedural obstacle variants.
func genObstacleImages() []*ebiten.Image {
	return []*ebiten.Image{
		genRoundBoulder(),
		genPointedRock(),
		genRoughBoulder(),
		genCactusImage(),
	}
}

func genRoundBoulder() *ebiten.Image {
	const size = 100
	img := ebiten.NewImage(size, size)
	c := func(x, y, r float32, col color.RGBA) {
		vector.DrawFilledCircle(img, x, y, r, col, true)
	}
	c(50, 65, 36, color.RGBA{140, 132, 122, 255}) // shadow base
	c(50, 55, 35, color.RGBA{190, 183, 173, 255}) // main body
	c(40, 58, 28, color.RGBA{180, 173, 163, 255})
	c(62, 57, 26, color.RGBA{185, 178, 168, 255})
	c(48, 47, 25, color.RGBA{210, 203, 193, 255}) // upper surface
	c(56, 49, 19, color.RGBA{220, 213, 203, 255})
	c(44, 41, 14, color.RGBA{235, 228, 220, 255}) // highlight
	c(63, 65, 7, color.RGBA{150, 143, 133, 255})  // crevices
	c(34, 63, 6, color.RGBA{155, 148, 138, 255})
	return img
}

func genPointedRock() *ebiten.Image {
	const size = 100
	img := ebiten.NewImage(size, size)
	c := func(x, y, r float32, col color.RGBA) {
		vector.DrawFilledCircle(img, x, y, r, col, true)
	}
	dark := color.RGBA{125, 118, 108, 255}
	mid := color.RGBA{165, 158, 148, 255}
	light := color.RGBA{198, 191, 181, 255}
	hi := color.RGBA{218, 211, 201, 255}
	c(50, 80, 28, dark)  // wide base
	c(50, 68, 24, mid)   // mid section
	c(50, 55, 20, mid)   // upper mid
	c(48, 48, 17, light) // narrowing
	c(50, 38, 14, light) // top section
	c(48, 28, 10, hi)    // near tip
	c(50, 20, 6, hi)     // pointed tip
	c(38, 72, 5, dark)   // crevice
	c(60, 75, 4, dark)
	return img
}

func genRoughBoulder() *ebiten.Image {
	const size = 100
	img := ebiten.NewImage(size, size)
	c := func(x, y, r float32, col color.RGBA) {
		vector.DrawFilledCircle(img, x, y, r, col, true)
	}
	base := color.RGBA{115, 108, 100, 255}
	mid := color.RGBA{150, 143, 133, 255}
	light := color.RGBA{175, 168, 158, 255}
	dark := color.RGBA{88, 82, 76, 255}
	c(50, 68, 34, dark)  // dark base
	c(45, 58, 30, base)  // left body
	c(58, 60, 26, base)  // right body
	c(35, 50, 18, mid)   // left bump
	c(62, 48, 20, mid)   // right bump
	c(48, 42, 16, light) // top
	c(42, 36, 10, light) // highlight
	c(65, 68, 8, dark)   // dark spots
	c(30, 63, 7, dark)
	c(55, 73, 6, dark)
	return img
}

func genCactusImage() *ebiten.Image {
	const size = 100
	img := ebiten.NewImage(size, size)
	green := color.RGBA{65, 145, 55, 255}
	darkGreen := color.RGBA{42, 115, 38, 255}
	lightGreen := color.RGBA{88, 172, 72, 255}
	spine := color.RGBA{225, 220, 185, 255}
	ci := func(x, y, r float32, col color.RGBA) {
		vector.DrawFilledCircle(img, x, y, r, col, true)
	}
	ri := func(x, y, w, h float32, col color.RGBA) {
		vector.DrawFilledRect(img, x, y, w, h, col, false)
	}
	// Main trunk
	ri(38, 15, 24, 85, green)
	ci(50, 17, 12, green)
	// Left arm
	ri(16, 38, 22, 12, green)
	ri(16, 24, 12, 26, green)
	ci(22, 24, 6, green)
	// Right arm
	ri(62, 48, 22, 12, green)
	ri(74, 32, 12, 28, green)
	ci(80, 32, 6, green)
	// Highlights (lighter centre stripe)
	ri(45, 20, 6, 70, lightGreen)
	ri(19, 28, 5, 18, lightGreen)
	ri(76, 36, 5, 20, lightGreen)
	// Dark edges
	ri(38, 15, 4, 80, darkGreen)
	ri(58, 15, 4, 80, darkGreen)
	// Spines
	for _, p := range [][2]float32{
		{50, 8}, {35, 28}, {65, 28},
		{50, 45}, {35, 58}, {65, 58},
		{50, 72}, {13, 32}, {87, 42},
		{22, 18}, {80, 26},
	} {
		ci(p[0], p[1], 2, spine)
	}
	return img
}
