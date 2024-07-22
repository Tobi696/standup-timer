package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/martinlindhe/notify"
)

const bufferTime = 2 * time.Minute

var standupInterval = 30 * time.Minute
var sitdownInterval = 30 * time.Minute

const (
	sitdown = "sitdown"
	standup = "standup"
)

var phaseTimer *PhaseTimer
var isActive = true

func main() {
	if len(os.Args) > 1 {
		sitdownIntervalMinutes, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal("Invalid sitdown interval")
		}
		sitdownInterval = time.Duration(sitdownIntervalMinutes) * time.Minute
	}
	if len(os.Args) > 2 {
		standupIntervalMinutes, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid standup interval")
		}
		standupInterval = time.Duration(standupIntervalMinutes) * time.Minute
	}

	go isActiveListener()

	for {
		if phaseTimer != nil && phaseTimer.State == sitdown {
			phaseTimer = &PhaseTimer{
				State:            standup,
				RemainingSeconds: int(standupInterval.Seconds()),
			}
		} else {
			phaseTimer = &PhaseTimer{
				State:            sitdown,
				RemainingSeconds: int(sitdownInterval.Seconds()),
			}
		}
		phaseTimer.Start()
	}
}

type PhaseTimer struct {
	State            string
	RemainingSeconds int
}

func (p *PhaseTimer) Start() {
	notify.Alert("StandUp", "StandUp", "Phase started: "+p.State, "")
	for {
		if isActive {
			p.RemainingSeconds--
		}
		time.Sleep(time.Second)
		if p.RemainingSeconds == 0 {
			return
		}
	}
}

func isActiveListener() {
	var lastMousePosX, lastMousePosY int
	var lastStateChangeTime time.Time
	for {
		mousePosX, mousePosY := robotgo.Location()
		if lastMousePosX != mousePosX || lastMousePosY != mousePosY {
			if !isActive {
				notify.Alert("StandUp", "StandUp", fmt.Sprintf("You are active again (%s %s remaining)", getRemainingTime(), phaseTimer.State), "")
			}
			isActive = true
			lastStateChangeTime = time.Now()
		} else {
			if time.Since(lastStateChangeTime) < bufferTime {
				continue
			}
			if isActive {
				notify.Alert("StandUp", "StandUp", "You are inactive", "")
			}
			isActive = false
			lastStateChangeTime = time.Now()
		}
		lastMousePosX, lastMousePosY = mousePosX, mousePosY
	}
}

// phaseTimer.RemainingSeconds in full minutes
func getRemainingTime() string {
	minutes := int(phaseTimer.RemainingSeconds / 60)
	return fmt.Sprintf("%d min", minutes)
}
