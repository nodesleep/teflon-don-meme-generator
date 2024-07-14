package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/image/font"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")
	e.File("/", "static/index.html")
	e.POST("/add-text", addTextHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

func addTextHandler(c echo.Context) error {
	text := strings.ToUpper(c.FormValue("text"))

	imgFile, err := os.Open("static/input.jpg")
	if err != nil {
		return err
	}
	defer imgFile.Close()

	img, err := jpeg.Decode(imgFile)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	fontBytes, err := os.ReadFile("static/impact.ttf")
	if err != nil {
		return err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(f)
	fc.SetClip(newImg.Bounds())
	fc.SetDst(newImg)
	fc.SetSrc(image.NewUniform(color.White))

	maxFontSize := 100.0
	minFontSize := 10.0
	imgWidth := float64(bounds.Dx())

	var fontSize float64
	for fontSize = maxFontSize; fontSize >= minFontSize; fontSize-- {
		opts := truetype.Options{
			Size: fontSize,
		}
		face := truetype.NewFace(f, &opts)
		width := font.MeasureString(face, text).Ceil()
		if float64(width) <= imgWidth {
			break
		}
	}

	fc.SetFontSize(fontSize)

	face := truetype.NewFace(f, &truetype.Options{Size: fontSize})
	textWidth := font.MeasureString(face, text).Ceil()

	pt := freetype.Pt((bounds.Dx()-textWidth)/2, bounds.Dy()-10)

	_, err = fc.DrawString(text, pt)
	if err != nil {
		return err
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	outputFileName := fmt.Sprintf("output_%d.jpg", rng.Int())
	outputFilePath := filepath.Join("static/output", outputFileName)
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, newImg, nil)
	if err != nil {
		return err
	}

	return c.String(200, "/static/output/"+outputFileName)
}
