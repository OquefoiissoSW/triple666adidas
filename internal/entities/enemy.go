package entities

import (
	"math"
	"path/filepath"

	"example.com/my2dgame/internal/anim"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Enemy struct {
	X, Y       float32
	Speed      float32
	Scale      float32
	Anim       anim.Animator
	Idle       *anim.Clip
	Alive      bool
	Kind       string
	FacesRight bool

	// ---- боевые характеристики ----
	HP            int // новое поле
	CanShoot      bool
	FireTimer     float32
	FirePeriod    float32
	FireRange     float32
	Shots         []*Projectile
	MeleeRange    float32
	AttackCD      float32
	AttackTimer   float32
	ContactDamage int
}

func NewEnemyKind(assetsRoot, kind string, x, y, speed, scale float32) (*Enemy, error) {
	jsonPath := filepath.Join(assetsRoot, "textures", kind, "idle", "anim.json")
	clip, err := anim.LoadFromJSON(jsonPath)
	if err != nil {
		return nil, err
	}

	e := &Enemy{
		X: x, Y: y,
		Speed: speed,
		Scale: scale,
		Idle:  clip,
		Alive: true,
		Kind:  kind,

		HP: 50,

		FacesRight:    false,           // базовый кадр смотрит влево
		CanShoot:      kind != "melee", // <-- только не-melee
		FirePeriod:    1.5,
		FireRange:     600,
		MeleeRange:    28,  // подгони под размер спрайта
		AttackCD:      0.8, // минимум 0.8с между ударами
		AttackTimer:   0,
		ContactDamage: 10,
	}
	e.Anim.Play(e.Idle, true)

	return e, nil
}

func NewEnemy(assetsRoot string, x, y float32) (*Enemy, error) {
	return NewEnemyKind(assetsRoot, "melee", x, y, 80, 2.25)
}

func (e *Enemy) Update(dt float32, targetX, targetY float32) {
	e.AttackTimer -= dt
	if e.AttackTimer < 0 {
		e.AttackTimer = 0
	}

	if !e.Alive {
		return
	}

	dx := targetX - e.X
	dy := targetY - e.Y
	dist := float32(math.Hypot(float64(dx), float64(dy)))

	// движение к цели (как раньше)
	stopDist := float32(16)
	if dist > 0.001 && dist > stopDist {
		nx := dx / dist
		ny := dy / dist
		e.X += nx * e.Speed * dt
		e.Y += ny * e.Speed * dt

		// корректный флип (базово смотрит влево)
		if !e.FacesRight {
			if nx > 0 {
				e.Anim.FlipX = true
			} else if nx < 0 {
				e.Anim.FlipX = false
			}
		} else {
			if nx < 0 {
				e.Anim.FlipX = true
			} else if nx > 0 {
				e.Anim.FlipX = false
			}
		}
	}

	if e.CanShoot {
		e.FireTimer -= dt
		if dist <= e.FireRange && e.FireTimer <= 0 {
			e.Shots = append(e.Shots, NewSlimeBolt(e.X, e.Y, dx, dy))
			e.FireTimer = e.FirePeriod
		}
		// апдейт пуль и очистка мёртвых
		out := e.Shots[:0]
		for _, p := range e.Shots {
			p.Update(dt)
			if p.Alive {
				out = append(out, p)
			}
		}
		e.Shots = out
	}

	e.Anim.Update(dt)
}

func (e *Enemy) Draw() {
	if !e.Alive {
		return
	}
	e.Anim.Draw(e.X, e.Y, e.Scale, rl.White)

	if e.CanShoot {
		for _, p := range e.Shots {
			p.Draw()
		}
	}
}

func (e *Enemy) TakeDamage(dmg int) {
	if !e.Alive || dmg <= 0 {
		return
	}
	e.HP -= dmg
	if e.HP <= 0 {
		e.HP = 0
		e.Alive = false
	}
}
