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
	IsTrunk    bool // If true, draw bottom flat
}

func NewCanvas(rng rand.Rand, width, height int) *Canvas {
	return NewCanvasWithOptions(rng, width, height, 4, 8.0, 40.0)
}

func NewCanvasWithOptions(rng rand.Rand, width, height, depth int, trunkWidth, trunkHeightPct float64) *Canvas {
	c := &Canvas{
		Width:    width,
		Height:   height,
		Branches: make([]*Branch, 0, 20),
		Rand:     rng,
	}
	trunk := c.Trunk(trunkWidth, trunkHeightPct)
	c.Branches = append(c.Branches, trunk)
	// Generate branches breadth-first
	c.GenerateBranchesBFS(trunk, depth)
	return c
}

func (c *Canvas) Trunk(trunkWidth, trunkHeightPct float64) *Branch {
	// Create the trunk of the tree
	trunk := &Branch{
		Start:      Point{X: float64(c.Width)/2 - 0.5, Y: float64(c.Height)},
		Angle:      math.Pi/2 + .2*(c.Rand.Float64()-0.5),
		Length:     float64(c.Height) * (trunkHeightPct / 100.0),
		StartWidth: trunkWidth + 0.2*trunkWidth*c.Rand.Float64(),
		EndWidth:   trunkWidth*0.75 + 0.2*trunkWidth*c.Rand.Float64(),
		Rand:       c.Rand,
		IsTrunk:    true,
	}
	trunk.SetEnd()
	return trunk
}

func (b *Branch) SetEnd() {
	b.End.X = b.Start.X + b.Length*math.Cos(b.Angle)
	b.End.Y = b.Start.Y - b.Length*math.Sin(b.Angle)
}

// Direction returns the normalized direction vector of the branch.
func (b *Branch) Direction() (dirX, dirY float64) {
	return math.Cos(b.Angle), -math.Sin(b.Angle)
}

// Perpendicular returns the normalized perpendicular vector to the branch direction.
func (b *Branch) Perpendicular() (perpX, perpY float64) {
	dirX, dirY := b.Direction()
	return -dirY, dirX
}

// AdjustStartForParent adjusts the child branch start point to overlap with the
// parent branch end, eliminating gaps at the connection point.
func (b *Branch) AdjustStartForParent(parent *Branch, branchType BranchType) {
	// Move the start point back along the parent direction to create slight overlap
	// This ensures the angled child branch overlaps with the parent end without showing flat top
	parentDirX, parentDirY := parent.Direction()
	overlapDist := b.StartWidth * 0.6
	b.Start.X = parent.End.X - parentDirX*overlapDist
	b.Start.Y = parent.End.Y - parentDirY*overlapDist

	// Apply perpendicular offset to align edges
	// For LeftBranch: align left edges (move child right)
	// For RightBranch: align right edges (move child left)
	sign := 1.0
	if branchType == RightBranch {
		sign = -1.0
	}
	parentPerpX, parentPerpY := parent.Perpendicular()
	widthDiff := sign * (parent.EndWidth - b.StartWidth) / 2
	b.Start.X += parentPerpX * widthDiff
	b.Start.Y += parentPerpY * widthDiff
}

type BranchType int

const (
	MidBranch BranchType = iota
	LeftBranch
	RightBranch
)

// GenerateBranchesBFS generates branches in breadth-first order so branches
// at the same depth level are added together.
func (c *Canvas) GenerateBranchesBFS(root *Branch, maxDepth int) {
	type queueItem struct {
		branch *Branch
		depth  int
	}
	queue := []queueItem{{branch: root, depth: maxDepth}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth <= 0 {
			continue
		}
		nextDepth := item.depth - 1

		// Generate child branches
		var children []*Branch
		// 2 at the end
		if left := item.branch.Add(LeftBranch); left != nil {
			children = append(children, left)
		}
		if right := item.branch.Add(RightBranch); right != nil {
			children = append(children, right)
		}
		//  + 1 branch in middle except at last level
		if nextDepth >= 1 {
			if mid := item.branch.Add(MidBranch); mid != nil {
				children = append(children, mid)
			}
		}

		// Add all children of this branch to Branches slice (same depth together)
		for _, child := range children {
			c.Branches = append(c.Branches, child)
			queue = append(queue, queueItem{branch: child, depth: nextDepth})
		}
	}
}

// Add a branch.
func (b *Branch) Add(t BranchType) *Branch {
	if b == nil || b.Length < 1 {
		// parent non existent or too small, skip
		return nil
	}
	// Pick branch point along parent branch
	dist := b.Length // End of branch for left/right
	if t == MidBranch {
		dist = b.Length * (0.3 + 0.3*b.Rand.Float64()) // Random point along branch for mid
	}
	dirX, dirY := b.Direction()
	branchPoint := Point{
		X: b.Start.X + dist*dirX,
		Y: b.Start.Y + dist*dirY,
	}
	// Calculate new branch parameters
	newLength := b.Length * (0.4 + 0.5*b.Rand.Float64())
	newAngle := b.calculateChildAngle(t)
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
	// Adjust start point for terminal branches to make edges contiguous
	if t == LeftBranch || t == RightBranch {
		newB.AdjustStartForParent(b, t)
	}
	newB.SetEnd()
	return newB
}

// calculateChildAngle computes the angle for a child branch based on branch type.
func (b *Branch) calculateChildAngle(t BranchType) float64 {
	wiggle := (b.Rand.Float64() - 0.5) * (math.Pi / 20)
	switch t {
	case LeftBranch:
		return b.Angle - (math.Pi / 6) + wiggle
	case RightBranch:
		return b.Angle + (math.Pi / 6) + wiggle
	case MidBranch:
		// Pick side randomly
		sign := 1.0
		if b.Rand.Float64() < 0.5 {
			sign = -1.0
		}
		return b.Angle + sign*(math.Pi/8) + wiggle
	default:
		panic("unknown branch type")
	}
}
