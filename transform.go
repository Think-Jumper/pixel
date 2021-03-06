package pixel

import "github.com/go-gl/mathgl/mgl32"

// Transform holds space transformation information. Concretely, a transformation is specified
// by position, anchor, scale and rotation.
//
// All points are first rotated around the anchor. Then they are multiplied by the scale. If
// the scale factor is 2, the object becomes 2x bigger. Finally, all points are moved, so that
// the original anchor is located precisely at the position.
//
// Create a Transform object with Position/Anchor/Rotation/... function. This sets the position
// one of it's properties. Then use methods, like Scale and Rotate to change scale, rotation and
// achor. The order in which you apply these methods is irrelevant.
//
//   pixel.Position(pixel.V(100, 100)).Rotate(math.Pi / 3).Scale(1.5)
//
// Also note, that no method changes the Transform. All simply return a new, changed Transform.
type Transform struct {
	pos, anc, sca Vec
	rot           float64
}

// ZT stands for Zero-Transform. This Transform is a neutral Transform, does not change anything.
var ZT = Transform{}.Scale(1)

// Position returns a Zero-Transform with Position set to pos.
func Position(pos Vec) Transform {
	return ZT.Position(pos)
}

// Anchor returns a Zero-Transform with Anchor set to anchor.
func Anchor(anchor Vec) Transform {
	return ZT.Anchor(anchor)
}

// Scale returns a Zero-Transform with Scale set to scale.
func Scale(scale float64) Transform {
	return ZT.Scale(scale)
}

// ScaleXY returns a Zero-Transform with ScaleXY set to scale.
func ScaleXY(scale Vec) Transform {
	return ZT.ScaleXY(scale)
}

// Rotation returns a Zero-Transform with Rotation set to angle (in radians).
func Rotation(angle float64) Transform {
	return ZT.Rotation(angle)
}

// Position moves an object by the specified vector. A zero vector will end up precisely at pos.
func (t Transform) Position(pos Vec) Transform {
	t.pos = pos
	return t
}

// AddPosition adds delta to the existing Position of this Transform.
func (t Transform) AddPosition(delta Vec) Transform {
	t.pos += delta
	return t
}

// Anchor specifies the zero vector, point originally located at anchor will be treated as zero.
// This affects Rotation and Position.
func (t Transform) Anchor(anchor Vec) Transform {
	t.anc = anchor
	return t
}

// AddAnchor adds delta to the existing Anchor of this Transform.
func (t Transform) AddAnchor(delta Vec) Transform {
	t.anc += delta
	return t
}

// Scale specifies a factor by which an object will be scaled around it's Anchor.
//
// Same as:
//   t.ScaleXY(pixel.V(scale, scale)).
func (t Transform) Scale(scale float64) Transform {
	t.sca = V(scale, scale)
	return t
}

// MulScale multiplies the existing Scale of this Transform by factor.
//
// Same as:
//   t.MulScaleXY(pixel.V(factor, factor)).
func (t Transform) MulScale(factor float64) Transform {
	t.sca = t.sca.Scaled(factor)
	return t
}

// ScaleXY specifies a factor in each dimension, by which an object will be scaled around it's
// Anchor.
func (t Transform) ScaleXY(scale Vec) Transform {
	t.sca = scale
	return t
}

// MulScaleXY multiplies the existing ScaleXY of this Transform by factor, component-wise.
func (t Transform) MulScaleXY(factor Vec) Transform {
	t.sca = V(
		t.sca.X()*factor.X(),
		t.sca.Y()*factor.Y(),
	)
	return t
}

// Rotation specifies an angle by which an object will be rotated around it's Anchor.
//
// The angle is in radians.
func (t Transform) Rotation(angle float64) Transform {
	t.rot = angle
	return t
}

// AddRotation adds delta to the existing Angle of this Transform.
//
// The delta is in radians.
func (t Transform) AddRotation(delta float64) Transform {
	t.rot += delta
	return t
}

// GetPosition returns the Position of the Transform.
func (t Transform) GetPosition() Vec {
	return t.pos
}

// GetAnchor returns the Anchor of the Transform.
func (t Transform) GetAnchor() Vec {
	return t.anc
}

// GetScaleXY returns the ScaleXY of the Transform.
func (t Transform) GetScaleXY() Vec {
	return t.sca
}

// GetRotation returns the Rotation of the Transform.
func (t Transform) GetRotation() float64 {
	return t.rot
}

// Project transforms a vector by a transform.
func (t Transform) Project(v Vec) Vec {
	mat := t.Mat()
	vec := mgl32.Vec3{float32(v.X()), float32(v.Y()), 1}
	pro := mat.Mul3x1(vec)
	return V(float64(pro.X()), float64(pro.Y()))
}

// Unproject does the inverse operation to Project.
func (t Transform) Unproject(v Vec) Vec {
	mat := t.InvMat()
	vec := mgl32.Vec3{float32(v.X()), float32(v.Y()), 1}
	unp := mat.Mul3x1(vec)
	return V(float64(unp.X()), float64(unp.Y()))
}

// Mat returns a transformation matrix that satisfies previously set transform properties.
func (t Transform) Mat() mgl32.Mat3 {
	mat := mgl32.Ident3()
	mat = mat.Mul3(mgl32.Translate2D(float32(t.pos.X()), float32(t.pos.Y())))
	mat = mat.Mul3(mgl32.Rotate3DZ(float32(t.rot)))
	mat = mat.Mul3(mgl32.Scale2D(float32(t.sca.X()), float32(t.sca.Y())))
	mat = mat.Mul3(mgl32.Translate2D(float32(-t.anc.X()), float32(-t.anc.Y())))
	return mat
}

// InvMat returns an inverse transformation matrix to the matrix returned by Mat3 method.
func (t Transform) InvMat() mgl32.Mat3 {
	mat := mgl32.Ident3()
	mat = mat.Mul3(mgl32.Translate2D(float32(t.anc.X()), float32(t.anc.Y())))
	mat = mat.Mul3(mgl32.Scale2D(float32(1/t.sca.X()), float32(1/t.sca.Y())))
	mat = mat.Mul3(mgl32.Rotate3DZ(float32(-t.rot)))
	mat = mat.Mul3(mgl32.Translate2D(float32(-t.pos.X()), float32(-t.pos.Y())))
	return mat
}

// Camera is a convenience function, that returns a Transform that acts like a camera.	Center is
// the position in the world coordinates, that will be projected onto the center of the screen.
// One unit in world coordinates will be projected onto zoom pixels.
//
// It is possible to apply additional rotations, scales and moves to the returned transform.
func Camera(center, zoom, screenSize Vec) Transform {
	return Anchor(center).ScaleXY(2 * zoom).MulScaleXY(V(1/screenSize.X(), 1/screenSize.Y()))
}
