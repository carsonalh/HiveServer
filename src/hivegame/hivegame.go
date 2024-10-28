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
	var fromTile *HiveTile = nil

	for i, tile := range game.Tiles {
		if tile.Position == from {
			fromTile = &game.Tiles[i]
			break
		}
	}

	if fromTile == nil {
		// cannot move a tile which is not in play
		return false
	}

	if game.isPositionPinned(from) {
		// cannot move a tile if doing such would create multiple hives
		return false
	}

	var moves map[HexVectorInt]bool

	switch fromTile.PieceType {
	case PieceTypeQueenBee:
		moves = game.queenMoves(fromTile.Position)
	case PieceTypeSoldierAnt:
		moves = game.antMoves(fromTile.Position)
	case PieceTypeSpider:
		moves = game.spiderMoves(fromTile.Position)
	default:
		panic("unhandled case")
	}

	if _, ok := moves[to]; !ok {
		// not reachable via the rules
		return false
	}

	fromTile.Position = to
	return true
}

func (game *HiveGame) adjacentMoves(position, ignore HexVectorInt) []HexVectorInt {
	neighbours := position.AdjacentVectors()
	filledNeighbours := make(map[HexVectorInt]bool)

	withSharedNeighbours := make([]HexVectorInt, 0, 6)

	for _, neighbour := range neighbours {
		for _, tile := range game.Tiles {
			if tile.Position != ignore && tile.Position == neighbour {
				filledNeighbours[neighbour] = true
				break
			}
		}
	}

	for _, neighbour := range neighbours {
		if _, ok := filledNeighbours[neighbour]; ok {
			continue
		}

		for _, neighboursNeighbour := range neighbour.AdjacentVectors() {
			if _, ok := filledNeighbours[neighboursNeighbour]; ok {
				withSharedNeighbours = append(withSharedNeighbours, neighbour)
				break
			}
		}
	}

	// freedom to move
	withFreedomToMove := make([]HexVectorInt, 0, 6)

	for _, withSharedNeighbour := range withSharedNeighbours {
		direction := withSharedNeighbour.Subtract(position)
		clockwise := Rotate60().Transform(direction).Add(position)
		antiClockwise := Rotate300().Transform(direction).Add(position)

		_, clockwiseFilled := filledNeighbours[clockwise]
		_, antiClockwiseFilled := filledNeighbours[antiClockwise]

		if !clockwiseFilled || !antiClockwiseFilled {
			withFreedomToMove = append(withFreedomToMove, withSharedNeighbour)
		}
	}

	return withFreedomToMove
}

func (game *HiveGame) isPositionPinned(position HexVectorInt) bool {
	foundTile := false
	for _, tile := range game.Tiles {
		if tile.Position == position {
			foundTile = true
			break
		}
	}

	if !foundTile {
		panic("must only check for pins on tiles that are in the hive")
	}

	var neighbour *HexVectorInt = nil
	for _, tile := range game.Tiles {
		for _, adj := range position.AdjacentVectors() {
			if tile.Position == adj {
				neighbour = &adj
				break
			}
		}
	}

	if neighbour == nil {
		panic("every piece should have at least one neighbour by the time this function is called")
	}

	seen := make(map[HexVectorInt]bool)
	toExplore := make([]HexVectorInt, 0)

	toExplore = append(toExplore, *neighbour)

	for len(toExplore) > 0 {
		node := toExplore[0]
		toExplore = toExplore[1:]

		if _, ok := seen[node]; ok {
			continue
		}

		neighbours := make([]HexVectorInt, 0, 6)
		for _, adj := range node.AdjacentVectors() {
			if adj == position {
				continue
			}

			isOccupied := false
			for _, tile := range game.Tiles {
				if tile.Position == adj {
					isOccupied = true
					break
				}
			}

			if isOccupied {
				neighbours = append(neighbours, adj)
			}
		}

		toExplore = append(toExplore, neighbours...)

		seen[node] = true
	}

	return len(seen) != len(game.Tiles)-1
}

func (game *HiveGame) antMoves(from HexVectorInt) map[HexVectorInt]bool {
	seen := make(map[HexVectorInt]bool)
	toExplore := make([]HexVectorInt, 0)

	toExplore = append(toExplore, from)

	for len(toExplore) > 0 {
		node := toExplore[0]
		toExplore = toExplore[1:]

		if _, ok := seen[node]; ok {
			continue
		}

		next := game.adjacentMoves(node, from)

		toExplore = append(toExplore, next...)

		seen[node] = true
	}

	if _, ok := seen[from]; ok {
		// cannot move a tile to where it started
		delete(seen, from)
	}

	return seen
}

func (game *HiveGame) queenMoves(from HexVectorInt) map[HexVectorInt]bool {
	moves := game.adjacentMoves(from, from)
	asMap := make(map[HexVectorInt]bool)

	for _, move := range moves {
		asMap[move] = true
	}

	return asMap
}

func (game *HiveGame) spiderMoves(from HexVectorInt) map[HexVectorInt]bool {
	const SpiderMoveDistance = 3
	type searchNode struct {
		previous []HexVectorInt
		position HexVectorInt
	}

	search := make([]searchNode, 0)
	search = append(search, searchNode{previous: make([]HexVectorInt, 0), position: from})

	for range SpiderMoveDistance {
		for nodeIndex, node := range search {
			moves := game.adjacentMoves(node.position, from)
			newMoves := make([]HexVectorInt, 0, 6)
			for _, move := range moves {
				seenMove := false
				for _, previousMove := range node.previous {
					if move == previousMove {
						seenMove = true
						break
					}
				}

				if seenMove {
					continue
				}

				newMoves = append(newMoves, move)
			}

			for _, move := range newMoves {
				newPrevious := make([]HexVectorInt, 0, SpiderMoveDistance)

				for _, p := range node.previous {
					newPrevious = append(newPrevious, p)
				}

				newPrevious = append(newPrevious, node.position)

				newNode := searchNode{previous: newPrevious, position: move}

				search = append(search, newNode)
			}

			search = append(search[:nodeIndex], search[nodeIndex+1:]...)
		}
	}

	validMoves := make(map[HexVectorInt]bool)
	for _, move := range search {
		if move.position != from {
			validMoves[move.position] = true
		}
	}

	return validMoves
}
