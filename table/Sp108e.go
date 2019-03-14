package table

import (
	"io"
	"fmt"
	"time"
	"errors"
	"strconv"
	"net"
)

const cmdFrameStart = 0x38
const cmdFrameEnd = 0x83

const cmdCustomPreview = 0x24
const cmdBrightness = 0x2a

// Sp108e represents the connection to an SP108E.
type Sp108e struct {
	ip string
	port int
	connection net.Conn
	animationRunning bool
	currentAnimation Animation
 	animationBuffer []byte
}

// NewSp108e returns a new connection.
func NewSp108e(ip string, port int) (*Sp108e, error) {
	leds := new(Sp108e)
	leds.ip = ip
	leds.port = port
	leds.animationRunning = false
	leds.animationBuffer = make([]byte, 900, 900)
	var err error
	err = leds.CreateConnection()
	if err != nil {
		return nil, err
	}
	leds.startAnimationLoop()
	return leds, nil
}

// CreateConnection creates a connection.
func (leds *Sp108e) CreateConnection() error {
	var err error
	fmt.Println("establishing connection")
	leds.connection, err = net.Dial("tcp", leds.ip + ":" + strconv.Itoa(leds.port))
	if err != nil {
		return err
	}
	return nil
}

// CloseConnection creates a connection.
func (leds *Sp108e) CloseConnection() error {
	var err error
	fmt.Println("closing connection")
	err = leds.connection.Close()
	if err != nil {
		return err
	}
	return nil
}

// Reconnect recreates a connection. Animation setting is lost.
func (leds *Sp108e) Reconnect() error {
	if leds.animationRunning {
		leds.animationRunning = false
		leds.currentAnimation = nil
	}
	err := leds.CloseConnection()
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	err = leds.CreateConnection()
	if err != nil {
		return err
	}
	return nil
}

// IsConnectionEstablished returns true if a connection is established.
func (leds *Sp108e) IsConnectionEstablished() bool {
	if leds.connection == nil {
		return false
	}
	one := []byte{}
	// TODO deadline management
	//leds.connection.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, err := leds.connection.Read(one); err == io.EOF {
		fmt.Println("detected closed connection")
		leds.connection.Close()
		leds.CreateConnection()
	}
	return true
}

// GetFrameBuffer returns the frame buffer.
func (leds *Sp108e) GetFrameBuffer() *[]byte {
	return &leds.animationBuffer
}

func (leds *Sp108e) createCommandPacket(command byte, frame []byte) ([]byte, error) {
	if len(frame)!=3 {
		return nil, errors.New("command frame is not 3 bytes")
	}
	commandPacket := []byte{}
	commandPacket = append(commandPacket, cmdFrameStart)
	commandPacket = append(commandPacket, frame...)
	commandPacket = append(commandPacket, command)
	commandPacket = append(commandPacket, cmdFrameEnd)
	return commandPacket, nil
}

func (leds *Sp108e) sendCommand(command []byte, confirmExpected bool) error {
	// this needs to be called when sending data to the controller as
	// the timeout is established by the check method
	if !leds.IsConnectionEstablished() {
		fmt.Println("connection closed, not sending command")
	}
	leds.connection.Write(command)
	//time.Sleep(10 * time.Millisecond)
	if confirmExpected {
		tmp := make([]byte, 10)
		leds.connection.Read(tmp)
		if tmp[0] != 0x31 {
			fmt.Println("response not 0x31", tmp)
			return errors.New("response not 0x31")
		}
	}	
	return nil
}

func (leds *Sp108e) startAnimationLoop() {
	go func() {
		for {
			if (leds.animationRunning && leds.currentAnimation != nil) {
				(leds.currentAnimation).Step()
				err := leds.sendCommand(*(leds.currentAnimation.GetFrameBuffer()), true)
				if err != nil {
					fmt.Println("error rendering animation frame")
				}
			}
			// this is needed for the raspi
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// StartAnimation starts a new animation.
func (leds *Sp108e) StartAnimation(animation Animation) error {
	if leds.animationBuffer == nil {
		return errors.New("No framebuffer set for animation")
	}
	command, _ := leds.createCommandPacket(cmdCustomPreview, []byte {0x0, 0x0, 0x0})
	err := leds.sendCommand(command, true)
	if err != nil {
		return err
	} 
	leds.currentAnimation = animation
	leds.animationRunning = true
	return nil
}

// StopAnimation stops an animation.
func (leds *Sp108e) StopAnimation() error {
	if leds.animationRunning {
		leds.animationRunning = false
		time.Sleep(100 * time.Millisecond)
	}
	// this always succeeds, we ignore it if no animation is running
	return nil
}

// GetCurrentAnimation returns the current Animation.
func (leds *Sp108e) GetCurrentAnimation() Animation {
	return leds.currentAnimation
}

// SetBrightness sets the brightness.
func (leds *Sp108e) SetBrightness(value byte) error {
	command, _ := leds.createCommandPacket(cmdBrightness, []byte {byte(value), byte(value), byte(value)})
	if leds.animationRunning {
		leds.StopAnimation();
		err := leds.sendCommand(command, false)
		if err != nil {
			return err
		} 
		time.Sleep(100 * time.Millisecond)
		return leds.StartAnimation(leds.currentAnimation)
	}
	return leds.sendCommand(command, false)
}
