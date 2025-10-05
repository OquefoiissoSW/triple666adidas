package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {
	MoveX, MoveY        float32
	MoveLeft, MoveRight bool
	MoveUp, MoveDown    bool
	Shoot, Hook, Pause  bool
	Confirm, Cancel     bool
	Fullscreen, Ulta    bool
}

var Current State

// Используем константы rl для геймпада
const (
	gpA            = rl.GamepadButtonLeftFaceDown  // A
	gpB            = rl.GamepadButtonLeftFaceRight // B
	gpX            = rl.GamepadButtonLeftFaceLeft  // X
	gpY            = rl.GamepadButtonRightFaceUp   // Y
	gpRightTrigger = rl.GamepadButtonRightTrigger1 // R1
	gpLeftTrigger  = rl.GamepadButtonLeftTrigger1  // L1
	gpStart        = 15                            // Start

	gpAxisLeftX = rl.GamepadAxisLeftX
	gpAxisLeftY = rl.GamepadAxisLeftY
)

func Init() {
	// Пока не требуется дополнительная инициализация
}

func Update() {
	gamepad := rl.IsGamepadAvailable(0)

	// Сброс состояния движения
	Current.MoveLeft = false
	Current.MoveRight = false
	Current.MoveUp = false
	Current.MoveDown = false
	Current.MoveX = 0
	Current.MoveY = 0
	Current.Shoot = false
	Current.Hook = false
	Current.Pause = false
	Current.Confirm = false
	Current.Cancel = false
	Current.Fullscreen = false
	Current.Ulta = false

	if gamepad {
		// Аналоговые оси движения с геймпада
		Current.MoveX = rl.GetGamepadAxisMovement(0, gpAxisLeftX)
		Current.MoveY = rl.GetGamepadAxisMovement(0, gpAxisLeftY)

		// Булевы направления с геймпада
		if Current.MoveX < -0.5 {
			Current.MoveLeft = true
		}
		if Current.MoveX > 0.5 {
			Current.MoveRight = true
		}
		if Current.MoveY < -0.5 {
			Current.MoveUp = true
		}
		if Current.MoveY > 0.5 {
			Current.MoveDown = true
		}

		// Действия с геймпада (только при нажатии)
		Current.Shoot = rl.IsGamepadButtonPressed(0, gpRightTrigger)
		Current.Hook = rl.IsGamepadButtonPressed(0, gpX)
		Current.Pause = rl.IsGamepadButtonPressed(0, gpStart)
		Current.Confirm = rl.IsGamepadButtonPressed(0, gpA)
		Current.Cancel = rl.IsGamepadButtonPressed(0, gpB)
		Current.Fullscreen = rl.IsGamepadButtonPressed(0, gpLeftTrigger)
		Current.Ulta = rl.IsGamepadButtonDown(0, rl.GamepadButtonRightFaceUp)

	} else {
		// Движение с клавиатуры
		if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
			Current.MoveLeft = true
		}
		if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
			Current.MoveRight = true
		}
		if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
			Current.MoveUp = true
		}
		if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
			Current.MoveDown = true
		}

		// Действия с клавиатуры и мыши (только при нажатии)
		Current.Shoot = rl.IsMouseButtonPressed(rl.MouseButtonRight)
		Current.Hook = rl.IsKeyPressed(rl.KeyQ)
		Current.Pause = rl.IsKeyPressed(rl.KeyEscape)
		Current.Confirm = rl.IsKeyPressed(rl.KeyEnter)
		Current.Cancel = rl.IsKeyPressed(rl.KeyEscape)
		Current.Fullscreen = rl.IsKeyPressed(rl.KeyF11)
		Current.Ulta = rl.IsKeyPressed(rl.KeyE)
	}
}
