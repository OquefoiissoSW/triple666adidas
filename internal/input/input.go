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
}

var Current State

const (
  gpA            = 0
  gpB            = 1
  gpX            = 2
  gpRightTrigger = 7
  gpStart        = 8
  gpAxisLeftX    = rl.GamepadAxisLeftX
  gpAxisLeftY    = rl.GamepadAxisLeftY
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

  if gamepad {
    // Аналоговые оси движения с геймпада
    Current.MoveX = rl.GetGamepadAxisMovement(0, gpAxisLeftX)
    Current.MoveY = rl.GetGamepadAxisMovement(0, gpAxisLeftY)

    // Булевы направления с геймпада (если осей мало или нет)
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

    // Действия с геймпада
    Current.Shoot = rl.IsGamepadButtonDown(0, gpRightTrigger)
    Current.Hook = rl.IsGamepadButtonDown(0, gpX)
    Current.Pause = rl.IsGamepadButtonDown(0, gpStart)
    Current.Confirm = rl.IsGamepadButtonDown(0, gpA)
    Current.Cancel = rl.IsGamepadButtonDown(0, gpB)

  } else {
    // Движение с клавиатуры
    if rl.IsKeyDown(rl.KeyA)  rl.IsKeyDown(rl.KeyLeft) {
      Current.MoveLeft = true
    }
    if rl.IsKeyDown(rl.KeyD)  rl.IsKeyDown(rl.KeyRight) {
      Current.MoveRight = true
    }
    if rl.IsKeyDown(rl.KeyW)  rl.IsKeyDown(rl.KeyUp) {
      Current.MoveUp = true
    }
    if rl.IsKeyDown(rl.KeyS)  rl.IsKeyDown(rl.KeyDown) {
      Current.MoveDown = true
    }

    // Аналоговых осей нет для клавиатуры
    Current.MoveX = 0
    Current.MoveY = 0

    // Действия с клавиатуры и мыши
    Current.Shoot = rl.IsMouseButtonDown(rl.MouseButtonRight)
    Current.Hook = rl.IsKeyDown(rl.KeyQ)
    Current.Pause = rl.IsKeyDown(rl.KeyEscape)
    Current.Confirm = rl.IsKeyDown(rl.KeyEnter)
    Current.Cancel = rl.IsKeyDown(rl.KeyEscape)
  }
}