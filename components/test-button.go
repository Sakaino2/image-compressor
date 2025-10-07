package components

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/Sakaino2/image-compressor/controllers"
	"github.com/chai2010/webp"
)

type App struct {
	window     *app.Window
	theme      *material.Theme
	inputPath  widget.Editor
	outputPath widget.Editor
	quality    widget.Editor
	convertBtn widget.Clickable
	statusText string
}

func main() {
	go func() {

		w := new(app.Window)
		w.Option(app.Size(unit.Dp(800), unit.Dp(700)))
		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	a := &App{
		window:     w,
		theme:      material.NewTheme(),
		statusText: "Ready to convert images",
	}

	// Set default quality
	a.quality.SetText("80")

	var ops op.Ops

	log.Println("running")

	for {
		e := <-make(chan event.Event)
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// Handle convert button click
	if a.convertBtn.Clicked(gtx) {
		go a.convertImage()
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(20),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				title := material.H5(a.theme, "WebP Image Compressor")
				return title.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(5),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(a.theme, "Input image path:")
				return label.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Bottom: unit.Dp(10),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				a.inputPath.SingleLine = true
				editor := material.Editor(a.theme, &a.inputPath, "e.g., /path/to/image.jpg")
				return editor.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(5),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(a.theme, "Output path (optional):")
				return label.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Bottom: unit.Dp(10),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				a.outputPath.SingleLine = true
				editor := material.Editor(a.theme, &a.outputPath, "Leave empty for auto-generated name")
				return editor.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(5),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Body1(a.theme, "Quality (1-100):")
				return label.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Bottom: unit.Dp(20),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				a.quality.SingleLine = true
				editor := material.Editor(a.theme, &a.quality, "80")
				return editor.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(20),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(a.theme, &a.convertBtn, "Convert to WebP")
				return btn.Layout(gtx)
			})
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Left:   unit.Dp(20),
				Right:  unit.Dp(20),
				Bottom: unit.Dp(20),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				status := material.Body2(a.theme, a.statusText)
				return status.Layout(gtx)
			})
		}),
	)
}

func (a *App) convertImage() {
	inputPath := a.inputPath.Text()
	outputPath := a.outputPath.Text()
	qualityStr := a.quality.Text()

	if inputPath == "" {
		a.statusText = "Error: Please provide an input path"
		a.window.Invalidate()
		return
	}

	// Parse quality
	var quality float32 = 80
	if qualityStr != "" {
		var q int
		_, err := fmt.Sscanf(qualityStr, "%d", &q)
		if err != nil || q < 1 || q > 100 {
			a.statusText = "Error: Quality must be between 1 and 100"
			a.window.Invalidate()
			return
		}
		quality = float32(q)
	}

	// Generate output path if not provided
	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		outputPath = strings.TrimSuffix(inputPath, ext) + ".webp"
	}

	a.statusText = "Converting..."
	a.window.Invalidate()

	// Open input file
	file, err := os.Open(inputPath)
	if err != nil {
		a.statusText = fmt.Sprintf("Error opening file: %v", err)
		a.window.Invalidate()
		return
	}
	defer file.Close()

	img, err := controllers.DecodeImage(file, inputPath)
	if err != nil {
		a.statusText = fmt.Sprintf("Error decoding image: %v", err)
		a.window.Invalidate()
		return
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		a.statusText = fmt.Sprintf("Error creating output file: %v", err)
		a.window.Invalidate()
		return
	}
	defer outFile.Close()

	// Encode as WebP
	err = webp.Encode(outFile, *img, &webp.Options{Quality: quality})
	if err != nil {
		a.statusText = fmt.Sprintf("Error encoding WebP: %v", err)
		a.window.Invalidate()
		return
	}

	a.statusText = fmt.Sprintf("Success! Saved to: %s", outputPath)
	a.window.Invalidate()
}
