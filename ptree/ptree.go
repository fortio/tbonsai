// Package ptree implements procedural tree generation.
package ptree

import (
	"math"

	"fortio.org/rand"
	"fortio.org/terminal/ansipixels/tcolor"
)

type Canvas struct {
	Width, Height int
	Branches      []*Branch
	MonoColor     tcolor.RGBColor
	Rand          rand.Rand
}

type Point struct {
	X, Y float64
}

type Branch struct {
	Start      Point
	End        Point
	Angle      float64
	Length     float64
	StartWidth float64
	EndWidth   float64
	Rand       rand.Rand
}

func NewCanvas(rng rand.Rand, width, height int) *Canvas {
	c := &Canvas{
		Width:    width,
		Height:   height,
		Branches: make([]*Branch, 0, 20),
		Rand:     rng,
	}
	trunk := c.Trunk()
	c.Branches = append(c.Branches, trunk)
	// Generate branches recursively
	c.GenerateBranches(trunk, 4)
	return c
}

func (c *Canvas) Trunk() *Branch {
	// Create the trunk of the tree
	trunk := &Branch{
		Start:      Point{X: float64(c.Width)/2 - 0.5, Y: float64(c.Height)},
		Angle:      math.Pi/2 + .2*(c.Rand.Float64()-0.5),
		Length:     float64(c.Height) * 0.4,
		StartWidth: 6 + c.Rand.Float64(),
		EndWidth:   5 + c.Rand.Float64(),
		Rand:       c.Rand,
	}
	trunk.SetEnd()
	return trunk
}

func (b *Branch) SetEnd() {
	b.End = Point{
		X: b.Start.X + b.Length*math.Cos(b.Angle),
		Y: b.Start.Y - b.Length*math.Sin(b.Angle),
	}
}

type BranchType int

const (
	MidBranch BranchType = iota
	LeftBranch
	RightBranch
)

func (c *Canvas) GenerateBranches(cur *Branch, depth int) {
	if depth <= 0 {
		return
	}
	depth--
	if depth >= 1 {
		// 1 branch in middle
		c.AddBranch(cur.Add(MidBranch), depth)
	}
	// 2 at the end
	c.AddBranch(cur.Add(LeftBranch), depth)
	c.AddBranch(cur.Add(RightBranch), depth)
}

func (c *Canvas) AddBranch(b *Branch, depth int) {
	if b == nil {
		return
	}
	c.Branches = append(c.Branches, b)
	c.GenerateBranches(b, depth)
}

// Add a branch.
func (b *Branch) Add(t BranchType) *Branch {
	if b == nil || b.Length < 1 {
		// parent non existent or too small, skip
		return nil
	}
	// pick branch point
	dist := b.Length
	if t == MidBranch {
		dist = b.Length * (0.3 + 0.3*b.Rand.Float64())
	}
	branchPoint := Point{
		X: b.Start.X + dist*math.Cos(b.Angle),
		Y: b.Start.Y - dist*math.Sin(b.Angle),
	}
	// new branch parameters
	newLength := b.Length * (0.4 + 0.5*b.Rand.Float64())
	// calculate branch angle based on type
	var newAngle float64
	wiggle := (b.Rand.Float64() - 0.5) * (math.Pi / 20)
	switch t {
	case LeftBranch:
		newAngle = b.Angle - (math.Pi / 6)
	case RightBranch:
		newAngle = b.Angle + (math.Pi / 6)
	case MidBranch:
		// pick side randomly
		sign := 1.0
		if b.Rand.Float64() < 0.5 {
			sign = -1.0
		}
		newAngle = b.Angle + sign*(math.Pi/8)
	default:
		panic("unknown branch type")
	}
	newAngle += wiggle
	startWidth := b.EndWidth * (0.6 + 0.1*b.Rand.Float64())
	endWidth := startWidth * (0.7 + 0.1*b.Rand.Float64())
	newB := &Branch{
		Start:      branchPoint,
		Angle:      newAngle,
		Length:     newLength,
		StartWidth: startWidth,
		EndWidth:   endWidth,
		Rand:       b.Rand,
	}
	newB.SetEnd()
	return newB
}
