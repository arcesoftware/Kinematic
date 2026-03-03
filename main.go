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
	baseSpeed    = 2.2
)

type segment struct {
	x, y  float64
	angle float64
}

type Brain struct {
	heading float64
	speed   float64
	energy  float64
}

type Limb struct {
	anchorIndex int
	phase       float64
	footX       float64
	footY       float64
}

type Insect struct {
	x, y   float64
	vx, vy float64
}

var (
	segments []segment
	brain    Brain
	limbs    []Limb
	insects  []Insect

	cv  *canvas.Canvas
	wnd *sdlcanvas.Window

	timeStep float64
)

// ================= INIT =================

func initWorld(w, h int) {

	// Lizard spine
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
		energy:  120,
	}

	// 4 limbs
	limbs = []Limb{
		{anchorIndex: 6, phase: 0},
		{anchorIndex: 6, phase: math.Pi},
		{anchorIndex: 18, phase: math.Pi},
		{anchorIndex: 18, phase: 0},
	}

	// Insects
	for i := 0; i < 300; i++ {
		insects = append(insects, Insect{
			x: rand.Float64() * float64(w),
			y: rand.Float64() * float64(h),
		})
	}
}

// ================= INSECTS =================

func updateInsects(w, h int) {

	carryingCapacity := 800.0
	birthBase := 0.03

	N := float64(len(insects))

	for i := range insects {

		in := &insects[i]

		in.vx += (rand.Float64() - 0.5) * 0.5
		in.vy += (rand.Float64() - 0.5) * 0.5

		in.x += in.vx
		in.y += in.vy

		in.vx *= 0.94
		in.vy *= 0.94

		// wrap
		if in.x < 0 {
			in.x = float64(w)
		}
		if in.x > float64(w) {
			in.x = 0
		}
		if in.y < 0 {
			in.y = float64(h)
		}
		if in.y > float64(h) {
			in.y = 0
		}
	}

	// Logistic growth
	growth := birthBase * N * (1 - N/carryingCapacity)

	if growth > 0 {
		newBirths := int(growth * rand.Float64())
		for i := 0; i < newBirths; i++ {
			insects = append(insects, Insect{
				x: rand.Float64() * float64(w),
				y: rand.Float64() * float64(h),
			})
		}
	}
}

// ================= LIZARD BRAIN =================

func updateBrain(w, h int) {

	head := &segments[0]

	brain.energy -= 0.05

	// Find nearest insect
	var targetIndex = -1
	minDist := 99999.0

	for i := range insects {
		d := math.Hypot(insects[i].x-head.x, insects[i].y-head.y)
		if d < minDist {
			minDist = d
			targetIndex = i
		}
	}

	if targetIndex != -1 && minDist < 250 {
		tx := insects[targetIndex].x
		ty := insects[targetIndex].y
		brain.heading = math.Atan2(ty-head.y, tx-head.x)
	} else {
		brain.heading += (rand.Float64() - 0.5) * 0.08
	}

	head.x += math.Cos(brain.heading) * brain.speed
	head.y += math.Sin(brain.heading) * brain.speed

	// Eat
	for i := 0; i < len(insects); i++ {
		if math.Hypot(insects[i].x-head.x, insects[i].y-head.y) < 10 {
			insects = append(insects[:i], insects[i+1:]...)
			brain.energy += 8
			break
		}
	}
}

// ================= SPINE =================

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

// ================= LIMBS =================

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

		l.footX = anchor.x + math.Cos(offsetAngle)*stepRadius +
			math.Cos(l.phase)*6

		l.footY = anchor.y + math.Sin(offsetAngle)*stepRadius +
			math.Sin(l.phase)*6
	}
}

// ================= DRAW =================

func drawInsects() {
	for _, in := range insects {
		cv.SetFillStyle("rgb(255,200,50)")
		cv.FillRect(in.x-2, in.y-2, 4, 4)
	}
}

func drawLizard() {

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

			cv.SetFillStyle("rgb(70,200,120)")
			cv.FillRect(px, py, 2, 2)
		}
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	var err error
	wnd, cv, err = sdlcanvas.CreateWindow(1200, 800, "Single Predator Ecosystem")
	if err != nil {
		panic(err)
	}

	fontPath := path.Join("assets", "fonts", "montserrat.ttf")
	cv.SetFont(fontPath, 14)

	w, h := wnd.Size()
	initWorld(w, h)

	wnd.MainLoop(func() {

		w, h := wnd.Size()

		cv.SetFillStyle("#101418")
		cv.FillRect(0, 0, float64(w), float64(h))

		updateInsects(w, h)
		updateBrain(w, h)
		updateSpine()
		updateLimbs()

		drawInsects()
		drawLizard()

		cv.SetFillStyle("#FFFFFF")
		cv.FillText(fmt.Sprintf("Insects: %d  Energy: %.1f",
			len(insects), brain.energy), 20, 20)
	})
}
