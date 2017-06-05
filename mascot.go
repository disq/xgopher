package main

import (
	"image"
	"image/color"
	"os"

	_ "image/png"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
)

/*
ref: width="401.98px" height="559.472px"
left eye
<path fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" stroke="#000000" stroke-width="2.8214" stroke-linecap="round" d="
	M83.103,104.35c14.047,54.85,101.864,40.807,98.554-14.213C177.691,24.242,69.673,36.957,83.103,104.35"/>
<g>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" cx="107.324" cy="95.404" rx="14.829" ry="16.062"/>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" cx="114.069" cy="99.029" rx="3.496" ry="4.082"/>
</g>

right eye
<path fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" stroke="#000000" stroke-width="2.9081" stroke-linecap="round" d="
	M206.169,94.16c10.838,63.003,113.822,46.345,99.03-17.197C291.935,19.983,202.567,35.755,206.169,94.16"/>
<g>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" cx="231.571" cy="91.404" rx="14.582" ry="16.062"/>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" cx="238.204" cy="95.029" rx="3.438" ry="4.082"/>
</g>
*/
type Circle struct {
	Center image.Point
	Radius int
	Color  color.Color
}

type Eye struct {
	Background Circle
	Pupil      Circle
	Glare      Circle

	PupilEllipseRatio Point64 // Hack!

	Base    screen.Texture
	Texture screen.Texture
}

type Mascot struct {
	screen  screen.Screen
	Texture screen.Texture

	Eyes []Eye
}

func newMascot(filename string, s screen.Screen, eyePositions ...Eye) (*Mascot, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	im, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	buf, err := s.NewBuffer(im.Bounds().Size())
	if err != nil {
		return nil, err
	}

	m := &Mascot{screen: s, Eyes: eyePositions}

	draw.Draw(buf.RGBA(), buf.Bounds(), im, image.ZP, draw.Over)
	base, err := s.NewTexture(buf.Size())
	if err != nil {
		return nil, err
	}

	base.Upload(image.Point{}, buf, buf.Bounds())
	buf.Release()

	m.Texture = base

	err = m.prepareEyes()
	return m, err
}

func (m *Mascot) prepareEyes() error {
	for i, e := range m.Eyes {
		// base (whites)
		buf, err := m.screen.NewBuffer(e.Background.Bounds().Size())
		if err != nil {
			return err
		}

		bg := e.Background.translateFrom(e.Background.Bounds()) //bg.Bounds() is 0-based now

		draw.Draw(buf.RGBA(), bg.Bounds(), bg, image.ZP, draw.Over)

		t, err := m.screen.NewTexture(buf.Size())
		if err != nil {
			return err
		}
		t.Upload(image.ZP, buf, buf.Bounds())
		buf.Release()

		e.Base = t

		buf, err = m.screen.NewBuffer(e.Background.Bounds().Size())
		if err != nil {
			return err
		}

		// pupil + glare
		pu := e.Pupil.translateFrom(e.Pupil.Bounds())

		// poor man's ellipse
		pu2 := image.NewRGBA(pu.Bounds())
		sc, o := scaleRect(pu.Bounds(), e.PupilEllipseRatio)
		sc = image.Rectangle{Min: sc.Min.Add(o), Max: sc.Max.Add(o)}
		draw.ApproxBiLinear.Scale(pu2, sc.Bounds(), pu, pu.Bounds(), draw.Src, nil)
		draw.Draw(buf.RGBA(), bg.Bounds(), pu2, image.ZP, draw.Over)

		gl := e.Glare.translateTo(pu.Bounds())
		//gl2 := image.NewRGBA(gl.Bounds())
		//sc, o = scaleRect(gl.Bounds(), e.PupilEllipseRatio)
		//sc = image.Rectangle{Min: sc.Min.Add(o), Max: sc.Max.Add(o)}
		//draw.ApproxBiLinear.Scale(gl2, sc.Bounds(), gl, gl.Bounds(), draw.Src, nil)
		draw.Draw(buf.RGBA(), bg.Bounds(), gl, image.ZP, draw.Over)

		t, err = m.screen.NewTexture(buf.Size())
		if err != nil {
			return err
		}
		t.Upload(image.ZP, buf, buf.Bounds())
		buf.Release()

		e.Texture = t

		// set
		m.Eyes[i] = e
	}

	return nil
}

func (e Circle) Bounds() image.Rectangle {
	return image.Rect(e.Center.X-e.Radius, e.Center.Y-e.Radius, e.Center.X+e.Radius, e.Center.Y+e.Radius)
}

func (e Circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (e Circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-e.Center.X)+0.5, float64(y-e.Center.Y)+0.5, float64(e.Radius)
	if xx*xx+yy*yy < rr*rr {
		return e.Color
	}

	return color.Transparent
}

func (e Circle) translateFrom(parent image.Rectangle) Circle {
	return Circle{
		Center: e.Center.Sub(parent.Min),
		Radius: e.Radius,
		Color:  e.Color,
	}
}

func (e Circle) translateTo(parent image.Rectangle) Circle {

	return Circle{
		Center: e.Center.Add(parent.Min),
		Radius: e.Radius,
		Color:  e.Color,
	}
}
