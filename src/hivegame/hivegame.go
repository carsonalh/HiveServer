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
	Color       HiveColor
	Position    HexVectorInt
	PieceType   HivePieceType
	StackHeight int
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

		for _, adj := range position.AdjacentVectors() {
			tileAtAdj := game.tileAt(adj)

			if tileAtAdj == nil {
				continue
			}

			if tileAtAdj.Color == game.ColorToMove {
				touchesOwn = true
			} else {
				touchesOpposition = true
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

	game.incrementMove()

	return true
}

func (game *HiveGame) MoveTile(from, to HexVectorInt) bool {
	fromTile := game.tileAt(from)

	if fromTile == nil {
		// cannot move a tile which is not in play
		return false
	}

	if fromTile.Color != game.ColorToMove {
		return false
	}

	if game.isTilePinned(fromTile) {
		// cannot move a tile if doing such would create multiple hives
		return false
	}

	var moves map[HexVectorInt]bool

	switch fromTile.PieceType {
	case PieceTypeQueenBee:
		moves = game.queenBeeMoves(fromTile.Position)
	case PieceTypeSoldierAnt:
		moves = game.soldierAntMoves(fromTile.Position)
	case PieceTypeSpider:
		moves = game.spiderMoves(fromTile.Position)
	case PieceTypeGrasshopper:
		moves = game.grasshopperMoves(fromTile.Position)
	case PieceTypeLadybug:
		moves = game.ladybugMoves(fromTile.Position)
	case PieceTypeBeetle:
		moves = game.beetleMoves(fromTile.Position)
	case PieceTypeMosquito:
		moves = game.mosquitoMoves(fromTile.Position)
	default:
		panic("unhandled case")
	}

	if _, ok := moves[to]; !ok {
		// not reachable via the rules
		return false
	}

	fromTile.StackHeight = game.nextStackHeight(to)
	fromTile.Position = to
	game.incrementMove()
	return true
}

func (game *HiveGame) incrementMove() {
	if game.ColorToMove == ColorWhite {
		game.Move++
	}

	if game.ColorToMove == ColorBlack {
		game.ColorToMove = ColorWhite
	} else {
		game.ColorToMove = ColorBlack
	}

}

func (game *HiveGame) tileAt(position HexVectorInt) *HiveTile {
	greatestStackHeight := -1
	var found *HiveTile = nil

	for i, tile := range game.Tiles {
		if tile.Position == position && tile.StackHeight > greatestStackHeight {
			greatestStackHeight = tile.StackHeight
			found = &game.Tiles[i]
		}
	}

	return found
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

func (game *HiveGame) isTilePinned(gameTile *HiveTile) bool {
	if gameTile.StackHeight != game.tileAt(gameTile.Position).StackHeight {
		return true
	} else if gameTile.StackHeight > 0 {
		return false
	}

	type searchNode struct {
		position    HexVectorInt
		stackHeight int
	}

	var neighbour *searchNode = nil
	for _, tile := range game.Tiles {
		for _, adj := range gameTile.Position.AdjacentVectors() {
			if tile.Position == adj {
				neighbour = &searchNode{position: tile.Position, stackHeight: tile.StackHeight}
				break
			}
		}
	}

	if neighbour == nil {
		// definitely an edge case, but it is technically not pinned if it has no neighbours
		return false
	}

	seen := make(map[searchNode]bool)
	toExplore := make([]searchNode, 0)

	toExplore = append(toExplore, *neighbour)

	for len(toExplore) > 0 {
		node := toExplore[0]
		toExplore = toExplore[1:]

		if _, ok := seen[node]; ok {
			continue
		}

		neighbours := make([]searchNode, 0, 6)
		for _, adj := range node.position.AdjacentVectors() {
			if adj == gameTile.Position {
				continue
			}

			for _, tile := range game.Tiles {
				if tile.Position == adj {
					neighbours = append(neighbours, searchNode{tile.Position, tile.StackHeight})
				}
			}

		}

		toExplore = append(toExplore, neighbours...)

		seen[node] = true
	}

	return len(seen) != len(game.Tiles)-1
}

func (game *HiveGame) soldierAntMoves(from HexVectorInt) map[HexVectorInt]bool {
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

func (game *HiveGame) queenBeeMoves(from HexVectorInt) map[HexVectorInt]bool {
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

func (game *HiveGame) grasshopperMoves(from HexVectorInt) map[HexVectorInt]bool {
	adjacentTiles := from.AdjacentVectors()
	neighbours := make([]HexVectorInt, 0, 6)
	for _, tile := range game.Tiles {
		for _, adj := range adjacentTiles {
			if adj == tile.Position {
				neighbours = append(neighbours, adj)
				break
			}
		}
	}

	moves := make(map[HexVectorInt]bool)

	for _, neighbour := range neighbours {
		direction := neighbour.Subtract(from)
		const LoopMax = 26 // computed as the greatest distance a hopper could ever jump + 1
		exitedEarly := false

		// we have already implicitly checked i = 1 by finding 'neighbour'
		for i := 2; i <= LoopMax; i++ {
			toCheck := direction.MultiplyScalar(i).Add(from)

			occupied := false
			for _, tile := range game.Tiles {
				if tile.Position == toCheck {
					occupied = true
					break
				}
			}

			if !occupied {
				exitedEarly = true
				moves[toCheck] = true
				break
			}
		}

		if !exitedEarly {
			panic("in no case should a hopper jump more than 25 tiles")
		}
	}

	return moves
}

func (game *HiveGame) ladybugMoves(from HexVectorInt) map[HexVectorInt]bool {
	// do the same as the spider but enforce that it must be on top of another tile for the first
	// two moves

	type searchNode struct {
		previous []HexVectorInt
		position HexVectorInt
	}

	search := make([]searchNode, 0)
	search = append(search, searchNode{previous: make([]HexVectorInt, 0), position: from})

	for moveIndex := 0; moveIndex < 2; moveIndex++ {
		for nodeIndex := len(search) - 1; nodeIndex >= 0; nodeIndex-- {
			node := search[nodeIndex]
			potentialNextMoves := make([]HexVectorInt, 0, 6)

			for _, adj := range node.position.AdjacentVectors() {
				isOccupied := false

				for _, tile := range game.Tiles {
					if tile.Position == adj {
						isOccupied = true
						break
					}
				}

				if isOccupied {
					potentialNextMoves = append(potentialNextMoves, adj)
				}
			}

			for _, move := range potentialNextMoves {
				alreadyMovedHere := false
				for _, pastPosition := range node.previous {
					if pastPosition == move {
						alreadyMovedHere = true
						break
					}
				}

				if !alreadyMovedHere {
					newNode := searchNode{previous: make([]HexVectorInt, 0, 2), position: move}

					for _, previousMove := range node.previous {
						newNode.previous = append(newNode.previous, previousMove)
					}

					newNode.previous = append(newNode.previous, node.position)

					search = append(search, newNode)
				}
			}

			search = append(search[:nodeIndex], search[nodeIndex+1:]...)
		}
	}

	// now search for unoccupied positions
	validMoves := make(map[HexVectorInt]bool)
	for _, node := range search {
		potentialNextMoves := make([]HexVectorInt, 0, 6)
		for _, adj := range node.position.AdjacentVectors() {
			isOccupied := false

			for _, tile := range game.Tiles {
				if tile.Position == adj {
					isOccupied = true
					break
				}
			}

			if !isOccupied {
				potentialNextMoves = append(potentialNextMoves, adj)
			}
		}

		// because the current position is "stacked", anything here is a valid move
		for _, move := range potentialNextMoves {
			validMoves[move] = true
		}
	}

	return validMoves
}

func (game *HiveGame) beetleMoves(from HexVectorInt) map[HexVectorInt]bool {
	validMoves := make(map[HexVectorInt]bool)

	greatestStackHeight := -1
	var beetleOnStack *HiveTile = nil

	for i, tile := range game.Tiles {
		if tile.Color == game.ColorToMove && tile.PieceType == PieceTypeBeetle {
			if tile.StackHeight > greatestStackHeight {
				greatestStackHeight = tile.StackHeight
				beetleOnStack = &game.Tiles[i]
			}
		}
	}

	if beetleOnStack == nil {
		return validMoves
	}

	for _, adj := range from.AdjacentVectors() {
		isOccupied := false

		for _, tile := range game.Tiles {
			if tile.Position == adj {
				isOccupied = true
				break
			}
		}

		if isOccupied {
			validMoves[adj] = true
		}
	}

	for move := range validMoves {
		validMoves[Rotate60().Transform(move.Subtract(from)).Add(from)] = true
		validMoves[Rotate300().Transform(move.Subtract(from)).Add(from)] = true
	}

	return validMoves
}

func (game *HiveGame) nextStackHeight(position HexVectorInt) int {
	greatestStackHeight := -1

	for _, tile := range game.Tiles {
		if tile.Position == position && tile.StackHeight > greatestStackHeight {
			greatestStackHeight = tile.StackHeight
		}
	}

	return greatestStackHeight + 1
}

func (game *HiveGame) mosquitoMoves(from HexVectorInt) map[HexVectorInt]bool {
	neighbourTiles := make([]HiveTile, 0, 6)

	for _, adj := range from.AdjacentVectors() {
		for _, tile := range game.Tiles {
			if tile.Position == adj {
				neighbourTiles = append(neighbourTiles, tile)
			}
		}
	}

	validMoves := make(map[HexVectorInt]bool)

	addAll := func(dest, src map[HexVectorInt]bool) {
		for k, v := range src {
			dest[k] = v
		}
	}

	for _, tile := range neighbourTiles {
		switch tile.PieceType {
		case PieceTypeQueenBee:
			addAll(validMoves, game.queenBeeMoves(from))
		case PieceTypeSoldierAnt:
			addAll(validMoves, game.soldierAntMoves(from))
		case PieceTypeGrasshopper:
			addAll(validMoves, game.grasshopperMoves(from))
		case PieceTypeSpider:
			addAll(validMoves, game.spiderMoves(from))
		case PieceTypeBeetle:
			addAll(validMoves, game.beetleMoves(from))
		case PieceTypeLadybug:
			addAll(validMoves, game.ladybugMoves(from))
		case PieceTypeMosquito:
			continue
		default:
			panic("unhandled case")
		}
	}

	return validMoves
}
