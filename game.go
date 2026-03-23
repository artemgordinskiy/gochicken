package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 480
	// groundY is the y-coordinate of the ground surface.
	groundY    = screenHeight - 111
	spriteSize = 50

	gravity   = 0.22
	jumpSpeed = -5.2

	sampleRate = 44100
)

type GameState int

const (
	StateTitle GameState = iota
	StatePlaying
	StateGameOver
)

// rect is an axis-aligned bounding box used for collision detection.
type rect struct{ x1, y1, x2, y2 float64 }

func (r rect) overlaps(o rect) bool {
	return r.x2 > o.x1 && r.x1 < o.x2 && r.y2 > o.y1 && r.y1 < o.y2
}

type Game struct {
	state     GameState
	player    *Player
	obstacles []*Obstacle
	particles *Particles
	score     int
	highScore int
	ticks     int
	speed     float64
	bgX       float64

	spawnTimer   int
	speedUpTimer int // frames to show "SPEED UP!" banner

	// flash overlay
	flashTimer int
	flashR     uint8
	flashG     uint8
	flashB     uint8

	// title-screen demo animation
	titleY    float64 // chicken Y offset from ground (positive = above)
	titleVY   float64
	titleTick int

	// screen-shake effect
	shakeTicksVal int
	shakeX        float64
	shakeY        float64

	// images
	bgImage      *ebiten.Image
	chickenGround [2]*ebiten.Image
	chickenAir    *ebiten.Image
	obsImages    []*ebiten.Image

	// audio
	audioCtx       *audio.Context
	jumpSound      []byte
	hitSound       []byte
	landSound      []byte
	milestoneSound []byte
}

func NewGame() (*Game, error) {
	g := &Game{}

	g.bgImage = genBackgroundImage()
	g.obsImages = genObstacleImages()
	g.chickenGround = [2]*ebiten.Image{genChickenGroundImage(0), genChickenGroundImage(1)}
	g.chickenAir = genChickenAirImage()

	g.audioCtx = audio.NewContext(sampleRate)

	// Try to load the original jump.wav; fall back to procedural.
	g.jumpSound = g.loadWAV("assets/jump.wav")
	if g.jumpSound == nil {
		g.jumpSound = genJumpSound()
	}
	g.hitSound = genHitSound()
	g.landSound = genLandSound()
	g.milestoneSound = genMilestoneSound()

	g.particles = &Particles{}
	g.reset()
	return g, nil
}

func (g *Game) loadWAV(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("open %s: %v", path, err)
		return nil
	}
	decoded, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		log.Printf("decode %s: %v", path, err)
		return nil
	}
	pcm, err := io.ReadAll(decoded)
	if err != nil {
		log.Printf("read PCM %s: %v", path, err)
		return nil
	}
	return pcm
}

func (g *Game) playSound(pcm []byte) {
	if pcm == nil || g.audioCtx == nil {
		return
	}
	p := g.audioCtx.NewPlayerFromBytes(pcm)
	p.Play()
}

func (g *Game) reset() {
	g.player = newPlayer()
	g.obstacles = nil
	g.particles.pool = g.particles.pool[:0]
	g.ticks = 0
	g.score = 0
	g.speed = 3.0
	g.bgX = 0
	g.spawnTimer = 120
	g.speedUpTimer = 0
	g.flashTimer = 0
}

func (g *Game) flash(r, g2, b uint8, frames int) {
	g.flashR, g.flashG, g.flashB = r, g2, b
	g.flashTimer = frames
}

// Update advances the game by one tick.
func (g *Game) Update() error {
	if g.flashTimer > 0 {
		g.flashTimer--
	}

	switch g.state {
	case StateTitle:
		g.updateTitle()
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.reset()
			g.state = StatePlaying
		}
	case StatePlaying:
		g.ticks++
		g.updatePlaying()
	case StateGameOver:
		g.particles.Update()
		if g.shakeTicksVal > 0 {
			g.shakeTicksVal--
			g.shakeX = (rand.Float64()*2 - 1) * 5
			g.shakeY = (rand.Float64()*2 - 1) * 3
		} else {
			g.shakeX, g.shakeY = 0, 0
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.reset()
			g.state = StatePlaying
		}
	}
	return nil
}

func (g *Game) updateTitle() {
	g.titleTick++
	// Background drifts slowly on title screen.
	g.bgX -= 1.0

	// Demo chicken does a little hop every ~3 seconds.
	if g.titleTick%180 == 0 {
		g.titleVY = jumpSpeed * 0.8
	}
	g.titleVY += gravity
	g.titleY += g.titleVY
	if g.titleY >= 0 {
		g.titleY = 0
		g.titleVY = 0
	}
}

func (g *Game) updatePlaying() {
	// Ramp speed every 10 seconds.
	newSpeed := 3.0 + float64(g.ticks/600)*0.5
	if newSpeed > g.speed {
		g.speed = newSpeed
		g.speedUpTimer = 150
		g.playSound(g.milestoneSound)
	}

	// Background parallax scroll.
	g.bgX -= g.speed * 0.5

	// Player input + physics.
	jumped, landed := g.player.update()
	if jumped {
		g.playSound(g.jumpSound)
	}
	if landed {
		g.playSound(g.landSound)
		g.particles.EmitDust(g.player.x, g.player.y)
	}

	// Feather trail while airborne.
	if !g.player.onGround && g.ticks%4 == 0 {
		g.particles.EmitFeather(g.player.x, g.player.y)
	}

	// Screen shake decay.
	if g.shakeTicksVal > 0 {
		g.shakeTicksVal--
		g.shakeX = (rand.Float64()*2 - 1) * 5
		g.shakeY = (rand.Float64()*2 - 1) * 3
	} else {
		g.shakeX, g.shakeY = 0, 0
	}

	// Particles.
	g.particles.Update()

	// Spawn obstacles.
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		g.obstacles = append(g.obstacles, newObstacle(len(g.obsImages)))
		interval := 180 - int(g.speed)*15
		if interval < 60 {
			interval = 60
		}
		g.spawnTimer = interval + rand.Intn(60)
	}

	// Move obstacles, detect collision, remove off-screen.
	prevScore := g.score
	pBox := g.player.bounds()
	alive := g.obstacles[:0]
	for _, o := range g.obstacles {
		o.x -= g.speed
		if o.x+spriteSize < 0 {
			g.score++ // cleared one obstacle
			g.particles.EmitScorePop(g.player.x+spriteSize/2, g.player.y)
			continue
		}
		if pBox.overlaps(o.bounds()) {
			g.handleCollision()
			return
		}
		alive = append(alive, o)
	}
	g.obstacles = alive

	// Time-based score: +1 per second survived.
	if g.ticks%60 == 0 {
		g.score++
	}

	// Milestone every 10 points.
	if g.score/10 > prevScore/10 && g.score > 0 {
		g.playSound(g.milestoneSound)
	}

	if g.speedUpTimer > 0 {
		g.speedUpTimer--
	}
}

func (g *Game) handleCollision() {
	if g.score > g.highScore {
		g.highScore = g.score
	}
	g.particles.EmitExplosion(g.player.x, g.player.y)
	g.playSound(g.hitSound)
	g.shakeTicksVal = 22
	g.flash(220, 30, 30, 35)
	g.state = StateGameOver
}

// Draw renders the current frame.
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateTitle:
		g.drawBG(screen, 0, 0)
		// Demo chicken: centred, doing the auto-hop.
		titleImg := g.chickenGround[0]
		if g.titleY < -2 {
			titleImg = g.chickenAir
		}
		cx := float64(screenWidth/2 - spriteSize/2)
		cy := float64(groundY-spriteSize) + g.titleY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-50, -50)
		op.GeoM.Scale(-0.5, 0.5) // flip to face right
		op.GeoM.Translate(cx+float64(spriteSize)/2, cy+float64(spriteSize)/2)
		screen.DrawImage(titleImg, op)
		g.drawTitleUI(screen)

	case StatePlaying:
		g.drawScene(screen)

	case StateGameOver:
		g.drawScene(screen)
		g.drawOverlay(screen, color.RGBA{0, 0, 0, 120})
		g.drawPanel(screen, screenWidth/2-160, screenHeight/2-50, 320, 110)
		ebitenutil.DebugPrintAt(screen, "GAME  OVER", screenWidth/2-38, screenHeight/2-36)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), screenWidth/2-34, screenHeight/2-14)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best:  %d", g.highScore), screenWidth/2-34, screenHeight/2+2)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to retry", screenWidth/2-76, screenHeight/2+24)
	}

	// Flash overlay on top of everything.
	if g.flashTimer > 0 {
		alpha := uint8(float64(g.flashTimer) / 35 * 110)
		g.drawOverlay(screen, color.RGBA{g.flashR, g.flashG, g.flashB, alpha})
	}
}

func (g *Game) drawTitleUI(screen *ebiten.Image) {
	g.drawPanel(screen, screenWidth/2-170, screenHeight/2-48, 340, 110)
	ebitenutil.DebugPrintAt(screen, "C H I C K E N  G A M E", screenWidth/2-84, screenHeight/2-36)
	ebitenutil.DebugPrintAt(screen, "Arrow keys: move   Space/Up: jump", screenWidth/2-118, screenHeight/2-10)
	ebitenutil.DebugPrintAt(screen, "Double-jump is allowed!", screenWidth/2-88, screenHeight/2+8)
	ebitenutil.DebugPrintAt(screen, "Press SPACE to start", screenWidth/2-76, screenHeight/2+32)
}

func (g *Game) drawBG(screen *ebiten.Image, shakeX, shakeY float64) {
	bgW := float64(g.bgImage.Bounds().Dx())
	drawX := math.Mod(g.bgX, bgW)
	if drawX > 0 {
		drawX -= bgW
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(drawX+shakeX, shakeY)
	screen.DrawImage(g.bgImage, op)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(drawX+bgW+shakeX, shakeY)
	screen.DrawImage(g.bgImage, op2)
}

func (g *Game) drawScene(screen *ebiten.Image) {
	g.drawBG(screen, g.shakeX, g.shakeY)
	g.player.drawShadow(screen, g.shakeX, g.shakeY)
	chickenImg := g.chickenGround[g.player.walkFrame()]
	if !g.player.onGround {
		chickenImg = g.chickenAir
	}
	g.player.draw(screen, chickenImg, g.shakeX, g.shakeY)
	for _, o := range g.obstacles {
		o.draw(screen, g.obsImages[o.variant], g.shakeX, g.shakeY)
	}
	g.particles.Draw(screen, g.shakeX, g.shakeY)
	g.drawHUD(screen)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	g.drawPanel(screen, 4, 4, 130, 52)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best:  %d", g.highScore), 10, 24)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Speed: %.1f", g.speed), 10, 38)

	if g.speedUpTimer > 0 {
		alpha := uint8(math.Min(1, float64(g.speedUpTimer)/30) * 255)
		vector.DrawFilledRect(screen,
			float32(screenWidth/2-55), 52, 110, 18,
			color.RGBA{0, 0, 0, alpha / 2}, false)
		ebitenutil.DebugPrintAt(screen, "  SPEED  UP!", screenWidth/2-52, 56)
	}
}

// drawPanel draws a semi-transparent dark rounded panel.
func (g *Game) drawPanel(screen *ebiten.Image, x, y, w, h int) {
	vector.DrawFilledRect(screen,
		float32(x), float32(y), float32(w), float32(h),
		color.RGBA{0, 0, 0, 160}, false)
}

// drawOverlay fills the entire screen with a translucent colour.
func (g *Game) drawOverlay(screen *ebiten.Image, c color.RGBA) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, c, false)
}

// Layout returns the logical screen dimensions.
func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func genBackgroundImage() *ebiten.Image {
	w, h := screenWidth, screenHeight
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		var r, g, b uint8
		switch {
		case y < groundY:
			// Sky gradient: light blue at top, paler near horizon.
			t := float64(y) / float64(groundY)
			r = uint8(135 + t*50)
			g = uint8(195 + t*25)
			b = uint8(255 - t*20)
		case y < groundY+3:
			// Thin dirt strip at ground line.
			r, g, b = 145, 115, 65
		default:
			// Grass ground, darkening with depth.
			t := float64(y-groundY) / float64(h-groundY)
			r = uint8(70 - t*35)
			g = uint8(140 - t*55)
			b = uint8(40 - t*20)
		}
		for x := 0; x < w; x++ {
			rgba.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}

	img := ebiten.NewImageFromImage(rgba)

	// Soft clouds.
	cc := color.RGBA{255, 255, 255, 75}
	drawCloud := func(cx, cy float32) {
		vector.DrawFilledCircle(img, cx, cy, 30, cc, true)
		vector.DrawFilledCircle(img, cx+24, cy-6, 22, cc, true)
		vector.DrawFilledCircle(img, cx-20, cy+3, 20, cc, true)
		vector.DrawFilledCircle(img, cx+8, cy-12, 18, cc, true)
	}
	drawCloud(110, 55)
	drawCloud(360, 40)
	drawCloud(530, 70)

	// Grass tufts at ground line.
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 80; i++ {
		x := float32(rng.Intn(w))
		bladeH := float32(4 + rng.Intn(10))
		green := uint8(110 + rng.Intn(70))
		vector.DrawFilledRect(img, x, float32(groundY)-bladeH, 2, bladeH,
			color.RGBA{25, green, 20, 200}, false)
	}

	return img
}
