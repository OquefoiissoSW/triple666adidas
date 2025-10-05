package entities

import (
	"math"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// --- Состояния крюка ---
type CrookState int

const (
	CrookIdle CrookState = iota
	CrookForward
	CrookReturning
)

type Crook struct {
	X, Y           float32
	StartX, StartY float32
	DirX, DirY     float32

	Speed   float32
	MaxDist float32
	State   CrookState
	Active  bool
	HitSoul *Soul // если зацепили душу

	// визуальные данные
	Tex    rl.Texture2D
	RotDeg float32
	Scale  float32

	SndThrow rl.Sound
	SndHit   rl.Sound
}

// Создание нового крюка
func NewCrook(assetsRoot string, playerX, playerY, targetX, targetY float32) *Crook {
	dx := targetX - playerX
	dy := targetY - playerY
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if dist != 0 {
		dx /= dist
		dy /= dist
	}

	texPath := filepath.Join(assetsRoot, "textures", "crook", "crook.png")
	tex := rl.LoadTexture(texPath)

	sndThrow := rl.LoadSound(filepath.Join(assetsRoot, "sounds", "crook.mp3"))
	sndHit := rl.LoadSound(filepath.Join(assetsRoot, "sounds", "headshot.mp3"))

	return &Crook{
		X:        playerX,
		Y:        playerY,
		StartX:   playerX,
		StartY:   playerY,
		DirX:     dx,
		DirY:     dy,
		Speed:    900,
		MaxDist:  400,
		State:    CrookForward,
		Active:   true,
		Tex:      tex,
		Scale:    2.0,
		SndThrow: sndThrow,
		SndHit:   sndHit,
	}
}

// Обновление крюка
func (c *Crook) Update(dt float32, playerX, playerY float32, souls []*Soul) {
	if !c.Active {
		return
	}

	switch c.State {
	case CrookForward:
		// движение вперёд
		c.X += c.DirX * c.Speed * dt
		c.Y += c.DirY * c.Speed * dt

		// обновляем угол (вверх = "нос" спрайта)
		c.RotDeg = float32(math.Atan2(float64(c.DirY), float64(c.DirX))*180/math.Pi) + 90

		// достигли максимальной дистанции
		dist := float32(math.Sqrt(float64((c.X-c.StartX)*(c.X-c.StartX) + (c.Y-c.StartY)*(c.Y-c.StartY))))
		if dist >= c.MaxDist {
			c.State = CrookReturning
		}

		// проверяем попадание в душу
		for _, s := range souls {
			if !s.Alive || s.IsAbsorbing {
				continue
			}
			dx := s.X - c.X
			dy := s.Y - c.Y
			if dx*dx+dy*dy < 25*25 { // радиус зацепа
				s.IsAbsorbing = true
				c.HitSoul = s
				c.State = CrookReturning
				rl.PlaySound(c.SndHit)
				break
			}
		}

	case CrookReturning:
		// направление обратно к игроку
		dx := playerX - c.X
		dy := playerY - c.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist > 0 {
			c.X += dx / dist * c.Speed * dt
			c.Y += dy / dist * c.Speed * dt
		}

		// обновляем угол (нос направлен к игроку)
		c.RotDeg = float32(math.Atan2(float64(dy), float64(dx))*180/math.Pi) - 90

		// если душа зацеплена — тянем её за крюком
		if c.HitSoul != nil && c.HitSoul.Alive {
			c.HitSoul.X = c.X
			c.HitSoul.Y = c.Y
			if dist < 30 {
				c.HitSoul.IsAbsorbing = true
				c.Active = false
				return
			}
		}

		if dist < 20 {
			c.Active = false
		}
	}
}

// Отрисовка
func (c *Crook) Draw(playerX, playerY float32) {
	if !c.Active {
		return
	}

	// Рисуем линию (верёвку)
	rl.DrawLineEx(
		rl.NewVector2(playerX, playerY),
		rl.NewVector2(c.X, c.Y),
		2,
		rl.NewColor(255, 230, 120, 200),
	)

	// Центр крюка
	src := rl.NewRectangle(0, 0, float32(c.Tex.Width), float32(c.Tex.Height))
	dest := rl.NewRectangle(c.X, c.Y, float32(c.Tex.Width)*c.Scale, float32(c.Tex.Height)*c.Scale)
	origin := rl.NewVector2(float32(c.Tex.Width)*c.Scale/2, float32(c.Tex.Height)*c.Scale) // нижняя точка = "хвост"

	rl.DrawTexturePro(c.Tex, src, dest, origin, c.RotDeg, rl.White)
}

// Освобождение текстуры при завершении
func (c *Crook) Unload() {
	if c.Tex.ID != 0 {
		rl.UnloadTexture(c.Tex)
	}
}
