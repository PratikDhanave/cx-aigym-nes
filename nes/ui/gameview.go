package ui

import (
	term "github.com/nsf/termbox-go"
	"image"
	"os"

	"github.com/fogleman/nes/nes"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const padding = 0

type GameView struct {
	director *Director
	console  *nes.Console
	manager *Manager
	title    string
	hash     string
	texture  uint32
	record   bool
	frames   []image.Image
}

func NewGameView(director *Director, console *nes.Console, manager *Manager, title, hash string) View {
	var texture uint32
	if !director.glDisabled {
		texture = createTexture()
	}
	return &GameView{director, console, manager,title, hash, texture, false, nil}
}

func (view *GameView) Enter() {
	if !view.director.glDisabled {
		gl.ClearColor(0, 0, 0, 1)
		view.director.SetTitle(view.title)
		view.director.window.SetKeyCallback(view.onKey)
	}

	if !view.director.audioDisabled {
		view.console.SetAudioChannel(view.director.audio.channel)
		view.console.SetAudioSampleRate(view.director.audio.sampleRate)
	}


	// load state
	if err := view.console.LoadState(savePath(view.hash)); err == nil {
		return
	} else {
		view.console.Reset()
	}
	// load sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash)); err == nil {
			cartridge.SRAM = sram
		}
	}
}

func (view *GameView) Exit() {
	if !view.director.glDisabled {
		view.director.window.SetKeyCallback(nil)
		view.console.SetAudioChannel(nil)
		view.console.SetAudioSampleRate(0)
	}


	// save sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		writeSRAM(sramPath(view.hash), cartridge.SRAM)
	}
	// save state
	view.console.SaveState(savePath(view.hash))

	// exit
	os.Exit(0)
}


func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	window := view.director.window
	console := view.console
	if !view.director.glDisabled {
		if joystickReset(glfw.Joystick1) {
			view.director.ShowMenu()
		}
		if joystickReset(glfw.Joystick2) {
			view.director.ShowMenu()
		}
		if readKey(window, glfw.KeyEscape) {
			view.director.ShowMenu()
		}

	}

	updateControllers(view.director, console)
	console.StepSeconds(dt)

	if !view.director.glDisabled {
		gl.BindTexture(gl.TEXTURE_2D, view.texture)
		setTexture(console.Buffer())
		drawBuffer(view.director.window)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}

	if view.record {
		view.frames = append(view.frames, copyImage(console.Buffer()))
	}
}

func reset() {
	term.Sync()
}
//
//func (view *GameView) checkButtons() {
//
//loop:
//	for {
//		switch ev := term.PollEvent(); ev.Type {
//		case term.EventKey:
//			switch ev.Key {
//			case term.KeyEsc:
//				break loop
//			case term.KeyF1:
//				reset()
//				fmt.Println("F1 pressed")
//				break loop
//			case term.KeyF2:
//				reset()
//				fmt.Println("F2 pressed")
//				break loop
//
//			}
//		case term.EventError:
//			panic(ev.Err)
//		}
//	}
//}

func (view *GameView) onKey(window *glfw.Window,
	key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeySpace:
			screenshot(view.console.Buffer())
		case glfw.KeyR:
			view.console.Reset()
		case glfw.KeyTab:
			if view.record {
				view.record = false
				animation(view.frames)
				view.frames = nil
			} else {
				view.record = true
			}
		case glfw.Key1:
			// load state
			if err := view.console.LoadState(savePath(view.hash)); err == nil {
				return
			} else {
				view.console.Reset()
			}
		case glfw.Key2:
			// save state
			view.console.SaveState(savePath(view.hash))
		}


	}
}

func drawBuffer(window *glfw.Window) {
	w, h := window.GetFramebufferSize()
	s1 := float32(w) / 256
	s2 := float32(h) / 240
	f := float32(1 - padding)
	var x, y float32
	if s1 >= s2 {
		x = f * s2 / s1
		y = f
	} else {
		x = f
		y = f * s1 / s2
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(-x, -y)
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x, -y)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x, y)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(-x, y)
	gl.End()
}

func updateControllers(director *Director, console *nes.Console) {
	turbo := console.PPU.Frame%6 < 3

	var j1, j2, k1 [8]bool

	if director.glDisabled || director.randomKeys {
		k1 = readRandomKeys()
	} else {
		k1 =  readKeys(director.window, turbo)
	}


	if !director.glDisabled {
		j1 = readJoystick(glfw.Joystick1, turbo)
		j2 = readJoystick(glfw.Joystick2, turbo)
	}


	console.SetButtons1(combineButtons(k1, j1))
	console.SetButtons2(j2)
}
