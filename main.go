package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
)

const mainMinutes, breakMinutes = 1, 1

var currentMinutes, currentSeconds = mainMinutes, 0
var quitGui chan bool
var paused bool

func main() {
	ticker := time.NewTicker(1 * time.Second)
	var flip, inBreak bool
	var wg sync.WaitGroup
	g := initGui()

	go func() {
		wg.Add(1)
		for {
			select {
			case <-quitGui:
				defer g.Close()
				defer ticker.Stop()
				defer wg.Done()
				return
			case <-ticker.C:
				if !paused {
					if currentMinutes == 0 && currentSeconds == 0 {
						flip = true
					} else {
						if currentSeconds <= 0 {
							currentSeconds = 59
							currentMinutes -= 1
						} else {
							currentSeconds -= 1
						}
					}
					if flip {
						if inBreak {
							currentMinutes, currentSeconds = mainMinutes, 0
						} else {
							currentMinutes, currentSeconds = breakMinutes, 0
						}
						inBreak = !inBreak
						flip = false
					}
					g.Update(update)
				}
			}
		}
	}()

	g.Update(update)
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	g.Close()
}

func initGui() *gocui.Gui {
	g, err := gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		log.Panicln(err)
	}

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'p', gocui.ModNone, pause); err != nil {
		log.Panicln(err)
	}

	return g
}

func update(g *gocui.Gui) error {
	if v, err := g.View("pomodogo"); err == nil {
		if gocui.IsUnknownView(err) {
			log.Panic(err)
			return err
		}
		v.Clear()
		fmt.Fprintln(v, fmt.Sprintf("%02d:%02d", currentMinutes, currentSeconds))
	} else {
		log.Panic("Something happened")
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("pomodogo", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2, 0); err != nil {
		if !gocui.IsUnknownView(err) {
			return err
		}
		fmt.Fprintln(v, "Oof")
		if _, err := g.SetCurrentView("pomodogo"); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func pause(g *gocui.Gui, v *gocui.View) error {
	paused = !paused
	return nil
}
