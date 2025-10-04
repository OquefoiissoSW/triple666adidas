package entities

import (
	"math"
	"math/rand"
	"path/filepath"

	"example.com/my2dgame/internal/anim"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Soul struct {
	X, Y         float32
	BaseX, BaseY float32
	Radius       float32
	Angle        float32
	Speed        float32
	ExpandSpeed  float32
	Alive        bool
	Time         float32
	MaxTime      float32

	Anim   anim.Animator
	Clip   *anim.Clip
	Scale  float32
	RotDeg float32

	// --- новое ---
	IsAbsorbing bool    // флаг: душа летит к игроку
	Alpha       float32 // прозрачность (0–1)
}

func NewSoul(assetsRoot string, x, y float32) (*Soul, error) {
	jsonPath := filepath.Join(assetsRoot, "textures", "soul", "anim.json")
	clip, err := anim.LoadFromJSON(jsonPath)
	if err != nil {
		return nil, err
	}

	dir := rand.Float32() * 2 * math.Pi

	s := &Soul{
		X: x, Y: y,
		BaseX: x, BaseY: y,
		Radius:      0,
		Angle:       dir,
		Speed:       2 + rand.Float32()*2,   // вращение
		ExpandSpeed: 15 + rand.Float32()*20, // радиальный рост
		Alive:       true,
		Scale:       2.0,
		MaxTime:     30.0 + rand.Float32()*1.5,
		Clip:        clip,
		Alpha:       1.0,
	}

	s.Anim.Play(s.Clip, true)
	return s, nil
}

func (s *Soul) Update(dt float32, playerX, playerY float32) {
	if !s.Alive {
		return
	}

	// Если душа поглощается игроком — летим к нему
	if s.IsAbsorbing {
		dx := playerX - s.X
		dy := playerY - s.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		// скорость притягивания
		speed := float32(600.0) * dt
		if dist > speed {
			s.X += dx / dist * speed
			s.Y += dy / dist * speed
		} else {
			s.X, s.Y = playerX, playerY
		}

		// делаем прозрачнее
		s.Alpha -= dt * 4.0 // исчезнет примерно за 0.25 с
		if s.Alpha <= 0 {
			s.Alive = false
		}

		return
	}

	// === обычное движение по спирали ===
	s.Time += dt
	if s.Time > s.MaxTime {
		s.Alive = false
		return
	}

	spiralSlow := float32(math.Exp(-float64(s.Radius) * 0.015))
	s.Angle += s.Speed * spiralSlow * dt

	targetRadius := 60 + 80*float32(math.Sin(float64(s.Time*0.5)))
	s.Radius += (targetRadius - s.Radius) * 0.5 * dt

	s.X = s.BaseX + float32(math.Cos(float64(s.Angle)))*s.Radius
	s.Y = s.BaseY + float32(math.Sin(float64(s.Angle)))*s.Radius*0.4 +
		5*float32(math.Sin(float64(s.Time*2)))

	dx := float64(math.Cos(float64(s.Angle)))
	dy := float64(math.Sin(float64(s.Angle)) * 0.4)
	baseAngle := math.Atan2(dy, dx)
	wobble := math.Sin(float64(s.Time*6)) * (15 * math.Pi / 180)
	s.RotDeg = float32((baseAngle+wobble)*180/math.Pi) + 180

	s.Anim.Update(dt)
}

func (s *Soul) Draw() {
	if !s.Alive {
		return
	}

	a := s.Alpha
	if a < 0 {
		a = 0
	}
	alpha := uint8(255 * a)
	color := rl.Color{R: 255, G: 255, B: 255, A: alpha}

	if s.Anim.DrawRotated != nil {
		s.Anim.DrawRotated(s.X, s.Y, s.Scale, s.RotDeg, color)
	} else {
		s.Anim.Draw(s.X, s.Y, s.Scale, color)
	}
}
