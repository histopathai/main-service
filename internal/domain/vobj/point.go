package vobj

type Point struct {
	X float64
	Y float64
}

func (p Point) Equals(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

func (p Point) String() string {
	return "(" + string(rune(p.X)) + ", " + string(rune(p.Y)) + ")"
}

func (p Point) GetMap() map[string]float64 {
	return map[string]float64{
		"X": p.X,
		"Y": p.Y,
	}
}
