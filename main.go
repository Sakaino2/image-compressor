package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/Sakaino2/image-compressor/controllers"
	"github.com/chai2010/webp"
	"github.com/sqweek/dialog"
)

type FileItem struct {
	path      string
	removeBtn widget.Clickable
}

type App struct {
	theme        *material.Theme
	list         widget.List
	outputDir    widget.Editor
	quality      widget.Editor
	convertBtn   widget.Clickable
	browseBtn    widget.Clickable
	browseDirBtn widget.Clickable
	clearBtn     widget.Clickable
	statusText   string
	fileItems    []*FileItem
	processing   bool
}

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("WebP Image Compressor"))
		w.Option(app.Size(unit.Dp(700), unit.Dp(600)))

		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	a := &App{
		theme:      material.NewTheme(),
		statusText: "Ready to convert images. Select files to add.",
		fileItems:  []*FileItem{},
	}

	// Configure list
	a.list.Axis = layout.Vertical

	// Set default quality
	a.quality.SetText("80")

	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Handle convert button click
			if a.convertBtn.Clicked(gtx) && !a.processing {
				go a.convertImages(w)
			}

			// Handle browse files button click
			if a.browseBtn.Clicked(gtx) {
				go a.browseFiles(w)
			}

			// Handle browse directory button click
			if a.browseDirBtn.Clicked(gtx) {
				go a.browseDirectory(w)
			}

			// Handle clear button click
			if a.clearBtn.Clicked(gtx) {
				a.fileItems = []*FileItem{}
				a.statusText = "Files cleared. Select new files to convert."
				w.Invalidate()
			}

			// Handle individual remove buttons
			for i := len(a.fileItems) - 1; i >= 0; i-- {
				if a.fileItems[i].removeBtn.Clicked(gtx) {
					// Remove this item
					a.fileItems = append(a.fileItems[:i], a.fileItems[i+1:]...)
					a.statusText = fmt.Sprintf("File removed. %d file(s) remaining.", len(a.fileItems))
					w.Invalidate()
					break
				}
			}

			a.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{
		Top:    unit.Dp(20),
		Bottom: unit.Dp(20),
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:    layout.Vertical,
			Spacing: layout.SpaceBetween,
		}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					title := material.H5(a.theme, "WebP Image Compressor - Batch Converter")
					return title.Layout(gtx)
				})
			}),

			// Files label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(a.theme, fmt.Sprintf("Selected files (%d):", len(a.fileItems)))
					return label.Layout(gtx)
				})
			}),

			// Files list area
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					border := widget.Border{
						Color:        a.theme.Fg,
						CornerRadius: unit.Dp(4),
						Width:        unit.Dp(1),
					}
					return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:    unit.Dp(8),
							Bottom: unit.Dp(8),
							Left:   unit.Dp(8),
							Right:  unit.Dp(8),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// Set fixed height for the list
							gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(150))
							gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(150))

							if len(a.fileItems) == 0 {
								// Show placeholder text
								return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									label := material.Body2(a.theme, "No files selected")
									return label.Layout(gtx)
								})
							}

							// Show scrollable list of files
							return material.List(a.theme, &a.list).Layout(gtx, len(a.fileItems), func(gtx layout.Context, index int) layout.Dimensions {
								item := a.fileItems[index]
								return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{
										Axis:      layout.Horizontal,
										Alignment: layout.Middle,
										Spacing:   layout.SpaceBetween,
									}.Layout(gtx,
										// Filename
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											label := material.Body2(a.theme, filepath.Base(item.path))
											return label.Layout(gtx)
										}),
										// Remove button
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											btn := material.Button(a.theme, &item.removeBtn, "✕")
											btn.CornerRadius = unit.Dp(4)
											btn.Inset = layout.Inset{
												Top:    unit.Dp(4),
												Bottom: unit.Dp(4),
												Left:   unit.Dp(8),
												Right:  unit.Dp(8),
											}
											return btn.Layout(gtx)
										}),
									)
								})
							})
						})
					})
				})
			}),

			// Button row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(15)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceBetween,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(a.theme, &a.browseBtn, "Add File...")
							btn.CornerRadius = unit.Dp(4)
							return btn.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(a.theme, &a.clearBtn, "Clear All")
								btn.CornerRadius = unit.Dp(4)
								return btn.Layout(gtx)
							})
						}),
					)
				})
			}),

			// Output directory label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(a.theme, "Output directory (optional):")
					return label.Layout(gtx)
				})
			}),

			// Output directory with browse button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(15)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceBetween,
					}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							a.outputDir.SingleLine = true
							editor := material.Editor(a.theme, &a.outputDir, "Leave empty to save next to originals")
							border := widget.Border{
								Color:        a.theme.Fg,
								CornerRadius: unit.Dp(4),
								Width:        unit.Dp(1),
							}
							return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Top:    unit.Dp(8),
									Bottom: unit.Dp(8),
									Left:   unit.Dp(8),
									Right:  unit.Dp(8),
								}.Layout(gtx, editor.Layout)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Left: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(a.theme, &a.browseDirBtn, "Browse...")
								btn.CornerRadius = unit.Dp(4)
								return btn.Layout(gtx)
							})
						}),
					)
				})
			}),

			// Quality label
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					label := material.Body1(a.theme, "Quality (1-100):")
					return label.Layout(gtx)
				})
			}),

			// Quality editor
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					a.quality.SingleLine = true
					editor := material.Editor(a.theme, &a.quality, "80")
					border := widget.Border{
						Color:        a.theme.Fg,
						CornerRadius: unit.Dp(4),
						Width:        unit.Dp(1),
					}
					return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Top:    unit.Dp(8),
							Bottom: unit.Dp(8),
							Left:   unit.Dp(8),
							Right:  unit.Dp(8),
						}.Layout(gtx, editor.Layout)
					})
				})
			}),

			// Convert button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(15)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					btnText := "Convert All to WebP"
					if a.processing {
						btnText = "Converting..."
					}
					btn := material.Button(a.theme, &a.convertBtn, btnText)
					btn.CornerRadius = unit.Dp(4)
					return btn.Layout(gtx)
				})
			}),

			// Status text
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				status := material.Body2(a.theme, a.statusText)
				return status.Layout(gtx)
			}),
		)
	})
}

func (a *App) browseFiles(w *app.Window) {
	filename, err := dialog.File().
		Title("Select Image to Add (click Add File again for more)").
		Filter("Image Files", "jpg", "jpeg", "png", "bmp").
		Filter("All Files", "*").
		Load()

	if err != nil {
		if err.Error() != "Cancelled" {
			a.statusText = fmt.Sprintf("Error opening file dialog: %v", err)
			w.Invalidate()
		}
		return
	}

	// Check for duplicates
	for _, item := range a.fileItems {
		if item.path == filename {
			a.statusText = "File already in list"
			w.Invalidate()
			return
		}
	}

	// Add file to the list
	a.fileItems = append(a.fileItems, &FileItem{path: filename})

	a.statusText = fmt.Sprintf("Added: %s (Total: %d files). Click Add File to add more.", filepath.Base(filename), len(a.fileItems))
	w.Invalidate()
}

func (a *App) browseDirectory(w *app.Window) {
	directory, err := dialog.Directory().
		Title("Select Output Directory").
		Browse()

	if err != nil {
		if err.Error() != "Cancelled" {
			a.statusText = fmt.Sprintf("Error opening directory dialog: %v", err)
			w.Invalidate()
		}
		return
	}

	a.outputDir.SetText(directory)
	a.statusText = fmt.Sprintf("Output directory: %s", directory)
	w.Invalidate()
}

func (a *App) convertImages(w *app.Window) {
	if len(a.fileItems) == 0 {
		a.statusText = "Error: No files selected"
		w.Invalidate()
		return
	}

	qualityStr := a.quality.Text()
	outputDir := a.outputDir.Text()

	// Parse quality
	var quality float32 = 80
	if qualityStr != "" {
		var q int
		_, err := fmt.Sscanf(qualityStr, "%d", &q)
		if err != nil || q < 1 || q > 100 {
			a.statusText = "Error: Quality must be between 1 and 100"
			w.Invalidate()
			return
		}
		quality = float32(q)
	}

	// Validate output directory if specified
	if outputDir != "" {
		if info, err := os.Stat(outputDir); err != nil || !info.IsDir() {
			a.statusText = "Error: Invalid output directory"
			w.Invalidate()
			return
		}
	}

	a.processing = true
	a.statusText = "Converting files..."
	w.Invalidate()

	// Convert files with progress tracking
	var wg sync.WaitGroup
	results := make(chan string, len(a.fileItems))

	for i, item := range a.fileItems {
		wg.Add(1)
		go func(path string, index int) {
			defer wg.Done()

			// Determine output path
			var outputPath string
			if outputDir != "" {
				base := filepath.Base(path)
				ext := filepath.Ext(base)
				name := strings.TrimSuffix(base, ext) + ".webp"
				outputPath = filepath.Join(outputDir, name)
			} else {
				ext := filepath.Ext(path)
				outputPath = strings.TrimSuffix(path, ext) + ".webp"
			}

			// Convert the image
			err := convertImage(path, outputPath, quality)
			if err != nil {
				results <- fmt.Sprintf("❌ %s: %v", filepath.Base(path), err)
			} else {
				results <- fmt.Sprintf("✓ %s", filepath.Base(path))
			}

			// Update progress
			a.statusText = fmt.Sprintf("Converting... %d/%d", index+1, len(a.fileItems))
			w.Invalidate()
		}(item.path, i)
	}

	// Wait for all conversions to complete
	wg.Wait()
	close(results)

	// Collect results
	successCount := 0
	var resultsSummary strings.Builder
	for result := range results {
		if strings.HasPrefix(result, "✓") {
			successCount++
		}
		resultsSummary.WriteString(result)
		resultsSummary.WriteString("\n")
	}

	a.processing = false
	a.statusText = fmt.Sprintf("Complete! %d/%d files converted successfully", successCount, len(a.fileItems))
	w.Invalidate()

	log.Println(resultsSummary.String())
}

func convertImage(inputPath, outputPath string, quality float32) error {
	// Open input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Decode image
	img, err := controllers.DecodeImage(file, inputPath)

	if err != nil {
		return fmt.Errorf("decoding image: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	defer outFile.Close()

	// Encode as WebP
	err = webp.Encode(outFile, *img, &webp.Options{Quality: quality})
	if err != nil {
		return fmt.Errorf("encoding webp: %w", err)
	}

	return nil
}
