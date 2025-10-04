package main

import (
	"example.com/my2dgame/internal/entities"
	"example.com/my2dgame/internal/ui"
	"example.com/my2dgame/internal/world"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func fitCameraToWorld(cam *rl.Camera2D, worldW, worldH float32) {
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	minZoomX := sw / worldW
	minZoomY := sh / worldH
	minZoom := float32(math.Max(float64(minZoomX), float64(minZoomY)))
	if cam.Zoom < minZoom {
		cam.Zoom = minZoom
	}
}

func findAssets() string {
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 5; i++ {
		try := filepath.Join(dir, "assets")
		if st, err := os.Stat(try); err == nil && st.IsDir() {
			return try
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		try := filepath.Join(exeDir, "assets")
		if st, err := os.Stat(try); err == nil && st.IsDir() {
			return try
		}
	}
	return "assets"
}

type AppState int

const (
	StateMenu AppState = iota
	StateGame
	StatePause
)

// -------- UI --------
var uiFont rl.Font

const uiSize float32 = 28
const uiSpacing float32 = 1

type Button struct {
	Bounds rl.Rectangle
	Label  string
	Hot    bool
}

func (b *Button) Draw() {
	col := rl.NewColor(255, 255, 255, 200)
	border := rl.Black
	if b.Hot {
		col = rl.NewColor(255, 255, 255, 240)
		border = rl.DarkGray
	}
	rl.DrawRectangleRounded(b.Bounds, 0.2, 8, col)
	rl.DrawRectangleRoundedLines(b.Bounds, 0.2, 8, border)
	ts := rl.MeasureTextEx(uiFont, b.Label, uiSize, uiSpacing)
	x := b.Bounds.X + b.Bounds.Width/2 - ts.X/2
	y := b.Bounds.Y + b.Bounds.Height/2 - ts.Y/2
	rl.DrawTextEx(uiFont, b.Label, rl.NewVector2(x, y), uiSize, uiSpacing, rl.Black)
}

func segmentCircleHit(ax, ay, bx, by, cx, cy, r float32) bool {
	abx, aby := bx-ax, by-ay
	acx, acy := cx-ax, cy-ay
	ab2 := abx*abx + aby*aby
	var t float32 = 0
	if ab2 > 0 {
		t = (acx*abx + acy*aby) / ab2
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
	}
	px := ax + abx*t
	py := ay + aby*t
	dx := px - cx
	dy := py - cy
	return dx*dx+dy*dy <= r*r
}

// closest distance^2 between two line segments A(ax,ay)->B(bx,by) and C(cx,cy)->D(dx,dy)
// closest distance^2 between two line segments A(ax,ay)->B(bx,by) and C(cx,cy)->D(dx,dy)
func segSegDistSq(ax, ay, bx, by, cx, cy, dx, dy float32) float32 {
	// vectors
	ux, uy := bx-ax, by-ay
	vx, vy := dx-cx, dy-cy
	wx, wy := ax-cx, ay-cy

	a := ux*ux + uy*uy // |u|^2
	b := ux*vx + uy*vy // u·v
	c := vx*vx + vy*vy // |v|^2
	d := ux*wx + uy*wy // u·w
	e := vx*wx + vy*wy // v·w
	D := a*c - b*b

	var sN, sD = D, D
	var tN, tD = D, D

	if D < 1e-8 {
		// почти параллельны
		sN = 0
		sD = 1
		tN = e
		tD = c
	} else {
		sN = (b*e - c*d)
		tN = (a*e - b*d)
		// clamp sN to [0, sD]
		if sN < 0 {
			sN = 0
		} else if sN > sD {
			sN = sD
		}
	}

	// clamp tN to [0, tD] и корректировка sN при необходимости
	if tN < 0 {
		tN = 0
		if -d < 0 {
			sN = 0
			sD = 1
		} else if -d > a {
			sN = sD
		} else {
			sN = -d
			sD = a
		}
	} else if tN > tD {
		tN = tD
		if (-d + b) < 0 {
			sN = 0
			sD = 1
		} else if (-d + b) > a {
			sN = sD
		} else {
			sN = (-d + b)
			sD = a
		}
	}

	// параметры на отрезках
	var sc float32
	if sD != 0 {
		sc = sN / sD
	}
	var tc float32
	if tD != 0 {
		tc = tN / tD
	}

	// ближайшие точки
	px := ax + sc*ux
	py := ay + sc*uy
	qx := cx + tc*vx
	qy := cy + tc*vy

	dx_ := px - qx
	dy_ := py - qy
	return dx_*dx_ + dy_*dy_
}

func main() {
	// Полноэкранный старт
	mon := rl.GetCurrentMonitor()
	W := int32(rl.GetMonitorWidth(mon))
	H := int32(rl.GetMonitorHeight(mon))
	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.InitWindow(W, H, "My 2D Game — Menu + Game + Music")
	rl.ToggleFullscreen()

	rl.SetExitKey(rl.KeyNull)

	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	rand.Seed(time.Now().UnixNano())
	bg := rl.NewColor(240, 243, 248, 255)
	assetsRoot := findAssets()

	// Мир (фон-карта). Если мылится — уменьшай scale.
	wrld, err := world.LoadBackdrop(assetsRoot, filepath.Join("textures", "maps", "village.png"), 3.0)
	if err != nil {
		fmt.Println("world:", err)
		return
	}
	defer wrld.Unload()

	// Проектайлы
	if err := entities.LoadProjectileAssets(assetsRoot); err != nil {
		fmt.Println("projectiles:", err)
	}

	// Шрифт
	charset := func() []int32 {
		cps := make([]int32, 0, 1024)
		add := func(a, b int32) {
			for cp := a; cp <= b; cp++ {
				cps = append(cps, cp)
			}
		}
		add(0x0020, 0x007E)
		add(0x0400, 0x04FF)
		add(0x0500, 0x052F)
		add(0x2DE0, 0x2DFF)
		add(0xA640, 0xA69F)
		add(0x2010, 0x205E)
		return cps
	}()
	hud, err := ui.LoadHealthHUD(assetsRoot, 20, 60, 1.9) // позиция (20,60), масштаб 1.0
	if err != nil {
		fmt.Println("health hud:", err)
	} // не фейлим игру, просто лог
	defer func() {
		if hud != nil {
			hud.Unload()
		}
	}()

	fontPath := filepath.Join(assetsRoot, "fonts", "NotoSans-Regular.ttf")
	uiFont = rl.LoadFontEx(fontPath, int32(48), charset) // если нужно: , int32(len(charset))
	rl.SetTextureFilter(uiFont.Texture, rl.FilterBilinear)
	defer rl.UnloadFont(uiFont)

	// Фон меню
	var menuBG rl.Texture2D
	if img := rl.LoadImage(filepath.Join(assetsRoot, "ui", "menu_bg.png")); img.Data != nil {
		menuBG = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
	}
	defer func() {
		if menuBG.ID != 0 {
			rl.UnloadTexture(menuBG)
		}
	}()

	// Аудио
	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
	var (
		menuMusic, gameMusic       rl.Music
		hasMenuMusic, hasGameMusic bool
	)
	if _, err := os.Stat(filepath.Join(assetsRoot, "music", "menu.mp3")); err == nil {
		menuMusic = rl.LoadMusicStream(filepath.Join(assetsRoot, "music", "menu.mp3"))
		menuMusic.Looping = true
		rl.SetMusicVolume(menuMusic, 0.7)
		rl.PlayMusicStream(menuMusic)
		hasMenuMusic = true
	}
	if _, err := os.Stat(filepath.Join(assetsRoot, "music", "game.mp3")); err == nil {
		gameMusic = rl.LoadMusicStream(filepath.Join(assetsRoot, "music", "game.mp3"))
		gameMusic.Looping = true
		rl.SetMusicVolume(gameMusic, 0.6)
		hasGameMusic = true
	}
	defer func() {
		if hasMenuMusic {
			rl.UnloadMusicStream(menuMusic)
		}
		if hasGameMusic {
			rl.UnloadMusicStream(gameMusic)
		}
	}()

	// Кнопки
	btnPlay := Button{Label: "Играть"}
	btnExit := Button{Label: "Выйти"}
	btnResume := Button{Label: "Продолжить"}
	btnToMenu := Button{Label: "Выйти в меню"}

	state := StateMenu

	// --- ИГРА ---
	var (
		player     *entities.Player
		enemies    []*entities.Enemy
		spawnT     float32
		spawnEvery = float32(2.0)
		cam        rl.Camera2D
	)

	startGame := func() {
		p, err := entities.NewPlayer(assetsRoot)
		if err != nil {
			fmt.Println("player load:", err)
			return
		}

		// центр спавна
		wpx, hpx := wrld.SizePx()
		p.X, p.Y = wpx*0.5, hpx*0.5

		player = p
		enemies = make([]*entities.Enemy, 0, 64)
		spawnT = 0
		cam = rl.Camera2D{
			Target: rl.NewVector2(player.X, player.Y),
			Offset: rl.NewVector2(float32(rl.GetScreenWidth())/2, float32(rl.GetScreenHeight())/2),
			Zoom:   1.0,
		}
		// гарантируем, что экран помещается в мире
		fitCameraToWorld(&cam, wpx, hpx)

		// музыка
		if hasMenuMusic {
			rl.StopMusicStream(menuMusic)
		}
		if hasGameMusic {
			rl.PlayMusicStream(gameMusic)
		}
		state = StateGame
	}

	spawnEnemy := func() {
		r := float32(600)
		ang := rand.Float64() * 2 * math.Pi
		sx := player.X + r*float32(math.Cos(ang))
		sy := player.Y + r*float32(math.Sin(ang))

		kind := "melee"
		speed := float32(80)
		scale := float32(1.2)
		if rand.Intn(2) == 0 {
			kind = "slime"
			speed = 70
			scale = 1.2
		}

		if e, err := entities.NewEnemyKind(assetsRoot, kind, sx, sy, speed, scale); err == nil {
			enemies = append(enemies, e)
		} else {
			fmt.Println("enemy load:", err)
		}
	}

	for !rl.WindowShouldClose() {
		dt := float32(rl.GetFrameTime())

		if rl.IsKeyPressed(rl.KeyF11) {
			rl.ToggleFullscreen()
		}

		if hasMenuMusic {
			rl.UpdateMusicStream(menuMusic)
		}
		if hasGameMusic {
			rl.UpdateMusicStream(gameMusic)
		}

		rl.BeginDrawing()
		rl.ClearBackground(bg)

		switch state {
		case StateMenu:
			if menuBG.ID != 0 {
				src := rl.NewRectangle(0, 0, float32(menuBG.Width), float32(menuBG.Height))
				dst := rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
				rl.DrawTexturePro(menuBG, src, dst, rl.NewVector2(0, 0), 0, rl.White)
			} else {
				rl.ClearBackground(rl.DarkGreen)
			}

			bw, bh := float32(320), float32(70)
			centerX := float32(rl.GetScreenWidth()) * 0.5
			startY := float32(rl.GetScreenHeight())*0.6 - bh
			spacing := float32(20)

			btnPlay.Bounds = rl.NewRectangle(centerX-bw/2, startY, bw, bh)
			btnExit.Bounds = rl.NewRectangle(centerX-bw/2, startY+bh+spacing, bw, bh)

			mx, my := float32(rl.GetMouseX()), float32(rl.GetMouseY())
			btnPlay.Hot = rl.CheckCollisionPointRec(rl.NewVector2(mx, my), btnPlay.Bounds)
			btnExit.Hot = rl.CheckCollisionPointRec(rl.NewVector2(mx, my), btnExit.Bounds)

			if rl.IsMouseButtonPressed(rl.MouseLeftButton) || rl.IsKeyPressed(rl.KeyEnter) {
				if btnPlay.Hot || rl.IsKeyPressed(rl.KeyEnter) {
					startGame()
				}
			}
			if (rl.IsMouseButtonPressed(rl.MouseLeftButton) && btnExit.Hot) || rl.IsKeyPressed(rl.KeyEscape) {
				if hasMenuMusic {
					rl.StopMusicStream(menuMusic)
				}
				if hasGameMusic {
					rl.StopMusicStream(gameMusic)
				}
				rl.EndDrawing()
				return
			}

			title := "666adididas"
			ts := rl.MeasureTextEx(uiFont, title, 48, uiSpacing)
			tx := float32(rl.GetScreenWidth())*0.5 - ts.X*0.5
			ty := float32(rl.GetScreenHeight()) * 0.25
			rl.DrawTextEx(uiFont, title, rl.NewVector2(tx, ty), 48, uiSpacing, rl.White)

			btnPlay.Draw()
			btnExit.Draw()

			hint := "Enter — Играть, Esc — Выйти"
			hs := rl.MeasureTextEx(uiFont, hint, 20, uiSpacing)
			rl.DrawTextEx(uiFont, hint, rl.NewVector2(20, float32(rl.GetScreenHeight())-hs.Y-20), 20, uiSpacing, rl.White)

		case StateGame:
			// Зум колесом + ограничение, чтобы мир не был уже экрана
			cam.Zoom += rl.GetMouseWheelMove() * 0.05
			if cam.Zoom < 0.3 {
				cam.Zoom = 0.3
			}
			if cam.Zoom > 3.0 {
				cam.Zoom = 3.0
			}
			wpx, hpx := wrld.SizePx()
			fitCameraToWorld(&cam, wpx, hpx)

			// Update
			player.Update(dt)
			player.X, player.Y = wrld.Clamp(player.X, player.Y)

			for _, e := range enemies {
				e.X, e.Y = wrld.Clamp(e.X, e.Y)
				if e.CanShoot {
					out := e.Shots[:0]
					for _, p := range e.Shots {
						if p.X < 0 || p.Y < 0 || p.X > wpx || p.Y > hpx {
							p.Alive = false
						}
						if p.Alive {
							out = append(out, p)
						}
					}
					e.Shots = out
				}
			}

			// --- УРОН ---
			// 1) Пули во врагах уже обновлены; проверим попадание по игроку
			for _, e := range enemies {
				if e.CanShoot {
					for _, p := range e.Shots {
						if !p.Alive {
							continue
						}
						// расстояние между отрезком траектории пули и отрезком траектории игрока за кадр
						r := player.Radius + p.HitRadius
						dist2 := segSegDistSq(
							p.PrevX, p.PrevY, p.X, p.Y,
							player.PrevX, player.PrevY, player.X, player.Y,
						)
						if dist2 <= r*r {
							p.Alive = false
							player.TakeDamage(10)
						}
					}
				}
			}

			// 2) Контактный урон ближника
			for _, e := range enemies {
				if e.Kind == "melee" {
					dx := e.X - player.X
					dy := e.Y - player.Y
					r := player.Radius + e.MeleeRange
					if dx*dx+dy*dy <= r*r {
						// удар, если таймер атаки врага готов и игрок не в инвулне
						if e.AttackTimer <= 0 && player.InvulnTimer <= 0 {
							player.TakeDamage(e.ContactDamage)
							e.AttackTimer = e.AttackCD
						}
					}
				}
			}

			// 3) Если здоровье закончилось — простая «смерть» -> выход в меню
			if player.HP <= 0 {
				// стоп игровую музыку, включим меню
				if hasGameMusic {
					rl.StopMusicStream(gameMusic)
				}
				if hasMenuMusic {
					rl.PlayMusicStream(menuMusic)
				}
				state = StateMenu
				// можно тут же очистить врагов
				enemies = enemies[:0]
				continue
			}

			spawnT -= dt
			if spawnT <= 0 {
				spawnEnemy()
				spawnT = spawnEvery
			}
			for _, e := range enemies {
				e.Update(dt, player.X, player.Y)
			}

			// Камера
			cam.Target = rl.NewVector2(player.X, player.Y)
			halfW := (float32(rl.GetScreenWidth()) / 2) / cam.Zoom
			halfH := (float32(rl.GetScreenHeight()) / 2) / cam.Zoom
			cx, cy := cam.Target.X, cam.Target.Y
			if cx < halfW {
				cx = halfW
			}
			if cy < halfH {
				cy = halfH
			}
			if cx > wpx-halfW {
				cx = wpx - halfW
			}
			if cy > hpx-halfH {
				cy = hpx - halfH
			}
			cam.Target = rl.NewVector2(cx, cy)

			// Рисование мира и объектов
			rl.BeginMode2D(cam)
			wrld.Draw(cam)
			for _, e := range enemies {
				e.Draw()
			}
			player.Draw()
			rl.EndMode2D()

			// HUD
			if hud != nil {
				hud.Draw(player.HP)
			}

			helpText := "Move: WASD/Arrows  |  Fullscreen: F11  |  Zoom: Wheel  |  Esc: пауза"
			rl.DrawTextEx(uiFont, helpText, rl.NewVector2(20, 20), 20, uiSpacing, rl.DarkGray)

			rl.DrawTextEx(uiFont, helpText, rl.NewVector2(20, 20), 20, uiSpacing, rl.DarkGray)
			rl.DrawFPS(int32(rl.GetScreenWidth())-90, 10)

			// Пауза
			if rl.IsKeyPressed(rl.KeyEscape) {
				if hasGameMusic {
					rl.PauseMusicStream(gameMusic)
				}
				state = StatePause
			}

		case StatePause:
			// Фон замороженной игры
			rl.BeginMode2D(cam)
			wrld.Draw(cam)
			for _, e := range enemies {
				e.Draw()
			}
			if player != nil {
				player.Draw()
			}
			rl.EndMode2D()

			// Вуаль
			rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()),
				rl.NewColor(0, 0, 0, 160))
			// Заголовок
			title := "Пауза"
			ts := rl.MeasureTextEx(uiFont, title, 48, uiSpacing)
			tx := float32(rl.GetScreenWidth())*0.5 - ts.X*0.5
			ty := float32(rl.GetScreenHeight()) * 0.28
			rl.DrawTextEx(uiFont, title, rl.NewVector2(tx, ty), 48, uiSpacing, rl.White)

			// Кнопки
			bw, bh := float32(360), float32(74)
			centerX := float32(rl.GetScreenWidth()) * 0.5
			startY := float32(rl.GetScreenHeight())*0.45 - bh
			spacing := float32(22)
			btnResume.Bounds = rl.NewRectangle(centerX-bw/2, startY, bw, bh)
			btnToMenu.Bounds = rl.NewRectangle(centerX-bw/2, startY+bh+spacing, bw, bh)

			mx, my := float32(rl.GetMouseX()), float32(rl.GetMouseY())
			btnResume.Hot = rl.CheckCollisionPointRec(rl.NewVector2(mx, my), btnResume.Bounds)
			btnToMenu.Hot = rl.CheckCollisionPointRec(rl.NewVector2(mx, my), btnToMenu.Bounds)

			btnResume.Draw()
			btnToMenu.Draw()

			// Управление
			if rl.IsKeyPressed(rl.KeyEscape) || rl.IsKeyPressed(rl.KeyEnter) ||
				(rl.IsMouseButtonPressed(rl.MouseLeftButton) && btnResume.Hot) {
				if hasGameMusic {
					rl.ResumeMusicStream(gameMusic)
				}
				state = StateGame
			}
			if rl.IsMouseButtonPressed(rl.MouseLeftButton) && btnToMenu.Hot {
				if hasGameMusic {
					rl.StopMusicStream(gameMusic)
				}
				if hasMenuMusic {
					rl.PlayMusicStream(menuMusic)
				}
				state = StateMenu
			}

			hint := "Enter/Esc — продолжить, ЛКМ — выбрать"
			hs := rl.MeasureTextEx(uiFont, hint, 20, uiSpacing)
			rl.DrawTextEx(uiFont, hint, rl.NewVector2(20, float32(rl.GetScreenHeight())-hs.Y-20), 20, uiSpacing, rl.White)
		}

		rl.EndDrawing()
	}
}
