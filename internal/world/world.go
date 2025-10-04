package world

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type World struct {
	// Режим 1: тайлы
	TileTex  rl.Texture2D
	TileSize int32
	Cols     int
	Rows     int

	// Режим 2: фон-карта одной картинкой
	Backdrop    rl.Texture2D
	UseBackdrop bool
	Scale       float32 // во сколько раз растягивать картинку

	// Общая метрика мира в пикселях
	WidthPx  float32
	HeightPx float32
}

// ---------- ТАЙЛЫ ----------
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

	w := &World{
		TileTex:  tex,
		TileSize: tileSize,
		Cols:     cols,
		Rows:     rows,
	}
	w.WidthPx = float32(tileSize) * float32(cols)
	w.HeightPx = float32(tileSize) * float32(rows)
	return w, nil
}

// ---------- ФОН-КАРТА (одно изображение) ----------
func LoadBackdrop(assetsRoot string, relPath string, scale float32) (*World, error) {
	img := rl.LoadImage(filepath.Join(assetsRoot, relPath))
	if img.Data == nil {
		return nil, fmt.Errorf("backdrop image not found: %s", relPath)
	}
	defer rl.UnloadImage(img)
	tex := rl.LoadTextureFromImage(img)
	if tex.ID == 0 {
		return nil, fmt.Errorf("backdrop texture failed: %s", relPath)
	}
	rl.SetTextureFilter(tex, rl.FilterPoint)

	w := &World{
		Backdrop:    tex,
		UseBackdrop: true,
		Scale:       scale,
	}
	w.WidthPx = float32(tex.Width) * scale
	w.HeightPx = float32(tex.Height) * scale
	return w, nil
}

func (w *World) Unload() {
	if w.TileTex.ID != 0 {
		rl.UnloadTexture(w.TileTex)
		w.TileTex = rl.Texture2D{}
	}
	if w.Backdrop.ID != 0 {
		rl.UnloadTexture(w.Backdrop)
		w.Backdrop = rl.Texture2D{}
	}
}

func (w *World) SizePx() (float32, float32) { return w.WidthPx, w.HeightPx }

func (w *World) Clamp(x, y float32) (float32, float32) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > w.WidthPx {
		x = w.WidthPx
	}
	if y > w.HeightPx {
		y = w.HeightPx
	}
	return x, y
}

// Рисуем только видимое (для тайлов) или целиком (для фон-карты)
func (w *World) Draw(cam rl.Camera2D) {
	if w.UseBackdrop {
		if w.Backdrop.ID == 0 {
			return
		}
		src := rl.NewRectangle(0, 0, float32(w.Backdrop.Width), float32(w.Backdrop.Height))
		dst := rl.NewRectangle(0, 0, w.WidthPx, w.HeightPx) // растянем под масштаб
		rl.DrawTexturePro(w.Backdrop, src, dst, rl.NewVector2(0, 0), 0, rl.White)
		return
	}

	// тайловая отрисовка
	if w.TileTex.ID == 0 {
		return
	}
	t := float32(w.TileSize)
	screenW := float32(rl.GetScreenWidth())
	screenH := float32(rl.GetScreenHeight())
	topLeft := rl.GetScreenToWorld2D(rl.NewVector2(0, 0), cam)
	botRight := rl.GetScreenToWorld2D(rl.NewVector2(screenW, screenH), cam)
	x0 := int(topLeft.X/t) - 1
	if x0 < 0 {
		x0 = 0
	}
	y0 := int(topLeft.Y/t) - 1
	if y0 < 0 {
		y0 = 0
	}
	x1 := int(botRight.X/t) + 1
	if x1 > w.Cols {
		x1 = w.Cols
	}
	y1 := int(botRight.Y/t) + 1
	if y1 > w.Rows {
		y1 = w.Rows
	}

	src := rl.NewRectangle(0, 0, float32(w.TileTex.Width), float32(w.TileTex.Height))
	for ty := y0; ty < y1; ty++ {
		for tx := x0; tx < x1; tx++ {
			dst := rl.NewRectangle(float32(tx)*t, float32(ty)*t, t, t)
			rl.DrawTexturePro(w.TileTex, src, dst, rl.NewVector2(0, 0), 0, rl.White)
		}
	}
}
