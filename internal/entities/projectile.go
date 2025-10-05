package entities

import (
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	slimeBoltTex rl.Texture2D // враг
	ghostBoltTex rl.Texture2D // игрок
	projLoaded   bool
)

// Вызываем один раз из main: загрузка текстур для снарядов
func LoadProjectileAssets(assetsRoot string) error {
	if projLoaded {
		return nil
	}

	// враг
	{
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
	}

	// игрок
	{
		path := filepath.Join(assetsRoot, "textures", "projectiles", "ghost", "bolt.png")
		img := rl.LoadImage(path)
		if img.Data == nil {
			return fmt.Errorf("open projectile: %s", path)
		}
		defer rl.UnloadImage(img)
		ghostBoltTex = rl.LoadTextureFromImage(img)
		if ghostBoltTex.ID == 0 {
			return fmt.Errorf("texture from: %s", path)
		}
		rl.SetTextureFilter(ghostBoltTex, rl.FilterPoint)
	}

	projLoaded = true
	return nil
}

type Projectile struct {
	X, Y         float32
	PrevX, PrevY float32
	VX, VY       float32
	Speed        float32
	Life         float32
	Alive        bool
	Scale        float32
	HitRadius    float32
	FromPlayer   bool
	tex          *rl.Texture2D
	Damage       int
}

// Универсальный конструктор
func NewProjectile(x, y, dirX, dirY float32, fromPlayer bool) *Projectile {
	l := float32(rl.Vector2Length(rl.NewVector2(dirX, dirY)))
	if l == 0 {
		l = 1
	}
	nx, ny := dirX/l, dirY/l

	var tex *rl.Texture2D
	if fromPlayer {
		tex = &ghostBoltTex
	} else {
		tex = &slimeBoltTex
	}

	speed := float32(400)
	scale := float32(1.3)
	if !fromPlayer {
		speed = 300
		scale = 1.4
	}

	return &Projectile{
		X: x, Y: y,
		PrevX: x, PrevY: y,
		VX: nx, VY: ny,
		Speed:      speed,
		Life:       3.0,
		Alive:      true,
		Scale:      scale,
		HitRadius:  8 * scale,
		FromPlayer: fromPlayer,
		tex:        tex,
		Damage:     10,
	}
}

// Вспомогательные обёртки
func NewSlimeBolt(x, y, dx, dy float32) *Projectile { return NewProjectile(x, y, dx, dy, false) }
func NewGhostBolt(x, y, dx, dy float32) *Projectile { return NewProjectile(x, y, dx, dy, true) }

func (p *Projectile) Update(dt float32) {
	if !p.Alive {
		return
	}
	p.PrevX, p.PrevY = p.X, p.Y
	p.X += p.VX * p.Speed * dt
	p.Y += p.VY * p.Speed * dt
	p.Life -= dt
	if p.Life <= 0 {
		p.Alive = false
	}
}

func (p *Projectile) Draw() {
	if !p.Alive || p.tex == nil || p.tex.ID == 0 {
		return
	}

	w := float32(p.tex.Width)
	h := float32(p.tex.Height)
	src := rl.NewRectangle(0, 0, w, h)
	dst := rl.NewRectangle(p.X, p.Y, w*p.Scale, h*p.Scale)
	origin := rl.NewVector2((w*p.Scale)/2, (h*p.Scale)/2)

	rl.DrawTexturePro(*p.tex, src, dst, origin, 0, rl.White)
}
