package hivegame

type HiveColor = int

const (
	ColorBlack HiveColor = iota
	ColorWhite
)

type HivePieceType = int

const (
	PieceTypeQueenBee HivePieceType = iota
	PieceTypeSoldierAnt
	PieceTypeGrasshopper
	PieceTypeSpider
	PieceTypeBeetle
	PieceTypeLadybug
	PieceTypeMosquito
)

type HiveTile struct {
	Color     HiveColor
	Position  HexVectorInt
	PieceType HivePieceType
}

type HiveGame struct {
	ColorToMove HiveColor
	// what Move of the game we are on; starts at 1
	Move         int
	WhiteReserve map[HivePieceType]int
	BlackReserve map[HivePieceType]int
	Tiles        []HiveTile
}

func CreateHiveGame() HiveGame {
	return HiveGame{
		ColorToMove: ColorBlack,
		Move:        1,
		Tiles:       make([]HiveTile, 0),
		WhiteReserve: map[HivePieceType]int{
			PieceTypeQueenBee:    1,
			PieceTypeGrasshopper: 3,
			PieceTypeSpider:      2,
			PieceTypeSoldierAnt:  3,
			PieceTypeBeetle:      2,
			PieceTypeLadybug:     1,
			PieceTypeMosquito:    1,
		},
		BlackReserve: map[HivePieceType]int{
			PieceTypeQueenBee:    1,
			PieceTypeGrasshopper: 3,
			PieceTypeSpider:      2,
			PieceTypeSoldierAnt:  3,
			PieceTypeBeetle:      2,
			PieceTypeLadybug:     1,
			PieceTypeMosquito:    1,
		},
	}
}

func (game *HiveGame) PlaceTile(position HexVectorInt, pieceType HivePieceType) bool {
	for _, tile := range game.Tiles {
		if tile.Position == position {
			// cannot place a tile atop another
			return false
		}
	}

	queenPlaced := false
	for _, tile := range game.Tiles {
		if tile.Color == game.ColorToMove && tile.PieceType == PieceTypeQueenBee {
			queenPlaced = true
			break
		}
	}

	if !queenPlaced && game.Move == 4 && pieceType != PieceTypeQueenBee {
		return false
	}

	var reserve map[HivePieceType]int

	if game.ColorToMove == ColorBlack {
		reserve = game.BlackReserve
	} else {
		reserve = game.WhiteReserve
	}

	count, ok := reserve[pieceType]

	if !ok {
		panic("we have already preallocated every piece type, so this should _not_ be happening")
	}

	if count < 0 {
		panic("illegal value in map")
	}

	if count == 0 {
		// we cannot place the piece in question if it is not in our reserve
		return false
	}

	if game.Move == 1 && game.ColorToMove == ColorWhite {
		if len(game.Tiles) != 1 {
			panic("there should be a (black) tile on the board by white's first Move")
		}

		found := false
		for _, adj := range position.AdjacentVectors() {
			if adj == game.Tiles[0].Position {
				found = true
				break
			}
		}

		if !found {
			// must place adjacent to the black tile
			return false
		}
	} else if game.Move > 1 {
		touchesOwn, touchesOpposition := false, false

		for _, tile := range game.Tiles {
			for _, adj := range position.AdjacentVectors() {
				if tile.Position == adj {
					if game.ColorToMove == tile.Color {
						touchesOwn = true
					} else {
						touchesOpposition = true
					}
				}
			}
		}

		if !touchesOwn || touchesOpposition {
			// violation of regular placing rule
			return false
		}
	}

	reserve[pieceType] = count - 1

	game.Tiles = append(game.Tiles, HiveTile{
		Color:     game.ColorToMove,
		Position:  position,
		PieceType: pieceType,
	})

	if game.ColorToMove == ColorWhite {
		game.Move++
	}

	if game.ColorToMove == ColorBlack {
		game.ColorToMove = ColorWhite
	} else {
		game.ColorToMove = ColorBlack
	}

	return true
}

func (game *HiveGame) MoveTile(from, to HexVectorInt) bool {
	return false
}
