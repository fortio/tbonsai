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
	c.GenerateBranches(trunk, 3, 2)
	return c
}

func (c *Canvas) Trunk() *Branch {
	// Create the trunk of the tree
	trunk := &Branch{
		Start:     Point{X: float64(c.Width)/2 - 0.5, Y: float64(c.Height)},
		Angle:     math.Pi/2 + .1*(rand.Float64()-0.5), //nolint:gosec // not crypto.
		Length:    float64(c.Height) * 0.45,
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

func (c *Canvas) GenerateBranches(cur *Branch, numBranches, depth int) {
	if depth <= 0 {
		return
	}
	for range numBranches {
		nb := cur.Add()
		c.Branches = append(c.Branches, nb)
		c.GenerateBranches(nb, numBranches, depth-1)
	}
}

// Add a branch.
func (b *Branch) Add() *Branch {
	// pick branch point
	dist := b.Length * (0.4 + 0.6*rand.Float64()) //nolint:gosec // not crypto.
	branchPoint := Point{
		X: b.Start.X + dist*math.Cos(b.Angle),
		Y: b.Start.Y - dist*math.Sin(b.Angle),
	}
	// new branch parameters
	newLength := b.Length * (0.4 + 0.5*rand.Float64()) //nolint:gosec // not crypto.
	newThickness := b.Thickness * 0.7
	newAngle := b.Angle + (rand.Float64()*0.5+0.2)*(1-2*rand.Float64()) //nolint:gosec // not crypto.
	newB := &Branch{
		Start:     branchPoint,
		Angle:     newAngle,
		Length:    newLength,
		Thickness: newThickness,
	}
	newB.SetEnd()
	return newB
}
