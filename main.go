package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

const (
	assetFile  = "assets/gopher.png"
	updateRate = 50 * time.Millisecond
)

var (
	debug       bool
	followMouse bool
)

func getBgForEye(num int) color.Color {
	if !debug {
		return color.White
	}
	switch num {
	case 0:
		return color.RGBA{0, 0, 255, 255} // Blue
	case 1:
		return color.RGBA{0, 255, 0, 255} // Green
	default:
		return color.RGBA{255, 0, 0, 255} // Red
	}
}

func debugLog(fmt string, args ...interface{}) {
	if debug {
		log.Printf(fmt, args...)
	}
}

func main() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&followMouse, "follow", true, "Follow mouse")
	flag.Parse()

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title: "xgopher",
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		var assetReader io.Reader
		if !debug {
			assetReader = bytes.NewReader(MustAsset(assetFile))
		} else {
			f, err := os.Open(assetFile)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			assetReader = f
		}

		mascot, err := newMascot(assetReader, s, Eye{
			Background:        Circle{image.Point{527, 338}, 176, getBgForEye(0)}, // Relative to image
			Pupil:             Circle{image.Point{80, 193}, 63, color.Black},      // Relative to Background (125x126 @ 362,286)
			Glare:             Circle{image.Point{81, 78}, 16, color.White},       // Relative to Pupil (32x32 @ 427,348)
			PupilEllipseRatio: Point64{0.925, 1},                                  // Found using refPupilW/refPupilH
		}, Eye{
			Background:        Circle{image.Point{1045, 314}, 180, getBgForEye(1)}, // (372x372 @ 859,128)
			Pupil:             Circle{image.Point{103, 207}, 60, color.Black},      // (121x121 @ 882,272) (c.Y copied from above, +14)
			Glare:             Circle{image.Point{87, 75}, 14, color.White},        // (28x28 @ 955,333)
			PupilEllipseRatio: Point64{0.90, 1},
		} /* */)
		if err != nil {
			log.Fatal(err)
		}

		var (
			sz       image.Point
			ratio    float64
			offset   Point64
			look     Point64 = Point64{0.5, 0.5}
			lastDraw time.Time
		)

		for {
			e := w.NextEvent()

			/*
				if _, ok := e.(mouse.Event); !ok {
					format := "got %#v\n"
					if _, ok := e.(fmt.Stringer); ok {
						format = "got %v\n"
					}
					if true {
						fmt.Printf(format, e)
					}
				}
			*/

			switch e := e.(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					os.Exit(0)
				}

			case key.Event:
				if e.Code == key.CodeEscape {
					os.Exit(0)
				}

			case mouse.Event:
				if followMouse || (e.Button > 0 && e.Direction == mouse.DirRelease) {
					t := time.Now()
					if lastDraw.Add(updateRate).Before(t) {
						lastDraw = t

						look := Point64{float64(e.X), float64(e.Y)}
						drawEyes(w, mascot, ratio, offset, look, sz)
						w.Publish()
					}

				}

			case size.Event:
				if e.WidthPx == 0 && e.HeightPx == 0 { // Close window (macOS)
					os.Exit(0)
				}
				sz = e.Size()

			case paint.Event:
				debugLog("%#+v", sz)

				ratio = findRatio(mascot.Texture.Size(), sz)
				scaled := resizeByRatio(mascot.Texture.Bounds(), ratio, Point64{})
				var centered image.Rectangle
				centered, offset = centerRect(scaled, sz)

				debugLog("Ratio: %v Centered: %v Offset: %v", ratio, centered, offset)

				// paint (we do this in drawEyes from now on)
				//w.Fill(Rect(sz), color.Black, screen.Src)
				//w.Scale(centered, mascot.Texture, mascot.Texture.Bounds(), screen.Src, nil)

				drawEyes(w, mascot, ratio, offset, look, sz)
				w.Publish()

			case error:
				log.Print(e)
			}
		}
	})
}

func drawEyes(w screen.Window, m *Mascot, ratio float64, offset, look Point64, sz image.Point) {
	// paint the main mascot first
	w.Fill(Rect(sz), color.Black, screen.Src)
	centered := resizeByRatio(m.Texture.Bounds(), ratio, offset)
	w.Scale(centered, m.Texture, m.Texture.Bounds(), screen.Src, nil)

	scaleRatio := Point64{ratio, ratio}

	/*
		// Average out eye positions
		var eyeLook Point64
		for idx, e := range m.Eyes {
			eMin := Point64(MakePoint64(e.Background.Bounds().Min)).Mul(scaleRatio).Add(offset)
			eMax := Point64(MakePoint64(e.Background.Bounds().Max)).Mul(scaleRatio).Add(offset)
			eBounds := Rect64{Min: eMin, Max: eMax}

			l := look.Clip64(eBounds).Normalize64(eBounds)
			if idx == 0 {
				eyeLook = l
			} else {
				eyeLook = eyeLook.Add(l).Div(Point64{2, 2})
			}
		}
		eyeLook = eyeLook.Clip64(Rect64{Min: Point64{X: 0.15, Y: 0.15}, Max: Point64{X: 0.85, Y: 0.85}})
		debugLog("eyelook ratios: %v", eyeLook)
	*/

	for idx, e := range m.Eyes {
		// place eye with custom offset
		od := MakePoint64(e.Background.Bounds().Min).Mul(scaleRatio)
		ofs := offset.Add(od)
		//log.Print("offset diff: ", od)

		sc := resizeByRatio(e.Base.Bounds(), ratio, ofs)
		w.Scale(sc, e.Base, e.Base.Bounds(), screen.Over, nil)

		// place pupil
		eMin := Point64(MakePoint64(e.Background.Bounds().Min)).Mul(scaleRatio).Add(offset)
		eMax := Point64(MakePoint64(e.Background.Bounds().Max)).Mul(scaleRatio).Add(offset)
		eBounds := Rect64{Min: eMin, Max: eMax}

		eyeLook := look.Clip64(eBounds).Normalize64(eBounds).Clip64(Rect64{Min: Point64{X: 0.15, Y: 0.15}, Max: Point64{X: 0.85, Y: 0.85}})
		debugLog("eyelook ratio for eye %d: %v", idx, eyeLook)

		pb := resizeByRatio(e.Pupil.Bounds().Sub(image.Point{e.Pupil.Radius, e.Pupil.Radius}), 2*ratio, ofs)
		od = MakePoint64(pb.Size()).Mul(eyeLook)

		ofs = ofs.Add(od)
		//debugLog("offset diff: %v", od)

		sc = resizeByRatio(e.Texture.Bounds(), ratio, ofs)
		w.Scale(sc, e.Texture, e.Texture.Bounds(), screen.Over, nil)
	}
}
