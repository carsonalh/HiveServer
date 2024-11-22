package hivegame

type HexMatrixInt struct {
	A00, A01, A10, A11 int
}

type HexVectorInt struct {
	Q int `json:"q"`
	R int `json:"r"`
}

func (v HexVectorInt) Add(u HexVectorInt) HexVectorInt {
	return HexVectorInt{Q: v.Q + u.Q, R: v.R + u.R}
}

func (v HexVectorInt) Subtract(u HexVectorInt) HexVectorInt {
	return HexVectorInt{Q: v.Q - u.Q, R: v.R - u.R}
}

func (v HexVectorInt) MultiplyScalar(s int) HexVectorInt {
	return HexVectorInt{Q: v.Q * s, R: v.R * s}
}

func (v HexVectorInt) AdjacentVectors() [6]HexVectorInt {
	adjacentUnits := [6]HexVectorInt{
		{1, 0},
		{1, -1},
		{0, -1},
		{-1, 0},
		{-1, 1},
		{0, 1},
	}

	for i, u := range adjacentUnits {
		adjacentUnits[i] = v.Add(u)
	}

	return adjacentUnits
}

func (m HexMatrixInt) Transform(v HexVectorInt) HexVectorInt {
	return HexVectorInt{
		Q: m.A00*v.Q + m.A01*v.R,
		R: m.A10*v.Q + m.A11*v.R,
	}
}

func Rotate60() HexMatrixInt {
	return HexMatrixInt{
		1, 1,
		-1, 0,
	}
}

func Rotate300() HexMatrixInt {
	return HexMatrixInt{
		0, -1,
		1, 1,
	}
}
