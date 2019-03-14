package table

import (
	"math"
	"strings"
	"fmt"
	"errors"
)

const maxFadeDegrees = 160
const startFadeDegrees = 20
const incrementFadeDegrees = 3

// Color describes a color.
type Color struct {
	r,g,b byte
}

// Direction describes a direction.
type Direction struct {
	start, end int
}

// AnimationPlayTable describes a table setup.
type AnimationPlayTable struct {
	frameBuffer *[]byte
	playerDirections *map[Direction]Color
	activeDirection *Direction
	currentFadeStep byte
	maxFadeSteps byte
}

// Colors is a set of predefined colors.
var Colors = map[string]Color {
	"red": Color{0xff, 0x00, 0x00},
	"green": Color{0x00, 0xff, 0x00},
	"blue": Color{0x00, 0x00, 0xff},
	"cyan": Color{0x00, 0xff, 0xff},
	"yellow": Color{0xff, 0xff, 0x00},
	"purple": Color{0xff, 0x00, 0xbf},
	"orange": Color{0xff, 0x80, 0x00},
	"white": Color{0xff, 0xff, 0xff},
}

// Directions is a set of predefined directions.
var Directions = map[string]Direction {
	"right": Direction{0, 40},
	"bottom": Direction{45, 115},
	"left": Direction{120, 156},
	"top": Direction{165, 236},
}

// NewAnimationPlayTable creates a new AnimationPlayTable from an encoded colormap.
func NewAnimationPlayTable(frameBuffer *[]byte) (*AnimationPlayTable) {
	newAnimation := new(AnimationPlayTable)
	newAnimation.frameBuffer = frameBuffer
	newAnimation.playerDirections = &map[Direction]Color{}
	newAnimation.currentFadeStep = startFadeDegrees
	newAnimation.activeDirection = nil
	return newAnimation
}

// SetFrameBuffer sets the frame buffer.
func (pt *AnimationPlayTable) SetFrameBuffer(frameBuffer *[]byte) {
	pt.frameBuffer = frameBuffer
}

// GetFrameBuffer gets the frame buffer.
func (pt *AnimationPlayTable) GetFrameBuffer() *[]byte {
	return pt.frameBuffer
}

// Step animates one increment.
func (pt *AnimationPlayTable) Step() {
	// if active player is set, overwrite the player's color with the gradient
	if pt.activeDirection != nil {
		// first, update buffer with original color
		pt.updateFrame()		
		if pt.currentFadeStep>=maxFadeDegrees {
			pt.currentFadeStep = startFadeDegrees
		} else {
			pt.currentFadeStep+=incrementFadeDegrees
		}
		currentFade := math.Sin(float64(pt.currentFadeStep)*math.Pi/180)
		for i:=(*pt.activeDirection).start*3; i<(*pt.activeDirection).end*3; i+=3 {
			if i<len(*pt.frameBuffer)-3 {
				fadedColorR := float64((*pt.frameBuffer)[i]) * currentFade
				fadedColorG := float64((*pt.frameBuffer)[i+1]) * currentFade
				fadedColorB := float64((*pt.frameBuffer)[i+2]) * currentFade
				(*pt.frameBuffer)[i] = byte(fadedColorR);
				(*pt.frameBuffer)[i+1] = byte(fadedColorG);
				(*pt.frameBuffer)[i+2] = byte(fadedColorB);
			} 
		}
	}
}

func (pt *AnimationPlayTable) updateFrame() error {
	if pt.frameBuffer == nil {
		return errors.New("no frame buffer declared")
	}
	for direction, color := range *pt.playerDirections {
		for i:=direction.start*3; i<direction.end*3; i+=3 {
			if i<len(*pt.frameBuffer)-3 {
				(*pt.frameBuffer)[i] = color.r;
				(*pt.frameBuffer)[i+1] = color.g;
				(*pt.frameBuffer)[i+2] = color.b;	
			} else {
				return errors.New("frame pixel out of bounds")
			}
		}
	}
	return nil
}

func (pt *AnimationPlayTable) checkDirection(input Direction) (Direction, error) {
	for _, direction := range Directions {
		if direction.start == input.start && direction.end == input.end {
			return direction, nil
		}
	}
	return input, errors.New("unknown direction")
}

// SetPlayerColor sets the color of a direction.
func (pt *AnimationPlayTable) SetPlayerColor(direction Direction, color Color) error {
	direction, err := pt.checkDirection(direction)
	if err != nil {
		return err
	}
	(*pt.playerDirections)[direction] = color
	// update the frame buffer
	pt.updateFrame()
	return nil
}

// SetPlayerColorFromString parses and sets the colors from a string-encoded mapping. 
// The mapping	needs to have the following format: s,e,r,g,b[-s,e,r,g,b]*.
func (pt *AnimationPlayTable) SetPlayerColorFromString(encoded string) error {
	var start int
	var end int
	var colorR, colorG, colorB byte
	mapSlice := strings.Split(encoded, "-")
	for _, encodedEntry := range mapSlice {
		fmt.Sscanf(encodedEntry, "%d,%d,%x,%x,%x", &start, &end, &colorR, &colorG, &colorB)
		direction, err := pt.checkDirection(Direction{start, end})
		if err != nil {
			return err
		}
		err = pt.SetPlayerColor(direction, Color{colorR, colorG, colorB})	
		if err != nil {
			return err
		}
	}
	return nil
}

// SetActiveDirection sets the active direction
func (pt *AnimationPlayTable) SetActiveDirection(direction Direction) error {
	direction, err := pt.checkDirection(direction)
	if err != nil {
		return err
	}
	pt.activeDirection = &direction
	return nil
}

// ActiveDirectionNext switches to the next active direction
func (pt *AnimationPlayTable) ActiveDirectionNext() error {
	if pt.activeDirection == nil {
		return errors.New("no active direction")
	}
	for name, direction := range Directions {
		if direction.start == pt.activeDirection.start && direction.end == pt.activeDirection.end {
			switch name {
			case "right":
				return pt.SetActiveDirection(Directions["bottom"])
			case "bottom":
				return pt.SetActiveDirection(Directions["left"])
			case "left":
				return pt.SetActiveDirection(Directions["top"])
			case "top":
				return pt.SetActiveDirection(Directions["right"])
			}
		}
	}
	return errors.New("active direction does not match known directions")
}

// ActiveDirectionOff turns active direction off
func (pt *AnimationPlayTable) ActiveDirectionOff() error {
	if pt.activeDirection == nil {
		return errors.New("no active direction")
	}
	pt.activeDirection = nil
	return nil
}

