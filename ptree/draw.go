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

// getBranchColor determines the color for a branch based on depth and mode.
func getBranchColor(c *Canvas, b *Branch) tcolor.RGBColor {
	if c.Rainbow {
		// Random color per branch
		clr := tcolor.Oklchf(.7, .7, c.Rand.Float64())
		ct, data := clr.Decode()
		return tcolor.ToRGB(ct, data)
	}
	// Depth-based gradient from dark trunk to lighter branches
	if c.MaxDepth == 0 {
		return c.TrunkColor
	}
	// Interpolate from trunk color to lighter version
	// Depth 0 (trunk) = dark, MaxDepth = lightest
	t := float64(b.Depth) / float64(c.MaxDepth)
	// Lighten by increasing RGB values toward white
	lightenFactor := 1.0 + t*1.5 // Gradually lighten up to 2.5x
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
	numLeavesTerminal := 5

	// If width < 200, we're in low-res ANSI mode
	if imgWidth < 200 {
		// Low-res: reduce leaf size significantly and use fewer leaves
		leafSizeMultiplier *= 0.25
		numLeavesBase = 1
		numLeavesTerminal = 2
	} else if imgWidth > 800 {
		// High-res large images: increase leaf size
		leafSizeMultiplier *= 1.5
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
func drawLeafTriangle(img draw.Image, x, y, angle, branchWidth float64, rgb tcolor.RGBColor, sizeMultiplier float64, isNRGBA bool) {
	// Leaf size proportional to branch width but larger
	baseSize := branchWidth * 4 * sizeMultiplier
	if baseSize < 8 {
		baseSize = 8 // Minimum visible size
	}
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

	// Fill the triangle
	fillTriangle(img, tipX, tipY, base1X, base1Y, base2X, base2Y, rgb, isNRGBA)
}

// fillTriangle fills a triangle with the given color.
func fillTriangle(img draw.Image, x1, y1, x2, y2, x3, y3 float64, rgb tcolor.RGBColor, isNRGBA bool) {
	// Find bounding box
	minX := min(x1, x2, x3)
	maxX := max(x1, x2, x3)
	minY := min(y1, y2, y3)
	maxY := max(y1, y2, y3)

	bounds := img.Bounds()
	// Clamp to image bounds
	px0 := int(math.Floor(minX))
	py0 := int(math.Floor(minY))
	px1 := int(math.Ceil(maxX)) + 1
	py1 := int(math.Ceil(maxY)) + 1
	if px0 < bounds.Min.X {
		px0 = bounds.Min.X
	}
	if py0 < bounds.Min.Y {
		py0 = bounds.Min.Y
	}
	if px1 > bounds.Max.X {
		px1 = bounds.Max.X
	}
	if py1 > bounds.Max.Y {
		py1 = bounds.Max.Y
	}

	// Scan through bounding box and test each pixel
	for py := py0; py < py1; py++ {
		for px := px0; px < px1; px++ {
			// Use barycentric coordinates to test if point is inside triangle
			pxf := float64(px) + 0.5
			pyf := float64(py) + 0.5
			if pointInTriangle(pxf, pyf, x1, y1, x2, y2, x3, y3) {
				if isNRGBA {
					nrgbaImg := img.(*image.NRGBA)
					nrgbaImg.Set(px, py, color.NRGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255})
				} else {
					setBlendedPixel(img.(*image.RGBA), px, py, rgb, 200) // Slight transparency
				}
			}
		}
	}
}

// pointInTriangle tests if a point is inside a triangle using barycentric coordinates.
func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 float64) bool {
	// Compute vectors
	v0x := x3 - x1
	v0y := y3 - y1
	v1x := x2 - x1
	v1y := y2 - y1
	v2x := px - x1
	v2y := py - y1

	// Compute dot products
	dot00 := v0x*v0x + v0y*v0y
	dot01 := v0x*v1x + v0y*v1y
	dot02 := v0x*v2x + v0y*v2y
	dot11 := v1x*v1x + v1y*v1y
	dot12 := v1x*v2x + v1y*v2y

	// Compute barycentric coordinates
	invDenom := 1 / (dot00*dot11 - dot01*dot01)
	u := (dot11*dot02 - dot01*dot12) * invDenom
	v := (dot00*dot12 - dot01*dot02) * invDenom

	// Check if point is in triangle
	return (u >= 0) && (v >= 0) && (u+v <= 1)
}

// setBlendedPixel blends a color with an existing pixel.
func setBlendedPixel(img *image.RGBA, x, y int, rgb tcolor.RGBColor, alpha uint8) {
	existing := img.RGBAAt(x, y)
	blendAlpha := float64(alpha) / 255.0
	r := uint8(float64(rgb.R)*blendAlpha + float64(existing.R)*(1-blendAlpha))
	g := uint8(float64(rgb.G)*blendAlpha + float64(existing.G)*(1-blendAlpha))
	blue := uint8(float64(rgb.B)*blendAlpha + float64(existing.B)*(1-blendAlpha))
	img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: blue, A: 255})
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

	// Calculate tight bounding box
	minX := min(s1x, s2x, e1x, e2x)
	maxX := max(s1x, s2x, e1x, e2x)
	minY := min(s1y, s2y, e1y, e2y)
	maxY := max(s1y, s2y, e1y, e2y)

	// Add 1px margin for anti-aliasing and clamp to image bounds
	imgBounds := img.Bounds()
	x0 := max(0, minX-1)
	y0 := max(0, minY-1)
	x1 := min(float64(imgBounds.Dx()), maxX+1)
	y1 := min(float64(imgBounds.Dy()), maxY+1)

	if x0 >= x1 || y0 >= y1 {
		return // Completely offscreen
	}
	x0Int := int(math.Floor(x0))
	y0Int := int(math.Floor(y0))
	x1Int := int(math.Ceil(x1))
	y1Int := int(math.Ceil(y1))
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
