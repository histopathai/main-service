package vobj

type Point struct {
	X float64
	Y float64
}

func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

func (p Point) Equals(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

func (p Point) String() string {
	return "(" + string(rune(p.X)) + ", " + string(rune(p.Y)) + ")"
}
