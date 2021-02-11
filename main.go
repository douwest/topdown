package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
	screenWidth          = 480
	screenHeight         = 320
	frameOX              = 0              //offset x
	frameOY              = 32             //offset y
	frameWidth           = 32             //width of char frame
	frameHeight          = 32             //height of char frame
	frameNum             =  8              //number of frames in animation cycle
	tileSize             = 16             // size wxh in px
	tileRows             = 16             // number of rows of tiles
	tileCols             = 16             // number of columns of tiles
	maxDashDistance      = 2.0 * tileSize //max dash distance in tiles
	frictionCoefficient  = 0.15           // reduce speed by this parameter every game-tick
	dashDelay            = 250            // dash delay in ms
	anticipationDelay    = (dashDelay / 6) * time.Millisecond
	walkSpeed            = 2.35
	walkAnimationSpeed   = 5
	speedIncrease        = 1
	sprintSpeed          = 3.6
	sprintAnimationSpeed = 3
	dashSpeedModifier    = 7
)

/*
GLOBAL VARIABLES -------------------------------------------------------------------------------------------------------
*/
var (
	maxSpeed       = walkSpeed // max movement speed
	characterImage *ebiten.Image
	groundImage    *ebiten.Image
	tileImage      *ebiten.Image
	tileColors     = [...]color.RGBA{
		{0x49, 0x63, 0x8c, 0xff},
		{0x5d, 0x74, 0x99, 0xff},
		{0x2b, 0x42, 0x66, 0xff},
		{0x42, 0x5a, 0x80, 0xff},
		{0x13, 0x34, 0x45, 0xff},
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
	location       Location
	speed          Speed
	mouseDirection float64
	attacking      bool
	attackIndex    int
	dashing        bool
	canDash        bool
	sprinting      bool
	collision      Collision
}

type Location struct {
	x              float64
	y              float64
}

type Speed struct {
	vertical        float64
	horizontal      float64
}

type Collision struct {
	verticalUp      bool
	verticalDown    bool
	horizontalLeft  bool
	horizontalRight bool
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
	var err error
	characterImage, _, err = ebitenutil.NewImageFromFile("runner.png")

	if err != nil {
		log.Fatal(err)
	}

	groundImage = ebiten.NewImage(tileSize*tileRows, tileSize*tileCols)
	tileImage = ebiten.NewImage(tileSize, tileSize)
	g := &Game{}
	g.character.canDash = true
	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

/*
LOGICAL GAME LOOP ------------------------------------------------------------------------------------------------------
*/
func (g *Game) Update() error {
	g.frameCount++

	g.checkHCollision()
	g.checkVCollision()
	g.checkSprinting()

	if (ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft)) && !g.character.collision.horizontalLeft {
		moveLeft(g)
	}
	if (ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight)) && !g.character.collision.horizontalRight {
		moveRight(g)
	}
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)) && !g.character.collision.verticalUp {
		moveUp(g)
	}
	if (ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)) && !g.character.collision.verticalDown {
		moveDown(g)
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.character.canDash && !g.character.hasCollision() {
		dash(g)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !g.character.attacking {
		attack(g)
	}
	if g.character.sprinting && g.character.hasCollision() && g.isPressingDirectionalKeys() {
		attack(g)
		fmt.Println("OUCH!")
	}

	g.updateCharacterPosition()
	g.reduceSpeedForFriction()
	g.updateCameraPosition()

	return nil
}

func (g *Game) checkHCollision() {
	if g.character.location.x-frameWidth/2 <= 0-screenWidth/2 {
		g.character.collision.horizontalLeft = true
	} else if g.character.location.x+frameWidth/2 >= tileCols*tileSize-screenWidth/2 {
		g.character.collision.horizontalRight = true
	} else {
		g.character.collision.horizontalLeft = false
		g.character.collision.horizontalRight = false
	}
	if g.character.collision.horizontalRight || g.character.collision.horizontalLeft {
		g.character.speed.horizontal = 0
	}

}

func (g *Game) checkVCollision() {
	if g.character.location.y-frameHeight/2 <= 0-screenHeight/2 {
		g.character.collision.verticalUp = true
	} else if g.character.location.y+frameHeight/2 >= tileRows*tileSize-screenHeight/2 {
		g.character.collision.verticalDown = true
	} else {
		g.character.collision.verticalUp = false
		g.character.collision.verticalDown = false
	}
	if g.character.collision.verticalUp || g.character.collision.verticalDown {
		g.character.speed.vertical = 0
	}
}

func (c Character) hasCollision() bool {
	return c.collision.horizontalLeft || c.collision.verticalUp || c.collision.verticalDown || c.collision.horizontalRight
}

func (c Character) isColliding() bool {
	return true
}

func (g *Game) checkSprinting()  {
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		maxSpeed = sprintSpeed
		g.character.sprinting = true
	} else {
		maxSpeed = walkSpeed
		g.character.sprinting = false
	}
}

func (g *Game) isPressingDirectionalKeys() bool {
	return ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)
}

/*
CHARACTER ACTIONS FUNCTIONS --------------------------------------------------------------------------------------------
*/
func moveLeft(g *Game) {
	if !g.character.dashing {
		g.character.speed.horizontal -= speedIncrease
		correctTooLowSpeed(g)
	}
}

func moveRight(g *Game) {
	if !g.character.dashing {
		g.character.speed.horizontal += speedIncrease
		correctTooHighSpeed(g)
	}
}

func moveDown(g *Game) {
	if !g.character.dashing {
		g.character.speed.vertical += speedIncrease
		correctTooHighSpeed(g)
	}

}

func moveUp(g *Game) {
	if !g.character.dashing {
		g.character.speed.vertical -= speedIncrease
		correctTooLowSpeed(g)
	}

}

func dash(g *Game) {
	g.character.canDash = false

	time.AfterFunc(anticipationDelay, func() {
		var x1, y1 = screenWidth / 2, screenHeight / 2
		var x2, y2 = ebiten.CursorPosition()
		var dashVector = Vector{float64(x1), float64(x2), float64(y1), float64(y2)}
		var dx, dy = dashVector.x2 - dashVector.x1, dashVector.y2 - dashVector.y1

		var dashVectorLength = math.Sqrt(dx*dx + dy*dy)
		var maxX = dx / dashVectorLength * maxDashDistance
		var maxY = dy / dashVectorLength * maxDashDistance

		g.character.dashing = true

		if dashVectorLength < maxDashDistance {
			g.character.speed.horizontal += dx / dashSpeedModifier
			g.character.speed.vertical += dy / dashSpeedModifier
		} else {
			g.character.speed.horizontal += maxX / dashSpeedModifier
			g.character.speed.vertical += maxY / dashSpeedModifier
		}

		startDashMovement(g)
	})
}

//TODO if g.character.attackIndex was not incremented for x seconds, reset to 0
func attack(g *Game) {
	g.character.attacking = true
	if g.character.attackIndex > 96 {
		g.character.attackIndex = 0
	}
	startComboTimer(g)
	startAttackTimer(g)
}

func startComboTimer(g *Game) *time.Timer {
	var currentIndex = g.character.attackIndex
	return time.AfterFunc(1000*time.Millisecond, func() {
		if g.character.attackIndex <= currentIndex {
			g.character.attackIndex = 0
		}
	})
}

func startAttackTimer(g *Game) *time.Timer {
	return time.AfterFunc(400*time.Millisecond, func() {
		g.character.attacking = false
		g.character.attackIndex += 32
	})
}

/*
DRAWING FUNCTIONS ------------------------------------------------------------------------------------------------------
*/

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, strconv.Itoa(int(ebiten.CurrentFPS())))
	g.drawGroundImage(screen, groundImage, tileImage)
	g.drawCharacterImage(screen, characterImage)
}

func (g *Game) drawCharacterImage(screen *ebiten.Image, characterImage *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(frameHeight)/2, -float64(frameHeight)/2) //translate image to center of bounding box
	opts.GeoM.Scale(getDirection(g), 1)                                   //scale x by relative mouse cursor position (-1 left of char, 1 right of char)
	opts.GeoM.Translate(screenWidth/2, screenHeight/2)                    //translate to center of screen

	var animationSpeed = walkAnimationSpeed

	if g.character.sprinting {
		animationSpeed = sprintAnimationSpeed
	}

	animationIndex := (g.frameCount / animationSpeed) % frameNum
	spriteX, spriteY := frameOX+animationIndex*frameWidth, frameOY

	if g.character.speed.horizontal == 0 && g.character.speed.vertical == 0 {
		spriteX, spriteY = 32, 0
	}

	if g.character.dashing {
		spriteX, spriteY = 128, 64
	}

	if g.character.attacking {
		spriteX, spriteY = g.character.attackIndex, 64
	}

	animationFrame := characterImage.SubImage(image.Rect(spriteX, spriteY, spriteX+frameWidth, spriteY+frameHeight)).(*ebiten.Image)
	screen.DrawImage(animationFrame, opts)
	// ebitenutil.DebugPrint(screen, "character (x, y) : "+strconv.FormatFloat(g.character.x, 'f', '1', 64)+", "+strconv.FormatFloat(g.character.y, 'f', '1', 64)+")")
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
	// do camera tracking logic here
	g.camera.x = g.character.location.x
	g.camera.y = g.character.location.y
}

func (g *Game) updateCharacterPosition() {
	g.character.location.x += g.character.speed.horizontal
	g.character.location.y += g.character.speed.vertical
}

func (g *Game) reduceSpeedForFriction() {
	if !g.character.dashing {
		if g.character.speed.horizontal > 0 {
			g.character.speed.horizontal -= frictionCoefficient
		} else if g.character.speed.horizontal < 0 {
			g.character.speed.horizontal += frictionCoefficient
		}
		if g.character.speed.vertical > 0 {
			g.character.speed.vertical -= frictionCoefficient
		} else if g.character.speed.vertical < 0 {
			g.character.speed.vertical += frictionCoefficient
		}

		g.checkIdle()
	}
}

func (g *Game) checkIdle() {
	// handles resetting speed back to 0 to prevent twitchy animations with floating point weirdness.
	const boundary = frictionCoefficient + 0.05
	if g.character.speed.horizontal < boundary && g.character.speed.horizontal > -boundary {
		g.character.speed.horizontal = 0
	}
	if g.character.speed.vertical < boundary && g.character.speed.vertical > -boundary {
		g.character.speed.vertical = 0
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
	if g.character.speed.horizontal < -(maxSpeed) {
		g.character.speed.horizontal = -(maxSpeed)
	}
	if g.character.speed.vertical < -(maxSpeed) {
		g.character.speed.vertical = -(maxSpeed)
	}
}

func correctTooHighSpeed(g *Game) {
	if g.character.speed.horizontal > maxSpeed {
		g.character.speed.horizontal = maxSpeed
	}
	if g.character.speed.vertical > maxSpeed {
		g.character.speed.vertical = maxSpeed
	}
}

func getDirection(g *Game) float64 {
	if g.character.speed.horizontal > 0 {
		return 1
	} else if g.character.speed.horizontal < 0 {
		return -1
	} else {
		return g.getCursorDirectionRelativeToCharacter()
	}
}

func startDashMovement(g *Game) *time.Timer {
	return time.AfterFunc(dashDelay*time.Millisecond, func() {
		g.character.dashing = false
		g.character.speed.horizontal = 0
		g.character.speed.vertical = 0
		time.AfterFunc(dashDelay*time.Millisecond, func() {
			g.character.canDash = true
		})
	})
}

func setupPlayground() {
	for i := 0; i < tileRows; i++ {
		for j := 0; j < tileCols; j++ {
			if i == 0 || i == tileRows-1 || j == 0 || j == tileCols-1 {
				playground[i][j] = 4
			} else {
				playground[i][j] = rand.Intn(3)
			}
		}
	}
}
