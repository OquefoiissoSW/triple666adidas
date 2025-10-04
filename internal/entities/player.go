package entities

import (
	"example.com/my2dgame/internal/anim"
	rl "github.com/gen2brain/raylib-go/raylib"
	"path/filepath"
)

type Player struct {
	X, Y         float32
	PrevX, PrevY float32
	Speed        float32
	Idle         *anim.Clip
	A            anim.Animator
	Scale        float32 // << добавить
	HP           int     // 0..100
	Radius       float32 // для коллизий (окружность)
	InvulnTimer  float32 // секунды неуязвимости после удара
	HurtFlash    float32
}

func NewPlayer(assetsRoot string) (*Player, error) {
	clip, err := anim.LoadFromJSON(filepath.Join(assetsRoot, "textures/ghost/idle/anim.json"))
	if err != nil {
		return nil, err
	}

	p := &Player{
		X: 200, Y: 300,
		Speed:  300,
		Idle:   clip,
		Scale:  1.25,
		HP:     100,
		Radius: 18,
		// HurtFlash: 0, // по умолчанию
	}

	p.PrevX, p.PrevY = p.X, p.Y
	p.A.Play(p.Idle, true)
	return p, nil
}

func (p *Player) Update(dt float32) {
	p.PrevX, p.PrevY = p.X, p.Y
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
	if p.InvulnTimer > 0 {
		p.InvulnTimer -= dt
		if p.InvulnTimer < 0 {
			p.InvulnTimer = 0
		}
	}
	if p.HurtFlash > 0 {
		p.HurtFlash -= dt
		if p.HurtFlash < 0 {
			p.HurtFlash = 0
		}
	}
}

func (p *Player) TakeDamage(dmg int) {
	if dmg <= 0 || p.HP <= 0 || p.InvulnTimer > 0 {
		return
	}
	p.HP -= dmg
	if p.HP < 0 {
		p.HP = 0
	}
	p.InvulnTimer = 0.5 // неуязвимость
	p.HurtFlash = 0.25  // 🔴 250 мс красный флэш
}

func (p *Player) Draw() {
	tint := rl.White
	if p.HurtFlash > 0 {
		// вариант А: ровный красный флэш
		tint = rl.NewColor(255, 64, 64, 255)

		// вариант B (мигание 10 Гц):
		// if int(p.HurtFlash*20)%2 == 0 { tint = rl.NewColor(255, 64, 64, 255) } else { tint = rl.White }
	}
	p.A.Draw(p.X, p.Y, p.Scale, tint)
}
