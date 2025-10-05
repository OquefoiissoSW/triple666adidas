package ui

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type UltHUD struct {
	Textures [4]rl.Texture2D
	X, Y     float32
	Scale    float32
}

func LoadUltHUD(assetsRoot string, x, y, scale float32) (*UltHUD, error) {
	var hud UltHUD
	hud.X, hud.Y, hud.Scale = x, y, scale
	for i := 0; i < 4; i++ {
		path := filepath.Join(assetsRoot, "ui", fmt.Sprintf("charge_%d.png", i))
		img := rl.LoadImage(path)
		if img.Data == nil {
			return nil, fmt.Errorf("не найден %s", path)
		}
		hud.Textures[i] = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
	}
	return &hud, nil
}

func (h *UltHUD) Unload() {
	for _, t := range h.Textures {
		if t.ID != 0 {
			rl.UnloadTexture(t)
		}
	}
}

func (h *UltHUD) Draw(charge int) {
	if charge < 0 {
		charge = 0
	}
	if charge > 3 {
		charge = 3
	}
	t := h.Textures[charge]
	w := float32(t.Width) * h.Scale
	x := float32(rl.GetScreenWidth()) - w - h.X
	y := h.Y
	rl.DrawTextureEx(t, rl.NewVector2(x, y), 0, h.Scale, rl.White)
}
