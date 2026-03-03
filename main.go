package main

import (
	"fmt"
	"math"
	"math/rand"
	"path"
	"time"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/sdlcanvas"
)

const (
	segmentCount = 30
	segmentLen   = 16
	maxThickness = 14
	baseSpeed    = 2.0
)

type segment struct {
	x, y  float64
	angle float64
}

type Brain struct {
	heading float64
	speed   float64
}

type Limb struct {
	anchorIndex int
	phase       float64
	footX       float64
	footY       float64
}

var (
	segments []segment
	brain    Brain
	limbs    []Limb
	cv       *canvas.Canvas
	wnd      *sdlcanvas.Window
	timeStep float64
)

func initCreature(w, h int) {

	segments = make([]segment, segmentCount)

	startX := float64(w / 2)
	startY := float64(h / 2)

	for i := 0; i < segmentCount; i++ {
		segments[i] = segment{
			x: startX - float64(i)*segmentLen,
			y: startY,
		}
	}

	brain = Brain{
		heading: rand.Float64() * math.Pi * 2,
		speed:   baseSpeed,
	}

	// 4 legs
	limbs = []Limb{
		{anchorIndex: 5, phase: 0},
		{anchorIndex: 5, phase: math.Pi},
		{anchorIndex: 15, phase: math.Pi},
		{anchorIndex: 15, phase: 0},
	}
}

func updateBrain(w, h int) {

	head := &segments[0]

	// Random wandering
	brain.heading += (rand.Float64() - 0.5) * 0.1

	// Soft wall avoidance
	margin := 80.0
	avoidForce := 0.05

	if head.x < margin {
		brain.heading += avoidForce
	}
	if head.x > float64(w)-margin {
		brain.heading -= avoidForce
	}
	if head.y < margin {
		brain.heading += avoidForce
	}
	if head.y > float64(h)-margin {
		brain.heading -= avoidForce
	}

	head.x += math.Cos(brain.heading) * brain.speed
	head.y += math.Sin(brain.heading) * brain.speed
}

func updateSpine() {

	for i := 1; i < segmentCount; i++ {

		prev := segments[i-1]
		curr := &segments[i]

		dx := prev.x - curr.x
		dy := prev.y - curr.y
		angle := math.Atan2(dy, dx)

		curr.angle = angle

		curr.x = prev.x - math.Cos(angle)*segmentLen
		curr.y = prev.y - math.Sin(angle)*segmentLen
	}
}

func updateLimbs() {

	timeStep += 0.08

	for i := range limbs {

		l := &limbs[i]
		l.phase += 0.08

		anchor := segments[l.anchorIndex]

		side := 1.0
		if i%2 == 0 {
			side = -1
		}

		offsetAngle := anchor.angle + math.Pi/2*side

		stepRadius := 20.0

		// Foot stepping arc
		l.footX = anchor.x + math.Cos(offsetAngle)*stepRadius +
			math.Cos(l.phase)*6

		l.footY = anchor.y + math.Sin(offsetAngle)*stepRadius +
			math.Sin(l.phase)*6
	}
}

func drawSpine() {

	for i := 0; i < segmentCount; i++ {

		p := segments[i]

		t := 1.0 - float64(i)/float64(segmentCount)
		thickness := maxThickness * t

		perpX := math.Cos(p.angle + math.Pi/2)
		perpY := math.Sin(p.angle + math.Pi/2)

		count := int(thickness / 2)

		for j := -count; j <= count; j++ {

			offset := float64(j) * 2
			px := p.x + perpX*offset
			py := p.y + perpY*offset

			cv.SetFillStyle("rgb(60,200,120)")
			cv.FillRect(px, py, 2, 2)
		}
	}
}

func drawLimbs() {

	for _, l := range limbs {

		anchor := segments[l.anchorIndex]

		cv.SetStrokeStyle("#FFFFFF")
		cv.BeginPath()
		cv.MoveTo(anchor.x, anchor.y)
		cv.LineTo(l.footX, l.footY)
		cv.Stroke()

		cv.SetFillStyle("#FF5555")
		cv.FillRect(l.footX-3, l.footY-3, 6, 6)
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	var err error
	wnd, cv, err = sdlcanvas.CreateWindow(1200, 800, "Agent Driven Lizard")
	if err != nil {
		panic(err)
	}

	fontPath := path.Join("assets", "fonts", "montserrat.ttf")
	cv.SetFont(fontPath, 14)

	w, h := wnd.Size()
	initCreature(w, h)

	wnd.MainLoop(func() {

		start := time.Now()

		w, h := wnd.Size()

		cv.SetFillStyle("#101418")
		cv.FillRect(0, 0, float64(w), float64(h))

		updateBrain(w, h)
		updateSpine()
		updateLimbs()

		drawSpine()
		drawLimbs()

		dt := time.Since(start)
		if dt > 0 {
			cv.SetFillStyle("#FFFFFF")
			cv.FillText(fmt.Sprintf("FPS: %d", int(1.0/dt.Seconds())), 10, 20)
		}
	})
}
