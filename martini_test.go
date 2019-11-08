// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package martini

import (
	"image/color"
	"image/png"
	"os"
	"testing"
)

var (
	Vertices = []uint16{
		320, 64, 256, 128, 320, 128, 384, 128, 256, 0, 288, 160, 256, 192, 288, 192, 320, 192, 304, 176, 256, 256, 288,
		224, 352, 160, 320, 160, 512, 0, 384, 0, 128, 128, 128, 0, 64, 64, 64, 0, 0, 0, 32, 32, 192, 192, 384, 384, 512,
		256, 384, 256, 320, 320, 320, 256, 512, 512, 512, 128, 448, 192, 384, 192, 128, 384, 256, 512, 256, 384, 0,
		512, 128, 256, 64, 192, 0, 256, 64, 128, 32, 96, 0, 128, 32, 64, 16, 48, 0, 64, 0, 32,
	}
	Triangles = []uint16{
		0, 1, 2, 3, 0, 2, 4, 1, 0, 5, 6, 7, 7, 8, 9, 5, 7, 9, 1, 6, 5, 6, 10, 11, 11, 8, 7, 6, 11, 7, 12, 2, 13, 8, 12,
		13, 3, 2, 12, 2, 1, 5, 13, 5, 9, 8, 13, 9, 2, 5, 13, 3, 14, 15, 15, 4, 0, 3, 15, 0, 16, 4, 17, 18, 17, 19, 19,
		20, 21, 18, 19, 21, 16, 17, 18, 1, 16, 22, 22, 10, 6, 1, 22, 6, 4, 16, 1, 23, 24, 25, 26, 25, 27, 10, 26, 27,
		23, 25, 26, 28, 24, 23, 29, 3, 30, 24, 29, 30, 14, 3, 29, 8, 25, 31, 31, 3, 12, 8, 31, 12, 27, 8, 11, 10, 27,
		11, 25, 8, 27, 25, 24, 30, 30, 3, 31, 25, 30, 31, 32, 33, 34, 10, 32, 34, 35, 33, 32, 33, 28, 23, 34, 23, 26,
		10, 34, 26, 33, 23, 34, 36, 16, 37, 38, 36, 37, 36, 10, 22, 16, 36, 22, 39, 18, 40, 41, 39, 40, 16, 18, 39, 42,
		21, 43, 44, 42, 43, 18, 21, 42, 21, 20, 45, 45, 44, 43, 21, 45, 43, 44, 41, 40, 40, 18, 42, 44, 40, 42, 41, 38,
		37, 37, 16, 39, 41, 37, 39, 38, 35, 32, 32, 10, 36, 38, 32, 36,
	}
)

func LoadPngData(path string) ([]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	m, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	rect := m.Bounds()

	gridSize := rect.Dx() + 1

	terrain := make([]float64, gridSize*gridSize)

	for x := 0; x < rect.Dx(); x++ {
		for y := 0; y < rect.Dy(); y++ {
			rgba := m.At(x, y).(color.NRGBA)
			terrain[y*gridSize+x] = float64(int(rgba.R)*256*256+int(rgba.G)*256.0+int(rgba.B))/10.0 - 10000.0
		}
	}
	for x := 0; x < gridSize-1; x++ {
		terrain[gridSize*(gridSize-1)+x] = terrain[gridSize*(gridSize-2)+x]
	}
	for y := 0; y < gridSize; y++ {
		terrain[gridSize*y+gridSize-1] = terrain[gridSize*y+gridSize-2]
	}
	return terrain, nil
}

func TestGenerates(t *testing.T) {
	var terrain []float64
	if fuji, err := LoadPngData("./tests/fuji.png"); err != nil {
		t.Error(err)
	} else {
		terrain = fuji
	}

	martini, _ := NewMartini(513)
	tile, _ := martini.CreateTile(terrain)
	vertices, triangles := tile.GetMesh(500)

	if len(terrain) == 0 || len(vertices) == 0 || len(triangles) == 0 {
		t.Error("ss")
	}
}
