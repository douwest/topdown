package main

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/examples/resources/images"
	"image"
	"image/color"
	_ "image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

/*
GLOBAL CONSTANTS -------------------------------------------------------------------------------------------------------
*/
const (
	screenWidth         = 480
	screenHeight        = 320
	frameOX             = 0              //offset x
	frameOY             = 32             //offset y
	frameWidth          = 32             //width of char frame
	frameHeight         = 32             //height of char frame
	frameNum            = 8              //number of frames in animation cycle
	tileSize            = 32             // size wxh in px
	tileRows            = 16             // number of rows of tiles
	tileCols            = 16             // number of columns of tiles
	maxDashDistance     = 2.0 * tileSize //max dash distance in tiles
	frictionCoefficient = 0.25           // reduce speed by this parameter every game-tick
	maxSpeed            = 2.35           // max movement speed
	dashDelay           = 450            // dash delay in ms
)

/*
GLOBAL VARIABLES -------------------------------------------------------------------------------------------------------
*/
var (
	characterImage *ebiten.Image
	groundImage    *ebiten.Image
	tileImage      *ebiten.Image
	tileColors     = [...]color.RGBA{
		{0x49, 0x63, 0x8c, 0xff},
		{0x5d, 0x74, 0x99, 0xff},
		{0x2b, 0x42, 0x66, 0xff},
		{0x42, 0x5a, 0x80, 0xff},
	}
	playground [tileRows][tileCols]int
)

/*
TYPES ------------------------------------------------------------------------------------------------------------------
*/
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
	x              float64
	y              float64
	vSpeed         float64
	hSpeed         float64
	dashing        bool
	mouseDirection float64
}

type Vector struct {
	x1 float64
	x2 float64
	y1 float64
	y2 float64
}

/**
SETUP AND DRIVER FUNCTIONS ---------------------------------------------------------------------------------------------
*/

func main() {
	setupPlayground()
	g := setupGame()

	ebiten.SetWindowSize(4*screenWidth, 4*screenHeight)
	ebiten.SetWindowTitle("Fun time")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func setupGame() *Game {
	runnerImg, _, err := image.Decode(bytes.NewReader(images.Runner_png))

	if err != nil {
		log.Fatal(err)
	}
	characterImage = ebiten.NewImageFromImage(runnerImg)
	groundImage = ebiten.NewImage(tileSize*tileRows, tileSize*tileCols)
	tileImage = ebiten.NewImage(tileSize, tileSize)

	return &Game{}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

/*
LOGICAL GAME LOOP ------------------------------------------------------------------------------------------------------
*/
func (g *Game) Update() error {
	g.frameCount++

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

	g.updateCharacterPosition()
	g.reduceSpeedForFriction()
	g.updateCameraPosition()

	return nil
}

/*
CHARACTER ACTIONS FUNCTIONS --------------------------------------------------------------------------------------------
*/
func moveLeft(g *Game) {
	if !g.character.dashing {
		g.character.hSpeed -= 1
		correctTooLowSpeed(g)
	}
}

func moveRight(g *Game) {
	if !g.character.dashing {
		g.character.hSpeed += 1
		correctTooHighSpeed(g)
	}
}

func moveDown(g *Game) {
	if !g.character.dashing {
		g.character.vSpeed += 1
		correctTooHighSpeed(g)
	}

}

func moveUp(g *Game) {
	if !g.character.dashing {
		g.character.vSpeed -= 1
		correctTooLowSpeed(g)
	}

}

func dash(g *Game) {
	g.character.dashing = true

	time.AfterFunc((dashDelay/6)*time.Millisecond, func() {
		var x1, y1 = screenWidth / 2, screenHeight / 2
		var x2, y2 = ebiten.CursorPosition()
		var dashVector = Vector{float64(x1), float64(x2), float64(y1), float64(y2)}
		var dashVectorLength = math.Sqrt((dashVector.x2-dashVector.x1)*(dashVector.x2-dashVector.x1) + (dashVector.y2-dashVector.y1)*(dashVector.y2-dashVector.y1))
		var maxX = (dashVector.x2 - dashVector.x1) / dashVectorLength * maxDashDistance
		var maxY = (dashVector.y2 - dashVector.y1) / dashVectorLength * maxDashDistance

		if dashVectorLength < maxDashDistance {
			g.character.hSpeed += (dashVector.x2 - dashVector.x1) / 8
			g.character.vSpeed += (dashVector.y2 - dashVector.y1) / 8
		} else {
			g.character.hSpeed += maxX / 8
			g.character.vSpeed += maxY / 8
		}

		startDashTimer(g)
	})
}

//TODO this would be fun.
func attack(g *Game) { fmt.Println("Ouch! You attacked but there is no attack implemented!") }

/*
DRAWING FUNCTIONS ------------------------------------------------------------------------------------------------------
*/

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawGroundImage(screen, groundImage, tileImage)
	g.drawCharacterImage(screen, characterImage)
}

func (g *Game) drawCharacterImage(screen *ebiten.Image, characterImage *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(frameHeight)/2, -float64(frameHeight)/2) //translate image to center of bounding box
	opts.GeoM.Scale(getDirection(g), 1)                                   //scale x by -1 when moving left, 1 when right
	opts.GeoM.Translate(screenWidth/2, screenHeight/2)                    //translate to center of screen

	animationIndex := (g.frameCount / 5) % frameNum
	spriteX, spriteY := frameOX+animationIndex*frameWidth, frameOY

	if g.character.hSpeed == 0 && g.character.vSpeed == 0 {
		spriteX, spriteY = 32, 0
	}

	if g.character.dashing {
		spriteX, spriteY = 64, 64
	}

	animationFrame := characterImage.SubImage(image.Rect(spriteX, spriteY, spriteX+frameWidth, spriteY+frameHeight)).(*ebiten.Image)
	screen.DrawImage(animationFrame, opts)
	ebitenutil.DebugPrint(screen, "x0, y0, x1, y1: ("+
		strconv.Itoa(spriteX)+", "+ // 0 - 224 start x
		strconv.Itoa(spriteY)+", "+ // always 32 start y
		strconv.Itoa(spriteX+frameWidth)+", "+ // 32 - 256 end x
		strconv.Itoa(spriteY+frameHeight)+"), a: "+ // always 64 end y
		strconv.Itoa(animationIndex)+ // always 64 end y
		"  , FPS:"+strconv.FormatFloat(ebiten.CurrentFPS(), 'f', 1, 64))
}

func (g *Game) drawGroundImage(screen *ebiten.Image, ground *ebiten.Image, tile *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-g.camera.x, -g.camera.y)
	g.drawTileImages(tile, ground) //draw tiles here to keep reference to translate above (camera position)
	screen.DrawImage(ground, opts)
}

func (g *Game) drawTileImages(tile *ebiten.Image, ground *ebiten.Image) {
	for x := 0; x < tileRows; x++ {
		for y := 0; y < tileCols; y++ {
			tile.Fill(tileColors[playground[x][y]])
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((x)*tileSize), float64((y)*tileSize))
			ground.DrawImage(tile, op)
		}
	}
}

/*
UTILITY FUNCTIONS ------------------------------------------------------------------------------------------------------
*/

func (g *Game) updateCameraPosition() {
	// do fancy tracking things here
	g.camera.x = g.character.x
	g.camera.y = g.character.y
}

func (g *Game) updateCharacterPosition() {
	g.character.x += g.character.hSpeed
	g.character.y += g.character.vSpeed
}

func (g *Game) reduceSpeedForFriction() {
	if g.character.hSpeed > 0 {
		g.character.hSpeed -= frictionCoefficient
	} else if g.character.hSpeed < 0 {
		g.character.hSpeed += frictionCoefficient
	}
	if g.character.vSpeed > 0 {
		g.character.vSpeed -= frictionCoefficient
	} else if g.character.vSpeed < 0 {
		g.character.vSpeed += frictionCoefficient
	}

	g.checkIdle()
}

func (g *Game) checkIdle() {
	// handles resetting speed back to 0 to prevent twitchy animations with floating point weirdness.
	const boundary = frictionCoefficient + 0.05
	if g.character.hSpeed < boundary && g.character.hSpeed > -boundary {
		g.character.hSpeed = 0
	}
	if g.character.vSpeed < boundary && g.character.vSpeed > -boundary {
		g.character.vSpeed = 0
	}
}

func (g *Game) getCursorDirectionRelativeToCharacter() float64 {
	x0, _ := ebiten.CursorPosition()
	x1 := screenWidth / 2
	if x0 > x1 {
		return 1
	} else {
		return -1
	}
}

func correctTooLowSpeed(g *Game) {
	if g.character.hSpeed < -(maxSpeed) {
		g.character.hSpeed = -(maxSpeed)
	}
	if g.character.vSpeed < -(maxSpeed) {
		g.character.vSpeed = -(maxSpeed)
	}
}

func correctTooHighSpeed(g *Game) {
	if g.character.hSpeed > maxSpeed {
		g.character.hSpeed = maxSpeed
	}
	if g.character.vSpeed > maxSpeed {
		g.character.vSpeed = maxSpeed
	}
}

func getDirection(g *Game) float64 {
	if g.character.hSpeed > 0 {
		return 1
	} else if g.character.hSpeed < 0 {
		return -1
	} else {
		return g.getCursorDirectionRelativeToCharacter()
	}
}

func startDashTimer(g *Game) *time.Timer {
	return time.AfterFunc(dashDelay*time.Millisecond, func() {
		g.character.dashing = false
	})
}

func setupPlayground() {
	for i := 0; i < tileRows; i++ {
		for j := 0; j < tileCols; j++ {
			playground[i][j] = rand.Intn(3)
		}
	}
}
