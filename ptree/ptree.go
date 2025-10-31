// Package ptree implements procedural tree generation.
package ptree

import (
	"math"
	"math/rand/v2"
)

type Canvas struct {
	Width, Height int
	Branches      []*Branch
}

type Point struct {
	X, Y float64
}

type Branch struct {
	Start     Point
	End       Point
	Angle     float64
	Length    float64
	Thickness float64
}

func NewCanvas(width, height int) *Canvas {
	c := &Canvas{
		Width:    width,
		Height:   height,
		Branches: make([]*Branch, 0, 20),
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
		Start:     Point{X: float64(c.Width)/2 - 0.5, Y: float64(c.Height)},
		Angle:     math.Pi/2 + .2*(rand.Float64()-0.5), //nolint:gosec // not crypto.
		Length:    float64(c.Height) * 0.4,
		Thickness: 1.0,
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
		dist = b.Length * (0.3 + 0.3*rand.Float64()) //nolint:gosec // not crypto.
	}
	branchPoint := Point{
		X: b.Start.X + dist*math.Cos(b.Angle),
		Y: b.Start.Y - dist*math.Sin(b.Angle),
	}
	// new branch parameters
	newLength := b.Length * (0.4 + 0.5*rand.Float64()) //nolint:gosec // not crypto.
	newThickness := b.Thickness * 0.7
	// calculate branch angle based on type
	var newAngle float64
	wiggle := (rand.Float64() - 0.5) * (math.Pi / 20) //nolint:gosec // not crypto.
	switch t {
	case LeftBranch:
		newAngle = b.Angle - (math.Pi / 6)
	case RightBranch:
		newAngle = b.Angle + (math.Pi / 6)
	case MidBranch:
		// pick side randomly
		sign := 1.0
		if rand.Float64() < 0.5 { //nolint:gosec // not crypto.
			sign = -1.0
		}
		newAngle = b.Angle + sign*(math.Pi/8)
	default:
		panic("unknown branch type")
	}
	newAngle += wiggle
	// newAngle := b.Angle + (rand.Float64()*0.5+0.2)*(1-2*rand.Float64()) //nolint:gosec // not crypto.
	newB := &Branch{
		Start:     branchPoint,
		Angle:     newAngle,
		Length:    newLength,
		Thickness: newThickness,
	}
	newB.SetEnd()
	return newB
}
