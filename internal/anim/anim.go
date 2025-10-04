package anim

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    rl "github.com/gen2brain/raylib-go/raylib"
)

type Def struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Image       string   `json:"image"`
    FrameWidth  int32    `json:"frameWidth"`
    FrameHeight int32    `json:"frameHeight"`
    Rows        int      `json:"rows"`
    Cols        int      `json:"cols"`
    FPS         float32  `json:"fps"`
    Loop        bool     `json:"loop"`
    Origin      [2]int32 `json:"origin"`
}

type Frame struct {
    Src   rl.Rectangle
    OrigX int32
    OrigY int32
}

type Clip struct {
    Name   string
    FPS    float32
    Loop   bool
    Tex    rl.Texture2D
    Frames []Frame
}

type Animator struct {
    Current    *Clip
    Elapsed    float32
    FrameIndex int
    FlipX      bool
}

func (a *Animator) Play(c *Clip, reset bool) {
    if a.Current == c && !reset { return }
    a.Current = c
    a.Elapsed = 0
    a.FrameIndex = 0
}

func (a *Animator) Update(dt float32) {
    if a.Current == nil || len(a.Current.Frames) == 0 { return }
    a.Elapsed += dt
    dur := 1.0 / a.Current.FPS
    for a.Elapsed >= dur {
        a.Elapsed -= dur
        a.FrameIndex++
        if a.FrameIndex >= len(a.Current.Frames) {
            if a.Current.Loop { a.FrameIndex = 0 } else { a.FrameIndex = len(a.Current.Frames)-1 }
        }
    }
}

func (a *Animator) Draw(x, y, scale float32, tint rl.Color) {
    if a.Current == nil || a.Current.Tex.ID == 0 { return }
    f := a.Current.Frames[a.FrameIndex]
    src := f.Src
    if a.FlipX { src.Width = -src.Width }
    dst := rl.NewRectangle(x-float32(f.OrigX)*scale, y-float32(f.OrigY)*scale, f.Src.Width*scale, f.Src.Height*scale)
    rl.DrawTexturePro(a.Current.Tex, src, dst, rl.NewVector2(0,0), 0, tint)
}

// ---------- загрузка ----------
func LoadFromJSON(jsonPath string) (*Clip, error) {
    data, err := os.ReadFile(jsonPath)
    if err != nil { return nil, err }
    var d Def
    if err := json.Unmarshal(data, &d); err != nil { return nil, err }
    if d.Type != "sheet" && d.Type != "" { return nil, fmt.Errorf("only 'sheet' supported in this minimal loader") }

    imgPath := filepath.Join(filepath.Dir(jsonPath), d.Image)
    img := rl.LoadImage(imgPath)
    if img.Data == nil { return nil, fmt.Errorf("open image: %s", imgPath) }
    defer rl.UnloadImage(img)

    tex := rl.LoadTextureFromImage(img)
    if tex.ID == 0 { return nil, fmt.Errorf("texture from: %s", imgPath) }
    rl.SetTextureFilter(tex, rl.FilterPoint)

    fw, fh := d.FrameWidth, d.FrameHeight
    if fw == 0 || fh == 0 {
        fw = img.Width / int32(d.Cols)
        fh = img.Height / int32(d.Rows)
    }

    frames := make([]Frame, 0, d.Rows*d.Cols)
    x, y := int32(0), int32(0)
    for r := 0; r < d.Rows; r++ {
        x = 0
        for c := 0; c < d.Cols; c++ {
            frames = append(frames, Frame{Src: rl.NewRectangle(float32(x), float32(y), float32(fw), float32(fh)), OrigX: d.Origin[0], OrigY: d.Origin[1]})
            x += fw
        }
        y += fh
    }

    return &Clip{ Name: d.Name, FPS: ifnz(d.FPS, 10), Loop: d.Loop, Tex: tex, Frames: frames }, nil
}

func ifnz(v, def float32) float32 { if v <= 0 { return def }; return v }
