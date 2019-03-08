package main

import (
	"github.com/gobuffalo/packr"
	"net"
	"net/http"
	"encoding/json"
	"time"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const restPort = 8080

const cmdFrameStart = 0x38
const cmdFrameEnd = 0x83

const cmdCustomPreview = 0x24
const cmdBrightness = 0x2a

var colors = map[string]string {
	"red": "ff,00,00",
	"green": "00,ff,00",
	"blue": "00,00,ff",
	"cyan": "00,ff,ff",
	"yellow": "ff,ff,00",
	"purple": "ff,00,bf",
	"orange": "ff,80,00",
	"white": "ff,ff,ff",
}
var directions = map[string]string {
	"right": "0,40,",
	"bottom": "45,115,",
	"left": "120,156,",
	"top": "165,236,",
}

var conn net.Conn
var frame []byte
var cmdCustomPreviewRunning = false

func createCommandPacket(command byte, frame []byte) ([]byte, error) {
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

func sendCommand(connection net.Conn, command []byte, confirmExpected bool) error {
	connection.Write(command)
	time.Sleep(10 * time.Millisecond)
	if confirmExpected {
		tmp := make([]byte, 10)
		connection.Read(tmp)
		if tmp[0] != 0x31 {
			return errors.New("response not 0x31")
		}
	}	
	return nil
}

func parseColormaps(mapdefs []string, frame *[]byte) {
	for _, mapEntry := range mapdefs {
		fmt.Printf("parsing colormap '%s'\n", mapEntry)
		var start int
		var end int
		var colorR, colorG, colorB byte
		fmt.Sscanf(mapEntry, "%d,%d,%x,%x,%x", &start, &end, &colorR, &colorG, &colorB)
		fmt.Printf("setting leds %d to %d with color %x/%x/%x\n", start, end, colorR, colorG, colorB)
		for i:=start*3; i<end*3; i+=3 {
			if i<len(*frame)-3 {
				(*frame)[i] = colorR;
				(*frame)[i+1] = colorG;
				(*frame)[i+2] = colorB;	
			}
		}	
	}
}

func setBrightness(conn net.Conn, brightness int) {
	fmt.Println("setting brightness to", brightness)
	if cmdCustomPreviewRunning {
		cmdCustomPreviewRunning = false
		time.Sleep(50 * time.Millisecond)
		command, _ := createCommandPacket(cmdBrightness, []byte {byte(brightness), byte(brightness), byte(brightness)})
		sendCommand(conn, command, false)
		time.Sleep(50 * time.Millisecond)
		cmdCustomPreviewRunning = true
		return
	}
	command, _ := createCommandPacket(cmdBrightness, []byte {byte(brightness), byte(brightness), byte(brightness)})
	sendCommand(conn, command, false)	
}

func startCustomPreview(conn net.Conn) {
	command, _ := createCommandPacket(cmdCustomPreview, []byte {0x0, 0x0, 0x0})
	sendCommand(conn, command, true)
}

func runCustomPreview(conn net.Conn) {
	for {
		if (cmdCustomPreviewRunning) {
			err := sendCommand(conn, frame, true)
			if err != nil {
				fmt.Println("error: ", err)
			}
		}
		// this is needed for the raspi
		time.Sleep(10 * time.Millisecond)
	}	
}

func handleSuccess(w *http.ResponseWriter, result interface{}) {
	writer := *w
	marshalled, err := json.Marshal(result)
	if err != nil {
		handleError(w, 500, "Internal Server Error", "Error marshalling response JSON", err)
		return
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.Header().Add("Access-Control-Allow-Origin", "*")
	writer.WriteHeader(200)
	writer.Write(marshalled)
}

func handleError(w *http.ResponseWriter, code int, responseText string, logMessage string, err error) {
	errorMessage := ""
	writer := *w
	if err != nil {
		errorMessage = err.Error()
	}
	fmt.Println(logMessage, errorMessage)
	writer.WriteHeader(code)
	writer.Write([]byte(responseText))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("incoming request:", r.URL)
	switch r.Method {
	case http.MethodGet:
		doRequest(w, r)
		break
	default:
		handleError(&w, 405, "Method not allowed", "Method not allowed", nil)
		break
	}
}

func doRequest(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()
	command, ok := keys["command"]
	if !ok || len(command) != 1 {
		handleError(&w, 500, "command not given", "command not given", nil)
		return;
	}
	switch command[0] {
	case "brightness":
		value, ok := keys["value"]
		if !ok || len(value) != 1 {
			handleError(&w, 500, "value not given", "value not given", nil)
			return;
		}
		intValue, err := strconv.Atoi(value[0])
		if err != nil {
			handleError(&w, 500, "invalid value given", "invalid value given", nil)
			return;
		}
		setBrightness(conn, intValue)
		handleSuccess(&w, "success")
		break
	case "startcolormap":
		colormap, ok := keys["map"]
		if !ok || len(colormap) != 1 {
			handleError(&w, 500, "colormap not given", "colormap not given", nil)
			return;
		}
		brightness, ok := keys["brightness"]
		if !ok || len(brightness) != 1 {
			handleError(&w, 500, "brightness not given", "brightness not given", nil)
			return;
		}
		intBrightness, err := strconv.Atoi(brightness[0])
		if err != nil {
			handleError(&w, 500, "invalid brightness given", "invalid brightness given", nil)
			return;
		}
		cmdCustomPreviewRunning = false
		time.Sleep(50 * time.Millisecond)
		setBrightness(conn, intBrightness)
		mapSlice := strings.Split(colormap[0], "-")
		parseColormaps(mapSlice, &frame)
		startCustomPreview(conn)
		cmdCustomPreviewRunning = true
		handleSuccess(&w, "success")
		break
	case "stopcolormap":
		cmdCustomPreviewRunning = false
		handleSuccess(&w, "success")
		break
	case "tablecolors":
		l, ok := keys["left"]
		if !ok || len(l) != 1 {
			handleError(&w, 500, "left not given", "left not given", nil)
			return;
		}
		r, ok := keys["right"]
		if !ok || len(r) != 1 {
			handleError(&w, 500, "right not given", "right not given", nil)
			return;
		}
		t, ok := keys["top"]
		if !ok || len(t) != 1 {
			handleError(&w, 500, "top not given", "top not given", nil)
			return;
		}
		b, ok := keys["bottom"]
		if !ok || len(b) != 1 {
			handleError(&w, 500, "bottom not given", "bottom not given", nil)
			return;
		}
		brightness, ok := keys["brightness"]
		if !ok || len(brightness) != 1 {
			handleError(&w, 500, "brightness not given", "brightness not given", nil)
			return;
		}
		intBrightness, err := strconv.Atoi(brightness[0])
		if err != nil {
			handleError(&w, 500, "invalid brightness given", "invalid brightness given", nil)
			return;
		}
		cmdCustomPreviewRunning = false
		time.Sleep(50 * time.Millisecond)
		setBrightness(conn, intBrightness)
		var mapSlice []string
		mapSlice = append(mapSlice, directions["right"] + colors[r[0]])
		mapSlice = append(mapSlice, directions["bottom"] + colors[b[0]])
		mapSlice = append(mapSlice, directions["left"] + colors[l[0]])
		mapSlice = append(mapSlice, directions["top"] + colors[t[0]])
		parseColormaps(mapSlice, &frame)
		startCustomPreview(conn)
		cmdCustomPreviewRunning = true
		handleSuccess(&w, "success")
		break
	default:
		handleError(&w, 405, "unknown command", "unknown command", nil)
		break
	}
}

func main() {
	serverPtr := flag.Bool("server", false, "start rest server")
	hostPtr := flag.String("host", "192.168.178.83", "controller host")
	portPtr := flag.Int("port", 8189, "port number")
	brightnessPtr := flag.Int("brightness", -1, "brightness value")
	colormapPtr := flag.Bool("colormap", false, "colormap definition given")
	colorRightPtr := flag.String("right", "", "color right")
	colorLeftPtr := flag.String("left", "", "color left")
	colorTopPtr := flag.String("top", "", "color top")
	colorBottomPtr := flag.String("bottom", "", "color bottom")

	flag.Parse()

	fmt.Println("Boardgame Table Control")
	fmt.Println("using sp108e host:", *hostPtr)
	fmt.Println("using sp108e port:", *portPtr)

	// connect to the sp108e
	conn, _ = net.Dial("tcp", *hostPtr + ":" + strconv.Itoa(*portPtr))

	// prepare the frame buffer
	frame = make([]byte, 900, 900)

	if *serverPtr {
		// server mode, start rest service
		fmt.Printf("starting rest service on port %d, terminate with ctrl-c\n", restPort)
		// start frame animation thread
		go runCustomPreview(conn)
		// setup web service
		staticResources := packr.NewBox("./static")
	  http.Handle("/", http.FileServer(staticResources))	
		http.HandleFunc("/api", handleRequest)
		var err = http.ListenAndServe(":"+strconv.Itoa(restPort), nil)	
		if err != nil {
			fmt.Println("server failed starting:", err)
		}
	} else {
		// cli mode, parse the available cli params
		if *brightnessPtr!=-1 {
			setBrightness(conn, *brightnessPtr)
			time.Sleep(100 * time.Millisecond)
		}	
		// directions OR colormap
		if *colorRightPtr != "" && *colorLeftPtr != "" && *colorTopPtr != "" && *colorBottomPtr != "" {
			var mapSlice []string
			mapSlice = append(mapSlice, directions["right"] + colors[*colorRightPtr])
			mapSlice = append(mapSlice, directions["bottom"] + colors[*colorBottomPtr])
			mapSlice = append(mapSlice, directions["left"] + colors[*colorLeftPtr])
			mapSlice = append(mapSlice, directions["top"] + colors[*colorTopPtr])
			parseColormaps(mapSlice, &frame)
			fmt.Println("directional colors given, starting display loop, terminate with ctrl-c")
			startCustomPreview(conn)
			cmdCustomPreviewRunning = true	
			runCustomPreview(conn)
		} else if *colormapPtr {
			tailArgs := flag.Args()
			if len(tailArgs)<4 {
				tailArgs = append(tailArgs, directions["right"] + "ff,ff,ff")
				tailArgs = append(tailArgs, directions["bottom"] + "ff,00,00")
				tailArgs = append(tailArgs, directions["left"] + "00,00,ff")
				tailArgs = append(tailArgs, directions["top"] + "00,ff,00")
			}
			parseColormaps(tailArgs, &frame)
			fmt.Println("colormap given, starting display loop, terminate with ctrl-c")
			startCustomPreview(conn)
			cmdCustomPreviewRunning = true
			runCustomPreview(conn)
		}	
	}
}