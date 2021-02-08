package main

import (
	"bytes"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/examples/resources/images"
	"image"
	_ "image/png"
	"log"
	"math"
	"time"
)

const (
	screenWidth  = 320
	screenHeight = 240

	frameOX                = 0
	frameOY                = 32
	frameWidth             = 32
	frameHeight            = 32
	frameNum               = 8
	gravityConstant        = 0.14
	horizontalAcceleration = 0.1
	maxSpeed               = 2
	maxDashDistance        = 80
)

var (
	characterImage *ebiten.Image
	groundImage    *ebiten.Image
)

type Game struct {
	frameCount int
	character  Character
	camera     Camera
}

type Camera struct {
	x float64
	y float64
}

type Character struct {
	x          float64
	y          float64
	vSpeed     float64
	hSpeed     float64
	vCollision bool
	dashing    bool
	hCollision bool
}

type Vector struct {
	x1 float64
	x2 float64
	y1 float64
	y2 float64
}

func (g *Game) Update() error {
	g.frameCount++

	friction(g)

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		moveLeft(g)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		moveRight(g)
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		moveUp(g)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		moveDown(g)
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !g.character.dashing {
		dash(g)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		attack(g)
	}

	g.camera.x += g.character.hSpeed
	g.camera.y += g.character.vSpeed

	return nil
}

func friction(g *Game) {
	g.character.hSpeed -= 0.1
	g.character.vSpeed -= 0.1
	if g.character.hSpeed < 0 {
		g.character.hSpeed = 0
	}
	if g.character.vSpeed < 0 {
		g.character.vSpeed = 0
	}
}

func dash(g *Game) {
	g.character.dashing = true

	var x2, y2 = ebiten.CursorPosition()
	var x1, y1 = g.camera.x, g.camera.y
	var vector = Vector{x1, float64(x2), y1, float64(y2)}
	var length = math.Sqrt(
		(vector.x2-vector.x1)*(vector.x2-vector.x1) +
			(vector.y2-vector.y1)*(vector.y2-vector.y1),
	)

	var maxX = (vector.x2 - vector.x1) / length * maxDashDistance
	var maxY = (vector.y2 - vector.y1) / length * maxDashDistance

	maxEuclideanDistance := math.Sqrt(maxDashDistance*maxDashDistance + maxDashDistance*maxDashDistance)

	if length < maxEuclideanDistance {
		g.camera.x = -(vector.x2 - vector.x1)
		g.camera.y = -(vector.y2 - vector.y1)
	} else {
		g.camera.x = -maxX
		g.camera.y = -maxY
	}

	time.AfterFunc(500*time.Millisecond, func() {
		g.character.dashing = false
	})

}

func attack(g *Game) {}

func moveLeft(g *Game) {
	g.character.hSpeed = -1
	g.camera.x += g.character.hSpeed
}

func moveRight(g *Game) {
	g.character.hSpeed = 1
	g.camera.x += g.character.hSpeed
}

func moveDown(g *Game) {
	g.character.vSpeed = 1
	g.camera.y += g.character.vSpeed
}

func moveUp(g *Game) {
	g.character.vSpeed = -1
	g.camera.y += g.character.vSpeed
}

func getDirection(character Character) float64 {
	if character.hSpeed > 0 {
		return 1
	} else {
		return -1
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawGroundImage(screen, groundImage)
	g.drawCharacterImage(screen, characterImage)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) drawGroundImage(screen *ebiten.Image, ground *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.camera.x, -g.camera.y)
	screen.DrawImage(ground, op)
}

func (g *Game) drawCharacterImage(screen *ebiten.Image, characterImage *ebiten.Image) {
	drawOptions := &ebiten.DrawImageOptions{}

	animationIndex := (g.frameCount / 5) % frameNum
	spriteX, spriteY := frameOX+animationIndex*frameWidth, frameOY

	drawOptions.GeoM.Translate(-float64(frameHeight)/2, -float64(frameHeight)/2) //translate image to center of bounding box
	drawOptions.GeoM.Scale(getDirection(g.character), 1) //scale x by -1 when moving left, 1 when right
	drawOptions.GeoM.Translate(screenWidth / 2, screenHeight / 2) //translate to center of screen

	screen.DrawImage(characterImage.SubImage(image.Rect(spriteX, spriteY, spriteX+frameWidth, spriteY+frameHeight)).(*ebiten.Image), drawOptions)
}

func main() {
	runnerImg, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	tileImg, _, err := image.Decode(bytes.NewReader(images.Tile_png))

	g := &Game{}
	g.character.vSpeed = gravityConstant

	if err != nil {
		log.Fatal(err)
	}
	characterImage = ebiten.NewImageFromImage(runnerImg)
	groundImage = ebiten.NewImageFromImage(tileImg)

	ebiten.SetWindowSize(4 * screenWidth, 4 * screenHeight)
	ebiten.SetWindowTitle("Fun time")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
