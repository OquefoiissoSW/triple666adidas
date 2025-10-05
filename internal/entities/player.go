package entities

import (
	"path/filepath"

	"example.com/my2dgame/internal/anim"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	X, Y         float32
	PrevX, PrevY float32
	Speed        float32
	Idle         *anim.Clip
	CrookThrow   *anim.Clip
	A            anim.Animator
	Scale        float32
	HP           int
	Radius       float32
	InvulnTimer  float32
	HurtFlash    float32

	// 🔫 Стрельба
	Shots      []*Projectile
	CanShoot   bool
	FireTimer  float32
	FirePeriod float32

	Souls int

	Crook         *Crook
	CrookReady    bool
	CrookTimer    float32
	CrookCooldown float32

	Ult *Ultimate
}

func NewPlayer(assetsRoot string) (*Player, error) {
	clip, err := anim.LoadFromJSON(filepath.Join(assetsRoot, "textures/ghost/idle/anim.json"))
	crookThrow, _ := anim.LoadFromJSON(filepath.Join(assetsRoot, "textures", "ghost", "crook", "anim.json"))
	if err != nil {
		return nil, err
	}

	p := &Player{
		X: 200, Y: 300,
		Speed:      300,
		Idle:       clip,
		CrookThrow: crookThrow,
		Scale:      1.25,
		HP:         100,
		Radius:     18,
		// HurtFlash: 0, // по умолчанию

		CanShoot:   true,
		FirePeriod: 0.4,

		CrookReady:    true,
		CrookCooldown: 3.0,
	}
	p.Ult = NewUltimate(assetsRoot)

	p.PrevX, p.PrevY = p.X, p.Y
	p.A.Play(p.Idle, true)
	return p, nil
}

func (p *Player) Update(dt float32, camera *rl.Camera2D) {
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
		moveX *= 0.7071
		moveY *= 0.7071
	}

	// зеркалирование анимации
	if moveX > 0 {
		p.A.FlipX = true
	}
	if moveX < 0 {
		p.A.FlipX = false
	}

	p.X += moveX * p.Speed * dt
	p.Y += moveY * p.Speed * dt

	// ⏳ таймер стрельбы
	p.FireTimer -= dt

	// 🔫 стрельба на ПКМ
	p.FireTimer -= dt
	if p.CanShoot && rl.IsMouseButtonDown(rl.MouseRightButton) && p.FireTimer <= 0 {
		mouse := rl.GetMousePosition()
		world := rl.GetScreenToWorld2D(mouse, *camera)

		// Центр игрока
		f := p.A.Current.Frames[p.A.FrameIndex]
		centerX := p.X - float32(f.OrigX)*p.Scale + float32(f.Src.Width)*p.Scale/2
		centerY := p.Y - float32(f.OrigY)*p.Scale + float32(f.Src.Height)*p.Scale/2

		// Вектор направления от центра снаряда к курсору
		dx := world.X - centerX
		dy := world.Y - centerY

		shot := NewGhostBolt(centerX, centerY, dx, dy)
		p.Shots = append(p.Shots, shot)
		p.FireTimer = p.FirePeriod
	}

	// обновляем все снаряды
	out := p.Shots[:0]
	for _, s := range p.Shots {
		s.Update(dt)
		if s.Alive {
			out = append(out, s)
		}
	}
	p.Shots = out

	// таймеры игрока
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

	p.A.Update(dt)

	// если бросок проигрался — вернуть idle
	if p.A.Done() && p.A.CurrentClip() == p.CrookThrow {
		p.A.Play(p.Idle, true)
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

func (p *Player) Unload() {
	if p.Ult != nil {
		p.Ult.Unload()
	}
}

func (p *Player) Draw(camera rl.Camera2D) {
	tint := rl.White
	if p.HurtFlash > 0 {
		tint = rl.NewColor(255, 64, 64, 255)
	}

	// если есть текущий кадр — пересчитаем позицию для аниматора
	if p.A.Current != nil && p.A.FrameIndex < len(p.A.Current.Frames) {
		f := p.A.Current.Frames[p.A.FrameIndex]

		// D = center + Orig*scale - (Width*scale)/2
		drawX := p.X + float32(f.OrigX)*p.Scale - float32(f.Src.Width)*p.Scale/2
		drawY := p.Y + float32(f.OrigY)*p.Scale - float32(f.Src.Height)*p.Scale/2

		p.A.Draw(drawX, drawY, p.Scale, tint)
	} else {
		// запасной вариант
		p.A.Draw(p.X, p.Y, p.Scale, tint)
	}

	if p.Crook != nil && p.Crook.Active {
		p.Crook.Draw(p.X, p.Y)
	}

	if p.Ult != nil && p.Ult.flashActive {
		t := p.Ult.flashTime / p.Ult.flashDur
		if t > 1 {
			t = 1
		}

		// плавное затухание прозрачности
		alpha := uint8(180 * (1 - t))
		color := rl.NewColor(50, 255, 50, alpha)

		radius := 120 * p.Scale
		rl.DrawCircleV(rl.NewVector2(p.X, p.Y), radius, color)
	}

	// отрисовка снарядов с учётом камеры (пули в мировых координатах)
	rl.BeginMode2D(camera)
	for _, s := range p.Shots {
		s.Draw()
	}
	rl.EndMode2D()
}
