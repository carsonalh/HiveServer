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
	color     HiveColor
	position  HexVectorInt
	pieceType HivePieceType
}

type HiveGame struct {
	colorToMove HiveColor
	// what move of the game we are on; starts at 1
	move  int
	tiles []HiveTile
}

func CreateHiveGame() HiveGame {
	return HiveGame{
		colorToMove: ColorBlack,
		move:        1,
		tiles:       make([]HiveTile, 0),
	}
}

func (game *HiveGame) PlaceTile(position HexVectorInt, pieceType HivePieceType) bool {
	// TODO check this is a valid placement

	game.tiles = append(game.tiles, HiveTile{
		color:     game.colorToMove,
		position:  position,
		pieceType: pieceType,
	})

	if game.colorToMove == ColorWhite {
		game.move++
	}

	if game.colorToMove == ColorBlack {
		game.colorToMove = ColorWhite
	} else {
		game.colorToMove = ColorBlack
	}

	return true
}

func (game *HiveGame) MoveTile(from, to HexVectorInt) bool {
	return false
}
