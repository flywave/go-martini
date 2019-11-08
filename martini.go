// Copyright (c) 2017-present FlyWave, Inc. All Rights Reserved.
// See License.txt for license information.

package martini

import (
	"errors"
	"math"
)

type Martini struct {
	GridSize           int
	NumTriangles       int
	NumParentTriangles int
	Indices            []uint16
	Coords             []uint16
}

func NewMartini(gridSize int) (*Martini, error) {
	mt := Martini{}
	mt.GridSize = gridSize
	tileSize := gridSize - 1
	if (tileSize & (tileSize - 1)) > 0 {
		return nil, errors.New("Expected grid size to be 2^n+1")
	}
	mt.NumTriangles = tileSize*tileSize*2 - 2
	mt.NumParentTriangles = mt.NumTriangles - tileSize*tileSize
	mt.Indices = make([]uint16, gridSize*gridSize)
	mt.Coords = make([]uint16, mt.NumTriangles*4)
	for i := 0; i < mt.NumTriangles; i++ {
		id := i + 2
		ax := 0
		ay := 0
		bx := 0
		by := 0
		cx := 0
		cy := 0
		if (id & 1) > 0 {
			bx = tileSize
			by = tileSize
			cx = tileSize
		} else {
			ax = tileSize
			ay = tileSize
			cy = tileSize
		}
		id >>= 1
		if id > 1 {
			for {
				mx := (ax + bx) >> 1
				my := (ay + by) >> 1

				if (id & 1) > 0 { // left half
					bx = ax
					by = ay
					ax = cx
					ay = cy
				} else { // right half
					ax = bx
					ay = by
					bx = cx
					by = cy
				}
				cx = mx
				cy = my
				id >>= 1
				if id <= 1 {
					break
				}
			}
		}

		k := i * 4
		mt.Coords[k+0] = uint16(ax)
		mt.Coords[k+1] = uint16(ay)
		mt.Coords[k+2] = uint16(bx)
		mt.Coords[k+3] = uint16(by)
	}
	return &mt, nil
}

func (m *Martini) CreateTile(terrain []float64) (*Tile, error) {
	return NewTile(terrain, m)
}

type Tile struct {
	Terrain []float64
	Martini *Martini
	Errors  []float64
}

func NewTile(terrain []float64, martini *Martini) (*Tile, error) {
	size := martini.GridSize
	if len(terrain) != size*size {
		return nil, errors.New("Expected terrain data of length ")
	}
	errors := make([]float64, len(terrain))
	t := Tile{Terrain: terrain, Martini: martini, Errors: errors}
	t.Update()
	return &t, nil
}

func (t *Tile) Update() {
	m := t.Martini
	size := m.GridSize

	for i := m.NumTriangles - 1; i >= 0; i-- {
		k := i * 4
		ax := m.Coords[k+0]
		ay := m.Coords[k+1]
		bx := m.Coords[k+2]
		by := m.Coords[k+3]
		mx := (ax + bx) >> 1
		my := (ay + by) >> 1
		cx := mx + my - ay
		cy := my + ax - mx

		interpolatedHeight := (t.Terrain[int(ay)*size+int(ax)] + t.Terrain[int(by)*size+int(bx)]) / 2
		middleIndex := int(my)*size + int(mx)
		middleError := math.Abs(interpolatedHeight - t.Terrain[middleIndex])

		t.Errors[middleIndex] = math.Max(t.Errors[middleIndex], middleError)

		if i < m.NumParentTriangles {
			leftChildIndex := (int(ay+cy)>>1)*size + (int(ax+cx) >> 1)
			rightChildIndex := (int(by+cy)>>1)*size + (int(bx+cx) >> 1)
			t.Errors[middleIndex] = math.Max(math.Max(t.Errors[middleIndex], t.Errors[leftChildIndex]), t.Errors[rightChildIndex])
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (t *Tile) countElements(ax, ay, bx, by, cx, cy int, maxError float64, numTriangles *int, numVertices *int) {
	m := t.Martini
	size := m.GridSize

	mx := (ax + bx) >> 1
	my := (ay + by) >> 1

	if abs(ax-cx)+abs(ay-cy) > 1 && t.Errors[my*size+mx] > maxError {
		t.countElements(cx, cy, ax, ay, mx, my, maxError, numTriangles, numVertices)
		t.countElements(bx, by, cx, cy, mx, my, maxError, numTriangles, numVertices)
	} else {
		if m.Indices[ay*size+ax] == 0 {
			(*numVertices)++
			m.Indices[ay*size+ax] = uint16(*numVertices)
		}
		if m.Indices[by*size+bx] == 0 {
			(*numVertices)++
			m.Indices[by*size+bx] = uint16(*numVertices)
		}
		if m.Indices[cy*size+cx] == 0 {
			(*numVertices)++
			m.Indices[cy*size+cx] = uint16(*numVertices)
		}
		(*numTriangles)++
	}
}

func (t *Tile) processTriangle(ax, ay, bx, by, cx, cy int, maxError float64, triIndex *int, vertices, triangles []uint16) {
	m := t.Martini
	size := m.GridSize

	mx := (ax + bx) >> 1
	my := (ay + by) >> 1

	if abs(ax-cx)+abs(ay-cy) > 1 && t.Errors[my*size+mx] > maxError {
		t.processTriangle(cx, cy, ax, ay, mx, my, maxError, triIndex, vertices, triangles)
		t.processTriangle(bx, by, cx, cy, mx, my, maxError, triIndex, vertices, triangles)

	} else {
		a := t.Martini.Indices[ay*size+ax] - 1
		b := t.Martini.Indices[by*size+bx] - 1
		c := t.Martini.Indices[cy*size+cx] - 1

		vertices[2*a] = uint16(ax)
		vertices[2*a+1] = uint16(ay)

		vertices[2*b] = uint16(bx)
		vertices[2*b+1] = uint16(by)

		vertices[2*c] = uint16(cx)
		vertices[2*c+1] = uint16(cy)

		triangles[*triIndex] = a
		(*triIndex)++
		triangles[*triIndex] = b
		(*triIndex)++
		triangles[*triIndex] = c
		(*triIndex)++
	}
}

func (t *Tile) GetMesh(maxError float64) ([]uint16, []uint16) {
	m := t.Martini
	size := m.GridSize

	numVertices := 0
	numTriangles := 0
	max := size - 1

	for i := range m.Indices {
		m.Indices[i] = 0
	}

	t.countElements(0, 0, max, max, max, 0, maxError, &numTriangles, &numVertices)
	t.countElements(max, max, 0, 0, 0, max, maxError, &numTriangles, &numVertices)

	vertices := make([]uint16, numVertices*2)
	triangles := make([]uint16, numTriangles*3)
	triIndex := 0

	t.processTriangle(0, 0, max, max, max, 0, maxError, &triIndex, vertices, triangles)
	t.processTriangle(max, max, 0, 0, 0, max, maxError, &triIndex, vertices, triangles)

	return vertices, triangles
}
