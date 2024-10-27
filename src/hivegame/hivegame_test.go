package hivegame

import "testing"

func TestCreateHiveGame(t *testing.T) {
	game := CreateHiveGame()

	if game.Move != 1 {
		t.Fatalf("CreateHiveGame failed. Expected to initialise with 1 Move, got %d Move(s)", game.Move)
	}

	if len(game.Tiles) != 0 {
		t.Fatalf("Game must be initialized with zero Tiles")
	}

	if game.ColorToMove != ColorBlack {
		t.Fatalf("A hive game always has black to Move first")
	}
}

func TestPlacesTheFirstPiece(t *testing.T) {
	game := CreateHiveGame()

	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee)

	if len(game.Tiles) != 1 {
		t.Fatalf("Expected to have one tile that was successfully placed into the game")
	}
}

func TestAlternatesBetweenBlackAndWhite(t *testing.T) {
	game := CreateHiveGame()

	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{1, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeGrasshopper)

	results := []struct{ actual, expected HiveColor }{
		{game.Tiles[0].Color, ColorBlack},
		{game.Tiles[1].Color, ColorWhite},
		{game.Tiles[2].Color, ColorBlack},
		{game.Tiles[3].Color, ColorWhite},
	}

	for i, result := range results {
		if result.actual != result.expected {
			t.Fatalf("Mismatch in expected Color on Move %d", i+1)
		}
	}
}

func TestCannotPlacePiecesAtopOthers(t *testing.T) {
	game := CreateHiveGame()

	var ok bool

	ok = game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	if !ok {
		t.Fatalf("Falsely flagged bad placement for initial Move")
	}

	ok = game.PlaceTile(HexVectorInt{0, 0}, PieceTypeBeetle)
	if ok {
		t.Fatalf("Cannot place Tiles atop other Tiles")
	}
}

func TestEnsuresQueenPlacedByMove4(t *testing.T) {
	game := CreateHiveGame()

	checkOk := func(ok bool) {
		if !ok {
			t.Fatalf("falsely failed a valid placement")
		}
	}

	var ok bool

	// Move 1
	ok = game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	checkOk(ok)
	ok = game.PlaceTile(HexVectorInt{1, 0}, PieceTypeGrasshopper)
	checkOk(ok)

	// Move 2
	ok = game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper)
	checkOk(ok)
	ok = game.PlaceTile(HexVectorInt{2, 0}, PieceTypeGrasshopper)
	checkOk(ok)

	// Move 3
	ok = game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeGrasshopper)
	checkOk(ok)
	ok = game.PlaceTile(HexVectorInt{3, 0}, PieceTypeGrasshopper)
	checkOk(ok)

	// Move 4
	ok = game.PlaceTile(HexVectorInt{-3, 0}, PieceTypeGrasshopper)
	if ok {
		t.Fatalf("Cannot pass this Move as the queen should have been placed (black)")
	}

	ok = game.PlaceTile(HexVectorInt{-3, 0}, PieceTypeQueenBee)
	checkOk(ok)

	ok = game.PlaceTile(HexVectorInt{4, 0}, PieceTypeGrasshopper)
	if ok {
		t.Fatalf("Cannot pass this Move as the queen should have been placed (white)")
	}

	ok = game.PlaceTile(HexVectorInt{4, 0}, PieceTypeQueenBee)
	checkOk(ok)
}

func TestCannotPlaceMorePiecesThanPlayerHas(t *testing.T) {
	game := CreateHiveGame()
	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{1, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{2, 0}, PieceTypeQueenBee)
	game.PlaceTile(HexVectorInt{-3, 0}, PieceTypeQueenBee)
	game.PlaceTile(HexVectorInt{3, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-4, 0}, PieceTypeGrasshopper)

	ok := game.PlaceTile(HexVectorInt{4, 0}, PieceTypeGrasshopper)
	if ok {
		t.Fatalf("Cannot allow a player to place more than three grasshoppers")
	}
}

func TestFollowsAdjacencyRulesForPlacement(t *testing.T) {
	game := CreateHiveGame()

	var ok bool

	ok = game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	if !ok {
		t.Fatalf("First Move need not follow the normal rules")
	}

	ok = game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper)
	if !ok {
		t.Fatalf("Second Move need not follow the normal rules")
	}

	ok = game.PlaceTile(HexVectorInt{0, -1}, PieceTypeGrasshopper)
	if ok {
		t.Fatalf("Should not be able to place a piece that touches the opposite Color")
	}

	// this touches nothing
	ok = game.PlaceTile(HexVectorInt{0, 2}, PieceTypeGrasshopper)
	if ok {
		t.Fatalf("A piece must be touching one if its own")
	}
}

func TestMoveAntAroundTheHive(t *testing.T) {
	canMove := func(from, to HexVectorInt) bool {
		game := CreateHiveGame()

		game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee)    // black
		game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee)   // white
		game.PlaceTile(HexVectorInt{1, 0}, PieceTypeSoldierAnt)  // black
		game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeSoldierAnt) // white

		return game.MoveTile(from, to)
	}

	blackAntLegalMoves := []HexVectorInt{
		{1, -1},
		{0, -1},
		{-1, -1},
		{-2, -1},
		{-3, 0},
		{-3, 1},
		{-2, 1},
		{-1, 1},
		{0, 1},
	}

	blackAntPosition := HexVectorInt{1, 0}

	for _, newPosition := range blackAntLegalMoves {
		if !canMove(blackAntPosition, newPosition) {
			t.Fatalf("Failed moving ant from %+v to %+v", blackAntPosition, newPosition)
		}
	}

	blackAntIllegalMoves := []HexVectorInt{
		blackAntPosition,
		{3, 7},
	}

	for _, newPosition := range blackAntIllegalMoves {
		if canMove(blackAntPosition, newPosition) {
			t.Fatalf("Falsely allowed to move ant from %+v to %+v", blackAntPosition, newPosition)
		}
	}
}

func TestRespectsFreedomToMove(t *testing.T) {
	game := CreateHiveGame()

	// setup the space for an illegal freedom to move
	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee)
	game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee)
	game.PlaceTile(HexVectorInt{1, -1}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-1, -1}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{1, -2}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeSoldierAnt)
	game.PlaceTile(HexVectorInt{1, 0}, PieceTypeGrasshopper)

	// try and pull off the illegal move violating freedom to move
	if game.MoveTile(HexVectorInt{-2, 0}, HexVectorInt{0, -1}) {
		t.Fatalf("Allowed ant to violate Freedom to Move")
	}
}
