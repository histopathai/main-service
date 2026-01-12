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

func ToJSONPoints(points []Point) []map[string]float64 {
	jsonPoints := make([]map[string]float64, len(points))
	for i, point := range points {
		jsonPoints[i] = map[string]float64{
			"X": point.X,
			"Y": point.Y,
		}
	}
	return jsonPoints
}

func FromJSONPoints(jsonPoints []map[string]float64) []Point {
	points := make([]Point, len(jsonPoints))
	for i, jp := range jsonPoints {
		points[i] = Point{
			X: jp["X"],
			Y: jp["Y"],
		}
	}
	return points
}
