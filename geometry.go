package main

import (
	"image"
	"math"
)

type Point64 struct {
	X, Y float64
}

func MakePoint64(p image.Point) Point64 {
	return Point64{float64(p.X), float64(p.Y)}
}

func (p Point64) Div(q Point64) Point64 {
	return Point64{p.X / q.X, p.Y / q.Y}
}

func (p Point64) Mul(q Point64) Point64 {
	return Point64{p.X * q.X, p.Y * q.Y}
}

func (p Point64) Add(q Point64) Point64 {
	return Point64{p.X + q.X, p.Y + q.Y}
}

func (p Point64) Normalize(bounds image.Rectangle) Point64 {
	p.X = math.Abs((p.X - float64(bounds.Min.X)) / float64(bounds.Dx()))
	p.Y = math.Abs((p.Y - float64(bounds.Min.Y)) / float64(bounds.Dy()))
	return p
}

func (p Point64) Clip(bounds image.Rectangle) Point64 {
	p.X = math.Min(math.Max(p.X, float64(bounds.Min.X)), float64(bounds.Max.X))
	p.Y = math.Min(math.Max(p.Y, float64(bounds.Min.Y)), float64(bounds.Max.Y))
	return p
}

func Rect1() image.Rectangle {
	return image.Rect(0, 0, 1, 1)
}

func Rect(p image.Point) image.Rectangle {
	return image.Rect(0, 0, p.X, p.Y)
}

type Rect64 struct {
	Min Point64
	Max Point64
}

func MakeRect64(r image.Rectangle) Rect64 {
	return Rect64{Min: MakePoint64(r.Min), Max: MakePoint64(r.Max)}
}

func (r Rect64) Dx() float64 {
	return r.Max.X - r.Min.X
}

func (r Rect64) Dy() float64 {
	return r.Max.Y - r.Min.Y
}

func (r Rect64) Size() Point64 {
	return Point64{
		r.Max.X - r.Min.X,
		r.Max.Y - r.Min.Y,
	}
}

func (p Point64) Normalize64(bounds Rect64) Point64 {
	p.X = math.Abs((p.X - bounds.Min.X) / bounds.Dx())
	p.Y = math.Abs((p.Y - bounds.Min.Y) / bounds.Dy())
	return p
}

func (p Point64) Clip64(bounds Rect64) Point64 {
	p.X = math.Min(math.Max(p.X, bounds.Min.X), bounds.Max.X)
	p.Y = math.Min(math.Max(p.Y, bounds.Min.Y), bounds.Max.Y)
	return p
}

func findRatio(curSize image.Point, newSize image.Point) (ratio float64) {
	// scale
	ratioW := float64(newSize.X) / float64(curSize.X)
	ratio = float64(newSize.Y) / float64(curSize.Y)
	if ratioW < ratio {
		ratio = ratioW
	}
	return
}

func resizeByRatio(cur image.Rectangle, ratio float64, offset Point64) (rect image.Rectangle) {
	rect.Min.X = int(float64(cur.Min.X)*ratio + offset.X)
	rect.Min.Y = int(float64(cur.Min.Y)*ratio + offset.Y)
	rect.Max.X = int(float64(cur.Max.X)*ratio + offset.X)
	rect.Max.Y = int(float64(cur.Max.Y)*ratio + offset.Y)
	return
}

func centerRect(src image.Rectangle, newSize image.Point) (rect image.Rectangle, offset Point64) {
	// center
	offset.X = float64(newSize.X-src.Max.X) / 2
	offset.Y = float64(newSize.Y-src.Max.Y) / 2
	rect.Min.X = int(offset.X)
	rect.Min.Y = int(offset.Y)
	rect.Max.X = src.Max.X + rect.Min.X
	rect.Max.Y = src.Max.Y + rect.Min.Y
	return
}

func scaleRect(r image.Rectangle, ratio Point64) (rect image.Rectangle, offset image.Point) {
	sz := r.Size()
	offset.X = int(float64(sz.X) * (1 - ratio.X) / 2)
	offset.Y = int(float64(sz.Y) * (1 - ratio.Y) / 2)

	rect.Min.X = int(float64(r.Min.X) * ratio.X)
	rect.Max.X = int(float64(r.Max.X) * ratio.X)

	rect.Min.Y = int(float64(r.Min.Y) * ratio.Y)
	rect.Max.Y = int(float64(r.Max.Y) * ratio.Y)

	return
}
