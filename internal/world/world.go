package world

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type World struct {
	TileTex  rl.Texture2D
	TileSize int32 // размер тайла в пикселях (например, 64)
	Cols     int   // ширина мира в тайлах
	Rows     int   // высота мира в тайлах
}

func Load(assetsRoot string, relPath string, tileSize int32, cols, rows int) (*World, error) {
	img := rl.LoadImage(filepath.Join(assetsRoot, relPath))
	if img.Data == nil {
		return nil, fmt.Errorf("tile image not found: %s", relPath)
	}
	defer rl.UnloadImage(img)
	tex := rl.LoadTextureFromImage(img)
	if tex.ID == 0 {
		return nil, fmt.Errorf("tile texture failed: %s", relPath)
	}
	rl.SetTextureFilter(tex, rl.FilterPoint)

	return &World{
		TileTex:  tex,
		TileSize: tileSize,
		Cols:     cols,
		Rows:     rows,
	}, nil
}

func (w *World) Unload() {
	if w.TileTex.ID != 0 {
		rl.UnloadTexture(w.TileTex)
		w.TileTex = rl.Texture2D{}
	}
}

func (w *World) SizePx() (float32, float32) {
	return float32(w.TileSize * int32(w.Cols)), float32(w.TileSize * int32(w.Rows))
}

func (w *World) Clamp(x, y float32) (float32, float32) {
	maxW, maxH := w.SizePx()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > maxW {
		x = maxW
	}
	if y > maxH {
		y = maxH
	}
	return x, y
}

// Рисуем только видимую область для производительности
func (w *World) Draw(cam rl.Camera2D) {
	if w.TileTex.ID == 0 {
		return
	}

	t := float32(w.TileSize)
	screenW := float32(rl.GetScreenWidth())
	screenH := float32(rl.GetScreenHeight())

	// мировые координаты видимого прямоугольника
	topLeft := rl.GetScreenToWorld2D(rl.NewVector2(0, 0), cam)
	botRight := rl.GetScreenToWorld2D(rl.NewVector2(screenW, screenH), cam)

	x0 := int(topLeft.X/t) - 1
	y0 := int(topLeft.Y/t) - 1
	x1 := int(botRight.X/t) + 1
	y1 := int(botRight.Y/t) + 1

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > w.Cols {
		x1 = w.Cols
	}
	if y1 > w.Rows {
		y1 = w.Rows
	}

	src := rl.NewRectangle(0, 0, float32(w.TileTex.Width), float32(w.TileTex.Height))

	for ty := y0; ty < y1; ty++ {
		for tx := x0; tx < x1; tx++ {
			dst := rl.NewRectangle(
				float32(tx)*t,
				float32(ty)*t,
				t, t,
			)
			rl.DrawTexturePro(w.TileTex, src, dst, rl.NewVector2(0, 0), 0, rl.White)
		}
	}
}
