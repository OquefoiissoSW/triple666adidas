package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type HealthHUD struct {
	tex   map[int]rl.Texture2D // ключ: процент (0..100)
	keys  []int                // отсортированный список доступных ключей
	X, Y  int32
	Scale float32
}

var numPrefix = regexp.MustCompile(`^(\d+)`)

// LoadHealthHUD загружает все PNG из assets/ui/health,
// чьё имя начинается с числа (пример: 7.png, 15.png, 35_hp.png).
func LoadHealthHUD(assetsRoot string, x, y int32, scale float32) (*HealthHUD, error) {
	dir := filepath.Join(assetsRoot, "ui", "health")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("open hud dir: %w", err)
	}
	h := &HealthHUD{
		tex:   make(map[int]rl.Texture2D),
		X:     x,
		Y:     y,
		Scale: scale,
	}

	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".png") {
			continue
		}

		m := numPrefix.FindStringSubmatch(name)
		if len(m) < 2 {
			continue
		} // нет числа в начале имени

		v, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}

		img := rl.LoadImage(filepath.Join(dir, name))
		if img.Data == nil {
			continue
		}
		t := rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		if t.ID == 0 {
			continue
		}

		rl.SetTextureFilter(t, rl.FilterPoint)
		// если один и тот же ключ встречается несколько раз, оставляем первый
		if _, exists := h.tex[v]; !exists {
			h.tex[v] = t
		} else {
			rl.UnloadTexture(t)
		}
	}

	if len(h.tex) == 0 {
		return nil, fmt.Errorf("no health images in %s", dir)
	}

	// подготовим отсортированные ключи
	for k := range h.tex {
		h.keys = append(h.keys, k)
	}
	sort.Ints(h.keys)
	return h, nil
}

func (h *HealthHUD) Unload() {
	for _, t := range h.tex {
		if t.ID != 0 {
			rl.UnloadTexture(t)
		}
	}
	h.tex = nil
	h.keys = nil
}

// nearestKey возвращает доступный ключ, ближайший к hp (0..100).
func (h *HealthHUD) nearestKey(hp int) int {
	if hp < 0 {
		hp = 0
	}
	if hp > 100 {
		hp = 100
	}
	bestK := h.keys[0]
	bestD := absInt(hp - bestK)
	for _, k := range h.keys[1:] {
		if d := absInt(hp - k); d < bestD {
			bestD, bestK = d, k
		}
	}
	return bestK
}

func (h *HealthHUD) Draw(hp int) {
	if len(h.tex) == 0 {
		return
	}
	k := h.nearestKey(hp)
	tex := h.tex[k]
	if tex.ID == 0 {
		return
	}

	w := float32(tex.Width) * h.Scale
	hh := float32(tex.Height) * h.Scale
	src := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
	dst := rl.NewRectangle(float32(h.X), float32(h.Y), w, hh)
	rl.DrawTexturePro(tex, src, dst, rl.NewVector2(0, 0), 0, rl.White)
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
