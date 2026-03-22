package main

import (
	"bytes"
	"fmt"
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
)

const (
	screenWidth  = 640
	screenHeight = 480
	// groundY is the y-coordinate of the ground surface.
	groundY    = screenHeight - 111
	spriteSize = 50

	gravity   = 0.35
	jumpSpeed = -6.5

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
	score     int
	highScore int
	ticks     int
	speed     float64
	bgX       float64

	spawnTimer   int
	speedUpTimer int // frames left to display "SPEED UP!" banner

	// images
	bgImage      *ebiten.Image
	chickenImage *ebiten.Image
	obsImage     *ebiten.Image

	// audio
	audioCtx  *audio.Context
	jumpSound []byte // decoded PCM bytes

	// screen-shake effect
	shakeTicks int
	shakeX     float64
	shakeY     float64
}

func NewGame() (*Game, error) {
	g := &Game{}

	var err error
	g.bgImage, _, err = ebitenutil.NewImageFromFile("assets/bg.png")
	if err != nil {
		return nil, fmt.Errorf("load bg: %w", err)
	}
	g.chickenImage, _, err = ebitenutil.NewImageFromFile("assets/chicken.png")
	if err != nil {
		return nil, fmt.Errorf("load chicken: %w", err)
	}
	g.obsImage, _, err = ebitenutil.NewImageFromFile("assets/obstacle.png")
	if err != nil {
		return nil, fmt.Errorf("load obstacle: %w", err)
	}

	g.audioCtx = audio.NewContext(sampleRate)
	if wavData, readErr := os.ReadFile("assets/jump.wav"); readErr == nil {
		if decoded, decErr := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(wavData)); decErr == nil {
			if pcm, pcmErr := io.ReadAll(decoded); pcmErr == nil {
				g.jumpSound = pcm
			} else {
				log.Printf("read jump PCM: %v", pcmErr)
			}
		} else {
			log.Printf("decode jump.wav: %v", decErr)
		}
	} else {
		log.Printf("open jump.wav: %v", readErr)
	}

	g.reset()
	return g, nil
}

func (g *Game) reset() {
	g.player = newPlayer()
	g.obstacles = nil
	g.ticks = 0
	g.score = 0
	g.speed = 3.0
	g.bgX = 0
	g.spawnTimer = 120
	g.speedUpTimer = 0
	g.shakeTicks = 0
	g.shakeX, g.shakeY = 0, 0
}

func (g *Game) playJump() {
	if g.jumpSound == nil {
		return
	}
	p, err := g.audioCtx.NewPlayerFromBytes(g.jumpSound)
	if err != nil {
		return
	}
	p.Play()
}

// Update advances the game by one tick.
func (g *Game) Update() error {
	// Screen shake runs regardless of state so it completes after collision.
	if g.shakeTicks > 0 {
		g.shakeTicks--
		g.shakeX = (rand.Float64()*2 - 1) * 5
		g.shakeY = (rand.Float64()*2 - 1) * 3
	} else {
		g.shakeX, g.shakeY = 0, 0
	}

	switch g.state {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.reset()
			g.state = StatePlaying
		}
	case StatePlaying:
		g.ticks++
		g.updatePlaying()
	case StateGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.reset()
			g.state = StatePlaying
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	// Ramp up speed every 10 seconds; notify player.
	newSpeed := 3.0 + float64(g.ticks/600)*0.5
	if newSpeed > g.speed {
		g.speed = newSpeed
		g.speedUpTimer = 120
	}

	// Scroll background at half game speed for a parallax feel.
	g.bgX -= g.speed * 0.5

	// Player input + physics.
	if g.player.update() {
		g.playJump()
	}

	// Spawn obstacles.
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		g.obstacles = append(g.obstacles, newObstacle())
		interval := 180 - int(g.speed)*15
		if interval < 60 {
			interval = 60
		}
		g.spawnTimer = interval + rand.Intn(60)
	}

	// Move obstacles, detect collision, remove off-screen ones.
	pBox := g.player.bounds()
	alive := g.obstacles[:0]
	for _, o := range g.obstacles {
		o.x -= g.speed
		if o.x+spriteSize < 0 {
			g.score++ // reward clearing an obstacle
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

	if g.speedUpTimer > 0 {
		g.speedUpTimer--
	}
}

func (g *Game) handleCollision() {
	if g.score > g.highScore {
		g.highScore = g.score
	}
	g.shakeTicks = 20
	g.state = StateGameOver
}

// Draw renders the current frame.
func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateTitle:
		g.drawBG(screen)
		g.player.draw(screen, g.chickenImage, 0, 0)
		ebitenutil.DebugPrintAt(screen, "CHICKEN GAME", screenWidth/2-48, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Arrow keys to move  Space to jump", screenWidth/2-130, screenHeight/2)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to start", screenWidth/2-76, screenHeight/2+20)

	case StatePlaying:
		g.drawScene(screen)

	case StateGameOver:
		g.drawScene(screen)
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-36, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d   Best: %d", g.score, g.highScore), screenWidth/2-90, screenHeight/2)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to retry", screenWidth/2-76, screenHeight/2+20)
	}
}

func (g *Game) drawBG(screen *ebiten.Image) {
	bgW := float64(g.bgImage.Bounds().Dx())
	drawX := math.Mod(g.bgX, bgW)
	if drawX > 0 {
		drawX -= bgW
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(drawX+g.shakeX, g.shakeY)
	screen.DrawImage(g.bgImage, op)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(drawX+bgW+g.shakeX, g.shakeY)
	screen.DrawImage(g.bgImage, op2)
}

func (g *Game) drawScene(screen *ebiten.Image) {
	g.drawBG(screen)
	g.player.draw(screen, g.chickenImage, g.shakeX, g.shakeY)
	for _, o := range g.obstacles {
		o.draw(screen, g.obsImage, g.shakeX, g.shakeY)
	}

	// HUD
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.score), 8, 8)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best:  %d", g.highScore), 8, 22)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Speed: %.1f", g.speed), 8, 36)

	if g.speedUpTimer > 0 {
		ebitenutil.DebugPrintAt(screen, "SPEED UP!", screenWidth/2-34, 60)
	}
}

// Layout returns the logical screen dimensions.
func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}
