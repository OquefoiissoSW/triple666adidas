package entities

import (
	"math"
	"path/filepath"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Ultimate struct {
	Charge      int     // 0–3 (всего 3 заряда)
	MaxCharge   int     // всегда 3
	Active      bool    // активна ли сейчас
	Duration    float32 // сколько секунд длится эффект
	Timer       float32 // таймер эффекта
	Sound       rl.Sound
	LastUsed    time.Time
	FreezeRange float32

	partialSouls int

	flashActive bool
	flashTime   float32
	flashDur    float32
}

// Создание ульты
func NewUltimate(assetsRoot string) *Ultimate {
	snd := rl.LoadSound(filepath.Join(assetsRoot, "sounds", "stop.mp3"))
	rl.SetSoundVolume(snd, 0.8)
	return &Ultimate{
		MaxCharge:   3,
		Charge:      0,
		Duration:    3.5,   // 3.5 сек заморозки
		FreezeRange: 700.0, // радиус действия
		Sound:       snd,
	}
}

func (u *Ultimate) Unload() {
	if u.Sound.FrameCount > 0 {
		rl.UnloadSound(u.Sound)
	}
}

// Пополнение заряда (2 души = +1 заряд)
func (u *Ultimate) AddSouls(numSouls int) {
	if numSouls <= 0 {
		return
	}
	u.partialSouls += numSouls // новая переменная
	for u.partialSouls >= 2 {
		u.Charge++
		u.partialSouls -= 2
		if u.Charge > u.MaxCharge {
			u.Charge = u.MaxCharge
			u.partialSouls = 0
			break
		}
	}
}

// Обновление (в каждом кадре)
func (u *Ultimate) Update(dt float32, enemies []*Enemy) {
	if !u.Active {
		return
	}

	u.Timer -= dt
	if u.Timer <= 0 {
		u.Active = false
		// размораживаем врагов
		for _, e := range enemies {
			if !e.Alive {
				continue
			}
			e.Speed = e.BaseSpeed
			e.CanShoot = e.Kind != "melee"
		}
	}

	if u.flashActive {
		u.flashTime += dt
		if u.flashTime >= u.flashDur {
			u.flashActive = false
		}
	}
}

// Активация ульты (при нажатии E)
func (u *Ultimate) TryActivate(player *Player, enemies []*Enemy) {
	if u.Active || u.Charge < u.MaxCharge {
		return
	}

	u.Active = true
	u.Timer = u.Duration
	u.Charge = 0

	u.flashActive = true
	u.flashTime = 0
	u.flashDur = 0.4 // длительность вспышки ~0.4 сек

	rl.PlaySound(u.Sound)

	// Заморозим врагов
	for _, e := range enemies {
		if !e.Alive {
			continue
		}
		dx := e.X - player.X
		dy := e.Y - player.Y
		if math.Hypot(float64(dx), float64(dy)) <= float64(u.FreezeRange) {
			e.Speed = 0
			e.CanShoot = false
			e.FreezeTimer = u.Duration
		}
		// очистим летящие снаряды
		if len(e.Shots) > 0 {
			e.Shots = e.Shots[:0]
		}
	}
}
