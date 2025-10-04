package entities

import (
	"example.com/my2dgame/internal/anim"
	rl "github.com/gen2brain/raylib-go/raylib"
	"path/filepath"
)

type Player struct {
	X, Y  float32
	Speed float32
	Idle  *anim.Clip
	A     anim.Animator
	Scale float32 // << добавить
}

func NewPlayer(assetsRoot string) (*Player, error) {
	clip, err := anim.LoadFromJSON(filepath.Join(assetsRoot, "textures/ghost/idle/anim.json"))
	if err != nil {
		return nil, err
	}

	p := &Player{
		X: 200, Y: 300,
		Speed: 300, // твой повышенный спид оставил
		Idle:  clip,
		Scale: 1.25, // было 3.0, теперь ~в 2.4 раза меньше
	}
	p.A.Play(p.Idle, true)
	return p, nil
}

func (p *Player) Update(dt float32) {
	moveX, moveY := float32(0), float32(0)
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		moveX -= 1
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		moveX += 1
	}
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		moveY -= 1
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		moveY += 1
	}

	// нормализация диагонали
	if moveX != 0 && moveY != 0 {
		moveX *= 0.70710678
		moveY *= 0.70710678
	}

	// БАЗОВЫЙ КАДР СМОТРИТ ВЛЕВО:
	// идём вправо -> нужно зеркалить; идём влево -> без флипа
	if moveX > 0 {
		p.A.FlipX = true
	}
	if moveX < 0 {
		p.A.FlipX = false
	}

	p.X += moveX * p.Speed * dt
	p.Y += moveY * p.Speed * dt

	p.A.Update(dt)
}

func (p *Player) Draw() {
	p.A.Draw(p.X, p.Y, p.Scale, rl.White)
}
