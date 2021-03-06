package pixel

import (
	"fmt"
	"image/color"
)

// TrianglesData specifies a list of Triangles vertices with three common properties: Position,
// Color and Texture.
type TrianglesData []struct {
	Position Vec
	Color    NRGBA
	Texture  Vec
}

// MakeTrianglesData creates TrianglesData of length len initialized with default property values.
//
// Prefer this function to make(TrianglesData, len), because make zeros them, while this function
// does a correct intialization.
func MakeTrianglesData(len int) TrianglesData {
	td := TrianglesData{}
	td.SetLen(len)
	return td
}

// Len returns the number of vertices in TrianglesData.
func (td *TrianglesData) Len() int {
	return len(*td)
}

// SetLen resizes TrianglesData to len, while keeping the original content.
//
// If len is greater than TrianglesData's current length, the new data is filled with default
// values ((0, 0), white, (-1, -1)).
func (td *TrianglesData) SetLen(len int) {
	if len > td.Len() {
		needAppend := len - td.Len()
		for i := 0; i < needAppend; i++ {
			*td = append(*td, struct {
				Position Vec
				Color    NRGBA
				Texture  Vec
			}{V(0, 0), NRGBA{1, 1, 1, 1}, V(-1, -1)})
		}
	}
	if len < td.Len() {
		*td = (*td)[:len]
	}
}

// Slice returns a sub-Triangles of this TrianglesData.
func (td *TrianglesData) Slice(i, j int) Triangles {
	s := TrianglesData((*td)[i:j])
	return &s
}

func (td *TrianglesData) updateData(t Triangles) {
	// fast path optimization
	if t, ok := t.(*TrianglesData); ok {
		copy(*td, *t)
		return
	}

	// slow path manual copy
	if t, ok := t.(TrianglesPosition); ok {
		for i := range *td {
			(*td)[i].Position = t.Position(i)
		}
	}
	if t, ok := t.(TrianglesColor); ok {
		for i := range *td {
			(*td)[i].Color = t.Color(i)
		}
	}
	if t, ok := t.(TrianglesTexture); ok {
		for i := range *td {
			(*td)[i].Texture = t.Texture(i)
		}
	}
}

// Update copies vertex properties from the supplied Triangles into this TrianglesData.
//
// TrianglesPosition, TrianglesColor and TrianglesTexture are supported.
func (td *TrianglesData) Update(t Triangles) {
	if td.Len() != t.Len() {
		panic(fmt.Errorf("%T.Update: invalid triangles length", td))
	}
	td.updateData(t)
}

// Copy returns an exact independent copy of this TrianglesData.
func (td *TrianglesData) Copy() Triangles {
	copyTd := TrianglesData{}
	copyTd.SetLen(td.Len())
	copyTd.Update(td)
	return &copyTd
}

// Position returns the position property of i-th vertex.
func (td *TrianglesData) Position(i int) Vec {
	return (*td)[i].Position
}

// Color returns the color property of i-th vertex.
func (td *TrianglesData) Color(i int) NRGBA {
	return (*td)[i].Color
}

// Texture returns the texture property of i-th vertex.
func (td *TrianglesData) Texture(i int) Vec {
	return (*td)[i].Texture
}

// TrianglesDrawer is a helper type that wraps Triangles and turns them into a Drawer.
//
// It does so by creating a separate Triangles instance for each Target. The instances are
// correctly updated alongside the wrapped Triangles.
type TrianglesDrawer struct {
	Triangles

	tris  map[Target]TargetTriangles
	dirty bool
}

func (td *TrianglesDrawer) flush() {
	if !td.dirty {
		return
	}
	td.dirty = false

	for _, t := range td.tris {
		t.SetLen(td.Len())
		t.Update(td.Triangles)
	}
}

// Draw draws the wrapped Triangles onto the provided Target.
func (td *TrianglesDrawer) Draw(target Target) {
	if td.tris == nil {
		td.tris = make(map[Target]TargetTriangles)
	}

	td.flush()

	tri := td.tris[target]
	if tri == nil {
		tri = target.MakeTriangles(td.Triangles)
		td.tris[target] = tri
	}
	tri.Draw()
}

// Dirty marks the underlying container as changed (dirty). Always call this when you change the
// underlying Triangles (by Update, SetLen, etc.).
func (td *TrianglesDrawer) Dirty() {
	td.dirty = true
}

// Sprite is a picture that can be drawn onto a Target. To change the position/rotation/scale of
// the Sprite, use Target's SetTransform method.
type Sprite struct {
	data TrianglesData
	td   TrianglesDrawer
	pic  *Picture
}

// NewSprite creates a Sprite with the supplied Picture. The dimensions of the returned Sprite match
// the dimensions of the Picture.
func NewSprite(pic *Picture) *Sprite {
	s := &Sprite{
		data: TrianglesData{
			{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(0, 0)},
			{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(1, 0)},
			{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(1, 1)},
			{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(0, 0)},
			{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(1, 1)},
			{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(0, 1)},
		},
	}
	s.td = TrianglesDrawer{Triangles: &s.data}
	s.SetPicture(pic)
	return s
}

// SetPicture changes the Picture of the Sprite and resizes it accordingly.
func (s *Sprite) SetPicture(pic *Picture) {
	oldPic := s.pic
	s.pic = pic
	if oldPic != nil && oldPic.Bounds().Size == pic.Bounds().Size {
		return
	}
	w, h := pic.Bounds().Size.XY()
	s.data[0].Position = V(0, 0)
	s.data[1].Position = V(w, 0)
	s.data[2].Position = V(w, h)
	s.data[3].Position = V(0, 0)
	s.data[4].Position = V(w, h)
	s.data[5].Position = V(0, h)
	s.td.Dirty()
}

// Picture returns the current Picture of the Sprite.
func (s *Sprite) Picture() *Picture {
	return s.pic
}

// Draw draws the Sprite onto the provided Target.
func (s *Sprite) Draw(t Target) {
	t.SetPicture(s.pic)
	s.td.Draw(t)
}

// Polygon is a convex polygon shape filled with a single color.
type Polygon struct {
	data TrianglesData
	td   TrianglesDrawer
	col  NRGBA
}

// NewPolygon creates a Polygon with specified color and points. Points can be in clock-wise or
// counter-clock-wise order, it doesn't matter. They should however form a convex polygon.
func NewPolygon(c color.Color, points ...Vec) *Polygon {
	p := &Polygon{
		data: TrianglesData{},
	}
	p.td = TrianglesDrawer{Triangles: &p.data}
	p.SetColor(c)
	p.SetPoints(points...)
	return p
}

// SetColor changes the color of the Polygon.
//
// If the Polygon is very large, this method might end up being too expensive. Consider using
// a color mask on a Target, in such a case.
func (p *Polygon) SetColor(c color.Color) {
	p.col = NRGBAModel.Convert(c).(NRGBA)
	for i := range p.data {
		p.data[i].Color = p.col
	}
	p.td.Dirty()
}

// Color returns the current color of the Polygon.
func (p *Polygon) Color() NRGBA {
	return p.col
}

// SetPoints sets the points of the Polygon. The number of points might differ from the original
// count.
//
// This method is more effective, than creating a new Polygon with the given points.
//
// However, it is less expensive than using a transform on a Target.
func (p *Polygon) SetPoints(points ...Vec) {
	p.data.SetLen(3 * (len(points) - 2))
	for i := 2; i < len(points); i++ {
		p.data[(i-2)*3+0].Position = points[0]
		p.data[(i-2)*3+1].Position = points[i-1]
		p.data[(i-2)*3+2].Position = points[i]
	}
	for i := range p.data {
		p.data[i].Color = p.col
	}
	p.td.Dirty()
}

// Points returns a slice of points of the Polygon in the order they where supplied.
func (p *Polygon) Points() []Vec {
	points := make([]Vec, p.data.Len())
	for i := range p.data {
		points[i] = p.data[i].Position
	}
	return points
}

// Draw draws the Polygon onto the Target.
func (p *Polygon) Draw(t Target) {
	t.SetPicture(nil)
	p.td.Draw(t)
}
