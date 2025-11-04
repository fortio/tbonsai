package ptree

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
	"golang.org/x/image/vector"
)

func DrawTree(img draw.Image, c *Canvas, useLines bool) {
	var rast *vector.Rasterizer
	if !useLines {
		// Reuse single rasterizer for all branches
		rast = vector.NewRasterizer(img.Bounds().Dx(), img.Bounds().Dy())
		rast.DrawOp = draw.Over
	}
	// Draw branches
	for _, b := range c.Branches {
		rgb := getBranchColor(c, b)
		if useLines {
			drawBranchLine(img.(*image.NRGBA), b, rgb)
		} else {
			drawBranchPolygon(img.(*image.RGBA), b, rgb, rast)
		}
	}
	// Draw leaves after branches
	if c.Leaves {
		drawLeaves(img, c, useLines)
	}
}

func drawBranchLine(img *image.NRGBA, b *Branch, rgb tcolor.RGBColor) {
	ansipixels.DrawAALine(img, b.Start.X, b.Start.Y, b.End.X, b.End.Y, toNRGBA(rgb))
}

func toNRGBA(rgb tcolor.RGBColor) color.NRGBA {
	return color.NRGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

func toRGBA(rgb tcolor.RGBColor) color.RGBA {
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

// calcBoundingBox computes the bounding box for a set of points, adds margin, clamps to image bounds, and returns integer coordinates.
func calcBoundingBox(points []float64, imgBounds image.Rectangle) (x0Int, y0Int, x1Int, y1Int int, offscreen bool) {
	if len(points)%2 != 0 {
		panic("points must have even length")
	}
	minX, maxX := points[0], points[0]
	minY, maxY := points[1], points[1]
	for i := 2; i < len(points); i += 2 {
		x, y := points[i], points[i+1]
		minX = min(minX, x)
		maxX = max(maxX, x)
		minY = min(minY, y)
		maxY = max(maxY, y)
	}
	// Add 1px margin for anti-aliasing and clamp to image bounds
	x0 := max(0.0, minX-1)
	y0 := max(0.0, minY-1)
	x1 := min(float64(imgBounds.Dx()), maxX+1)
	y1 := min(float64(imgBounds.Dy()), maxY+1)
	if x0 >= x1 || y0 >= y1 {
		return 0, 0, 0, 0, true // Completely offscreen
	}
	x0Int = int(math.Floor(x0))
	y0Int = int(math.Floor(y0))
	x1Int = int(math.Ceil(x1))
	y1Int = int(math.Ceil(y1))
	return x0Int, y0Int, x1Int, y1Int, false
}

// getBranchColor determines the color for a branch based on depth and mode.
func getBranchColor(c *Canvas, b *Branch) tcolor.RGBColor {
	if c.Rainbow {
		// Random color per branch
		clr := tcolor.Oklchf(.7, .7, c.Rand.Float64())
		ct, data := clr.Decode()
		return tcolor.ToRGB(ct, data)
	}
	// Depth-based gradient from dark trunk to lighter branches
	if c.MaxDepth == 0 || !c.Leaves {
		return c.TrunkColor
	}
	// Interpolate from trunk color to lighter version
	// Depth 0 (trunk) = dark, MaxDepth = lightest
	t := max(0, float64(b.Depth-2)/float64(c.MaxDepth-2)) // Start lightening from depth 2
	// Lighten by increasing RGB values toward white
	lightenFactor := 1 + t // Gradually lighten
	r := uint8(min(255, float64(c.TrunkColor.R)*lightenFactor))
	g := uint8(min(255, float64(c.TrunkColor.G)*lightenFactor))
	blue := uint8(min(255, float64(c.TrunkColor.B)*lightenFactor))
	return tcolor.RGBColor{R: r, G: g, B: blue}
}

// getLeafColor returns a green color for leaves with some variation.
func getLeafColor(c *Canvas) tcolor.RGBColor {
	// Green with slight hue variation
	hue := 0.33 + (c.Rand.Float64()-0.5)*0.1 // Around 120° (green) with variation
	lightness := 0.5 + c.Rand.Float64()*0.2  // 0.5-0.7 range
	chroma := 0.6 + c.Rand.Float64()*0.2     // 0.6-0.8 range
	clr := tcolor.Oklchf(lightness, chroma, hue)
	ct, data := clr.Decode()
	return tcolor.ToRGB(ct, data)
}

// drawLeaves renders leaves at terminal and near-terminal branches.
func drawLeaves(img draw.Image, c *Canvas, useLines bool) {
	// Auto-detect resolution and adjust leaf parameters
	// High-res (Kitty/PNG): bigger leaves, more of them
	// Low-res (ANSI): smaller leaves, fewer of them
	imgWidth := img.Bounds().Dx()
	leafSizeMultiplier := c.LeafSize
	numLeavesBase := 3
	numLeavesTerminal := 6

	// If width < 200, we're in low-res ANSI mode
	if imgWidth < 200 {
		// Low-res: reduce leaf size significantly and use fewer leaves
		leafSizeMultiplier *= 0.5
		numLeavesBase = 1
		numLeavesTerminal = 1
	} else if imgWidth > 800 {
		// High-res large images: increase leaf size
		leafSizeMultiplier *= 2.0
	}

	// Allow manual override via LeafDensity
	if c.LeafDensity > 0 {
		numLeavesBase = c.LeafDensity
		numLeavesTerminal = c.LeafDensity + 2
	}

	for _, b := range c.Branches {
		// Draw leaves on branches near the end (top 2 depth levels)
		if b.Depth < c.MaxDepth-1 {
			continue
		}
		// More leaves at terminal branches, fewer at depth-1
		numLeaves := numLeavesBase
		if b.Depth == c.MaxDepth {
			numLeaves = numLeavesTerminal
		}
		// Distribute leaves along the branch
		for range numLeaves {
			// Position along branch (random with bias toward end)
			t := 0.3 + c.Rand.Float64()*0.7
			if b.Depth == c.MaxDepth {
				t = 0.5 + c.Rand.Float64()*0.5 // Even more toward end for terminal branches
			}
			dirX, dirY := b.Direction()
			leafX := b.Start.X + dirX*b.Length*t
			leafY := b.Start.Y + dirY*b.Length*t
			leafColor := getLeafColor(c)
			// Random angle for leaf orientation
			angle := c.Rand.Float64() * math.Pi * 2
			if useLines {
				drawLeafTriangle(img.(*image.NRGBA), leafX, leafY, angle, b.EndWidth, leafColor, leafSizeMultiplier, true)
			} else {
				drawLeafTriangle(img.(*image.RGBA), leafX, leafY, angle, b.EndWidth, leafColor, leafSizeMultiplier, false)
			}
		}
	}
}

// drawLeafTriangle renders a triangular leaf at the given position.
func drawLeafTriangle(
	img draw.Image, x, y, angle, branchWidth float64, rgb tcolor.RGBColor, sizeMultiplier float64, useLines bool,
) {
	// Leaf size proportional to branch width but larger
	baseSize := branchWidth * 4
	if baseSize < 8 {
		baseSize = 8 // Minimum visible size
	}
	baseSize *= sizeMultiplier
	// Add some size variation (±20%)
	sizeVariation := 0.8 + 0.4*math.Sin(x+y) // deterministic variation based on position
	leafSize := baseSize * sizeVariation

	// Triangle vertices: pointing in random direction
	// Tip of the leaf
	tipX := x + math.Cos(angle)*leafSize
	tipY := y + math.Sin(angle)*leafSize
	// Base of the leaf (two points)
	baseAngle1 := angle + math.Pi*0.75
	baseAngle2 := angle - math.Pi*0.75
	baseRadius := leafSize * 0.4
	base1X := x + math.Cos(baseAngle1)*baseRadius
	base1Y := y + math.Sin(baseAngle1)*baseRadius
	base2X := x + math.Cos(baseAngle2)*baseRadius
	base2Y := y + math.Sin(baseAngle2)*baseRadius

	if useLines {
		// Line mode: draw just one edge of the triangle
		ansipixels.DrawAALine(img.(*image.NRGBA), tipX, tipY, base1X, base1Y, toNRGBA(rgb))
	} else {
		// Polygon mode: fill the triangle
		fillTriangle(img.(*image.RGBA), tipX, tipY, base1X, base1Y, base2X, base2Y, rgb)
	}
}

// fillTriangle fills a triangle using vector rasterizer for smooth anti-aliased rendering.
func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 float64, rgb tcolor.RGBColor) {
	points := []float64{x1, y1, x2, y2, x3, y3}
	x0Int, y0Int, x1Int, y1Int, offscreen := calcBoundingBox(points, img.Bounds())
	if offscreen {
		return
	}
	wInt := x1Int - x0Int
	hInt := y1Int - y0Int

	// Create rasterizer for this triangle
	rast := vector.NewRasterizer(wInt, hInt)
	rast.DrawOp = draw.Over

	// Translate coordinates to local space
	dx, dy := float32(x0Int), float32(y0Int)
	rast.MoveTo(float32(x1)-dx, float32(y1)-dy)
	rast.LineTo(float32(x2)-dx, float32(y2)-dy)
	rast.LineTo(float32(x3)-dx, float32(y3)-dy)
	rast.ClosePath()

	// Rasterize to the bounding box region
	subImg := img.SubImage(image.Rect(x0Int, y0Int, x1Int, y1Int)).(*image.RGBA)
	rast.Draw(subImg, subImg.Bounds(), image.NewUniform(toRGBA(rgb)), image.Point{})
}

func drawBranchPolygon(img *image.RGBA, b *Branch, rgb tcolor.RGBColor, rast *vector.Rasterizer) {
	perpX, perpY := b.Perpendicular()
	if perpX == 0 && perpY == 0 {
		return
	}

	startHalfWidth := b.StartWidth / 2
	endHalfWidth := b.EndWidth / 2

	// Calculate the 4 vertices of the trapezoid
	var s1x, s1y, s2x, s2y float64
	if b.Depth == 0 {
		// Flat bottom for trunk
		s1x, s1y = b.Start.X+startHalfWidth, b.Start.Y
		s2x, s2y = b.Start.X-startHalfWidth, b.Start.Y
	} else {
		s1x, s1y = b.Start.X+perpX*startHalfWidth, b.Start.Y+perpY*startHalfWidth
		s2x, s2y = b.Start.X-perpX*startHalfWidth, b.Start.Y-perpY*startHalfWidth
	}

	e1x, e1y := b.End.X+perpX*endHalfWidth, b.End.Y+perpY*endHalfWidth
	e2x, e2y := b.End.X-perpX*endHalfWidth, b.End.Y-perpY*endHalfWidth

	points := []float64{s1x, s1y, s2x, s2y, e1x, e1y, e2x, e2y}
	x0Int, y0Int, x1Int, y1Int, offscreen := calcBoundingBox(points, img.Bounds())
	if offscreen {
		return
	}
	wInt := x1Int - x0Int
	hInt := y1Int - y0Int

	// Reset rasterizer to tight bounding box
	rast.Reset(wInt, hInt)
	// Translate coordinates to local space
	dx, dy := float32(x0Int), float32(y0Int)
	rast.MoveTo(float32(s1x)-dx, float32(s1y)-dy)
	rast.LineTo(float32(e1x)-dx, float32(e1y)-dy)
	rast.LineTo(float32(e2x)-dx, float32(e2y)-dy)
	rast.LineTo(float32(s2x)-dx, float32(s2y)-dy)
	rast.ClosePath()

	// Rasterize to the bounding box region
	subImg := img.SubImage(image.Rect(x0Int, y0Int, x1Int, y1Int)).(*image.RGBA)
	rast.Draw(subImg, subImg.Bounds(), image.NewUniform(toRGBA(rgb)), image.Point{})
}
