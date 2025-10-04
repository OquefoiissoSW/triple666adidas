package main

import (
	"fmt"
	"log"
	"github.com/gen2brain/raylib-go/raylib"
)

func main() {
	fmt.Println("Программа запускается...")
	
	rl.SetTraceLogLevel(rl.LogAll) // Включаем все логи raylib
	rl.InitWindow(800, 450, "Debug Window")
	
	if !rl.IsWindowReady() {
		log.Fatal("ОШИБКА: Окно не инициализировано!")
	}
	
	fmt.Println("Окно создано успешно!")
	
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.DrawText("Если вы это видите - все работает!", 10, 10, 20, rl.Black)
		rl.EndDrawing()
	}
	
	rl.CloseWindow()
	fmt.Println("Программа завершена.")
}