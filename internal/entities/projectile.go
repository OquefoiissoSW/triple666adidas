package entities

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	slimeBoltTex rl.Texture2D
	projLoaded   bool
)

// Вызываем один раз из main: загрузка текстур для снарядов
func LoadProjectileAssets(assetsRoot string) error {
	if projLoaded {
		return nil
	}
	path := filepath.Join(assetsRoot, "textures", "projectiles", "slime", "bolt.png")
	img := rl.LoadImage(path)
	if img.Data == nil {
		return fmt.Errorf("open projectile: %s", path)
	}
	defer rl.UnloadImage(img)
	slimeBoltTex = rl.LoadTextureFromImage(img)
	if slimeBoltTex.ID == 0 {
		return fmt.Errorf("texture from: %s", path)
	}
	rl.SetTextureFilter(slimeBoltTex, rl.FilterPoint)
	projLoaded = true
	return nil
}

type Projectile struct {
	X, Y   float32
	VX, VY float32
	Speed  float32
	Life   float32
	Alive  bool
	Scale  float32
}

func NewSlimeBolt(x, y, dirX, dirY float32) *Projectile {
	// нормализуем dir
	l := float32(rl.Vector2Length(rl.NewVector2(dirX, dirY)))
	if l == 0 {
		l = 1
	}
	nx, ny := dirX/l, dirY/l
	return &Projectile{
		X: x, Y: y,
		VX: nx, VY: ny,
		Speed: 300, // скорость пули
		Life:  3.0, // живёт 3 секунды
		Alive: true,
		Scale: 1.4, // масштаб отрисовки
	}
}

func (p *Projectile) Update(dt float32) {
	if !p.Alive {
		return
	}
	p.X += p.VX * p.Speed * dt
	p.Y += p.VY * p.Speed * dt
	p.Life -= dt
	if p.Life <= 0 {
		p.Alive = false
	}
}

func (p *Projectile) Draw() {
	if !p.Alive || slimeBoltTex.ID == 0 {
		return
	}
	w := float32(slimeBoltTex.Width)
	h := float32(slimeBoltTex.Height)
	src := rl.NewRectangle(0, 0, w, h)
	dst := rl.NewRectangle(p.X, p.Y, w*p.Scale, h*p.Scale)
	// отрисуем с центровкой
	origin := rl.NewVector2((w*p.Scale)/2, (h*p.Scale)/2)
	rl.DrawTexturePro(slimeBoltTex, src, dst, origin, 0, rl.White)
}
