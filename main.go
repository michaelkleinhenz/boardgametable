package main

import (
	"time"
	"flag"
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"

	"github.com/gobuffalo/packr"

	table "boardgametable/table"
)

const restPort = 8080

var sp108e *table.Sp108e

func handleSuccess(w *http.ResponseWriter, result interface{}) {
	writer := *w
	marshalled, err := json.Marshal(result)
	if err != nil {
		handleError(w, 500, "Internal Server Error:", "Error marshalling response JSON:", err)
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
		err = sp108e.SetBrightness(byte(intValue))
		if err != nil {
			handleError(&w, 500, "error setting brightness:", "error setting brightness:", err)
			return;
		}
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
		err = sp108e.StopAnimation()
		if err != nil {
			handleError(&w, 500, "error stopping animation:", "error stopping animation:", err)
			return;
		}
		err = sp108e.SetBrightness(byte(intBrightness))
		if err != nil {
			handleError(&w, 500, "error setting brightness:", "error setting brightness:", err)
			return;
		}
		animation := table.NewAnimationPlayTable(sp108e.GetFrameBuffer())
		err = animation.SetPlayerColorFromString(colormap[0])
		if err != nil {
			handleError(&w, 500, "error setting up animation:", "error setting up animation:", err)
			return;
		}
		err = sp108e.StartAnimation(animation)
		if err != nil {
			handleError(&w, 500, "error starting animation:", "error starting animation:", err)
			return;
		}
		handleSuccess(&w, "success")
		break
	case "stopcolormap":
		err := sp108e.StopAnimation()
		if err != nil {
			handleError(&w, 500, "error stopping animation:", "error stopping animation:", err)
			return;
		}
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
		err = sp108e.StopAnimation()
		if err != nil {
			handleError(&w, 500, "error stopping animation:", "error stopping animation:", err)
			return;
		}
		err = sp108e.SetBrightness(byte(intBrightness))
		if err != nil {
			handleError(&w, 500, "error setting brightness:", "error setting brightness:", err)
			return;
		}
		animation := table.NewAnimationPlayTable(sp108e.GetFrameBuffer())
		err = animation.SetPlayerColor(table.Directions["right"], table.Colors[r[0]])
		if err != nil {
			handleError(&w, 500, "error creating player color for right:", "error creating player color for right:", err)
			return;
		}
		animation.SetPlayerColor(table.Directions["bottom"], table.Colors[b[0]])
		if err != nil {
			handleError(&w, 500, "error creating player color for bottom:", "error creating player color for bottom:", err)
			return;
		}
		animation.SetPlayerColor(table.Directions["left"], table.Colors[l[0]])
		if err != nil {
			handleError(&w, 500, "error creating player color for left:", "error creating player color for left:", err)
			return;
		}
		animation.SetPlayerColor(table.Directions["top"], table.Colors[t[0]])
		if err != nil {
			handleError(&w, 500, "error creating player color for top:", "error creating player color for top:", err)
			return;
		}
		err = sp108e.StartAnimation(animation)
		if err != nil {
			handleError(&w, 500, "error starting animation:", "error starting animation:", err)
			return;
		}
		handleSuccess(&w, "success")
		break
	case "active":
		d, ok := keys["direction"]
		if !ok || len(d) != 1 {
			handleError(&w, 500, "direction not given", "direction not given", nil)
			return;
		}
		if d[0] != "left" && d[0] != "right" && d[0] != "top" && d[0] != "bottom" {
			handleError(&w, 500, "unknown direction", "unknown direction", nil)
			return;
		}
		currentAnimation := sp108e.GetCurrentAnimation()
		if currentAnimation == nil {
			handleError(&w, 500, "no current animation", "no current animation", nil)
			return;
		}
		currentPlayTableAnimation, ok := currentAnimation.(*table.AnimationPlayTable)
		if !ok {
			handleError(&w, 500, "current animation does not support active direction", "current animation does not support active direction", nil)
			return;
		}
		err := currentPlayTableAnimation.SetActiveDirection(table.Directions[d[0]])
		if err != nil {
			handleError(&w, 500, "error setting active direction:", "error setting active direction:", err)
			return;
		}
		handleSuccess(&w, "success")
		break
	case "nextactive":
		currentAnimation := sp108e.GetCurrentAnimation()
		if currentAnimation == nil {
			handleError(&w, 500, "no current animation", "no current animation", nil)
			return;
		}
		currentPlayTableAnimation, ok := currentAnimation.(*table.AnimationPlayTable)
		if !ok {
			handleError(&w, 500, "current animation does not support active direction", "current animation does not support active direction", nil)
			return;
		}
		err := currentPlayTableAnimation.ActiveDirectionNext()
		if err != nil {
			handleError(&w, 500, "error setting active direction:", "error setting active direction:", err)
			return;
		}
		handleSuccess(&w, "success")
		break
	case "activeoff":
		currentAnimation := sp108e.GetCurrentAnimation()
		if currentAnimation == nil {
			handleError(&w, 500, "no current animation", "no current animation", nil)
			return;
		}
		currentPlayTableAnimation, ok := currentAnimation.(*table.AnimationPlayTable)
		if !ok {
			handleError(&w, 500, "current animation does not support active direction", "current animation does not support active direction", nil)
			return;
		}
		err := currentPlayTableAnimation.ActiveDirectionOff()
		if err != nil {
			handleError(&w, 500, "error stopping active direction:", "error stopping active direction:", err)
			return;
		}
		handleSuccess(&w, "success")
		break
	case "reconnect":
		err := sp108e.Reconnect(true)
		if err != nil {
			handleError(&w, 500, "error reconnecting:", "error reconnecting:", err)
			return
		}
		handleSuccess(&w, "success")
		break

	default:
		handleError(&w, 405, "unknown command", "unknown command", nil)
		break
	}
}

func timerTask() {
	fmt.Println("performing scheduled reconnect..")
	err := sp108e.Reconnect(true)
		if err != nil {
			fmt.Println("error performing scheduled reconnect:", err)
			return
		}
}

func main() {
	serverPtr := flag.Bool("server", false, "start rest server")
	hostPtr := flag.String("host", "192.168.178.83", "controller host")
	portPtr := flag.Int("port", 8189, "port number")
	brightnessPtr := flag.Int("brightness", -1, "brightness value")
	colormapPtr := flag.String("colormap", "0,100,ff,00,00-101,200,00,ff,00", "colormap definition")
	colorRightPtr := flag.String("right", "", "color right")
	colorLeftPtr := flag.String("left", "", "color left")
	colorTopPtr := flag.String("top", "", "color top")
	colorBottomPtr := flag.String("bottom", "", "color bottom")
	reconnectIntervalPtr := flag.Int("reconnect", 300, "reconnect interval in seconds")

	flag.Parse()

	fmt.Println("Boardgame Table Control")
	fmt.Println("using sp108e host:", *hostPtr)
	fmt.Println("using sp108e port:", *portPtr)

	// connect to the sp108e
	var err error
	sp108e, err = table.NewSp108e(*hostPtr, *portPtr)
	if err != nil {
		fmt.Println("error connecting to sp108", err)
		return
	}

	if *serverPtr {
		// server mode, start rest service
		fmt.Printf("starting rest service on port %d, terminate with ctrl-c\n", restPort)
		// start timer that reconnects every 5 minutes
		ticker := time.NewTicker(time.Duration(*reconnectIntervalPtr) * time.Second)
		quit := make(chan struct{})
		// end timer by closing the quit channel: close(quit)
		go func() {
    	for {
       select {
        case <- ticker.C:
            timerTask()
        case <- quit:
            ticker.Stop()
            return
        }
    	}
 		}()
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
			sp108e.SetBrightness(byte(*brightnessPtr))
		}	
		// directions OR colormap
		if *colorRightPtr != "" && *colorLeftPtr != "" && *colorTopPtr != "" && *colorBottomPtr != "" {
			animation := table.NewAnimationPlayTable(sp108e.GetFrameBuffer())
			err := animation.SetPlayerColor(table.Directions["right"], table.Colors[*colorRightPtr])
			if err != nil {
				fmt.Println("error creating player color for right:", err)
				return;
			}
			animation.SetPlayerColor(table.Directions["bottom"], table.Colors[*colorBottomPtr])
			if err != nil {
				fmt.Println("error creating player color for bottom:", err)
				return;
			}
			animation.SetPlayerColor(table.Directions["left"], table.Colors[*colorLeftPtr])
			if err != nil {
				fmt.Println("error creating player color for left:", err)
				return;
			}
			animation.SetPlayerColor(table.Directions["top"], table.Colors[*colorTopPtr])
			if err != nil {
				fmt.Println("error creating player color for top:", err)
				return;
			}
			fmt.Println("directional colors given, starting display loop, terminate with ctrl-c")
			err = sp108e.StartAnimation(animation)
			if err != nil {
				fmt.Println("error starting animation:", err)
				return;
			}
		} else if *colormapPtr != "" {
			animation := new(table.AnimationPlayTable)
			err := animation.SetPlayerColorFromString(*colormapPtr)
			if err != nil {
				fmt.Println("error creating player color for right:", err)
				return;
			}
			fmt.Println("colormap given, starting display loop, terminate with ctrl-c")
			err = sp108e.StartAnimation(animation)
			if err != nil {
				fmt.Println("error starting animation:", err)
				return;
			}
		}	
	}
}