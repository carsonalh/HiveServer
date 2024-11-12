package hivegame

import (
	"github.com/go-test/deep"
	"testing"
)

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

func TestOneHiveRule(t *testing.T) {
	game := CreateHiveGame()

	expectLegal := func(ret bool) {
		if !ret {
			t.Fatalf("Did not allow a legal move")
		}
	}

	expectLegal(game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{1, 0}, PieceTypeSoldierAnt))
	expectLegal(game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeSoldierAnt))

	if ok := game.MoveTile(HexVectorInt{0, 0}, HexVectorInt{1, -1}); ok {
		t.Fatalf("Allowed one-hive rule violation")
	}

	expectLegal(game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{1, -1}))

	if ok := game.MoveTile(HexVectorInt{-2, 0}, HexVectorInt{-2, -1}); ok {
		t.Fatalf("Allowed one-hive rule violation")
	}

	expectLegal(game.MoveTile(HexVectorInt{-2, 0}, HexVectorInt{-2, 1}))

	expectedState := []HiveTile{
		{PieceType: PieceTypeQueenBee, Position: HexVectorInt{0, 0}, Color: ColorBlack},
		{PieceType: PieceTypeQueenBee, Position: HexVectorInt{-1, 0}, Color: ColorWhite},
		{PieceType: PieceTypeSoldierAnt, Position: HexVectorInt{1, -1}, Color: ColorBlack},
		{PieceType: PieceTypeSoldierAnt, Position: HexVectorInt{-2, 1}, Color: ColorWhite},
	}

	tiles := game.Tiles

	diff := deep.Equal(expectedState, tiles)
	if diff != nil {
		t.Fatalf("Mismatched game state: %v", diff)
	}
}

func TestCannotMovePiecesOfOppositeColor(t *testing.T) {
	game := CreateHiveGame()

	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{1, 0}, PieceTypeQueenBee)
	game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeQueenBee)

	if ok := game.MoveTile(HexVectorInt{-2, 0}, HexVectorInt{-2, 1}); ok {
		t.Fatalf("Allowed black to move a white piece")
	}
}

func TestMoveQueenBee(t *testing.T) {
	game := CreateHiveGame()

	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper)
	game.PlaceTile(HexVectorInt{1, 0}, PieceTypeQueenBee)
	game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeQueenBee)

	var ok bool

	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{-3, 0})
	if ok {
		t.Fatalf("Allowed queen to move more than one tile")
	}

	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{2, 0})
	if ok {
		t.Fatalf("Allowed queen to move off the hive")
	}

	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{1, -1})
	if !ok {
		t.Fatalf("Did not allow valid queen move")
	}
}

func TestMoveSpider(t *testing.T) {
	initGame := func() HiveGame {
		game := CreateHiveGame()

		game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee)
		game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee)
		game.PlaceTile(HexVectorInt{1, 0}, PieceTypeSpider)
		game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeSoldierAnt)

		return game
	}

	var ok bool
	game := initGame()
	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{1, -1})
	if ok {
		t.Fatalf("Falsely allowed spider to move one space")
	}
	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{0, 1})
	if ok {
		t.Fatalf("Falsely allowed spider to move one space")
	}
	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{0, -1})
	if ok {
		t.Fatalf("Falsely allowed spider to move two spaces")
	}
	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{-1, 1})
	if ok {
		t.Fatalf("Falsely allowed spider to move two spaces")
	}
	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{-2, -1})
	if ok {
		t.Fatalf("Falsely allowed spider to move four spaces")
	}
	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{-3, 1})
	if ok {
		t.Fatalf("Falsely allowed spider to move four spaces")
	}

	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{-2, 1})
	if !ok {
		t.Fatalf("Would not let spider move three spaces")
	}

	game = initGame()

	ok = game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{-1, -1})
	if !ok {
		t.Fatalf("Would not let spider move three spaces")
	}
}

func TestMoveGrasshopper(t *testing.T) {
	initGame := func() HiveGame {
		game := CreateHiveGame()

		game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee)
		game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee)
		game.PlaceTile(HexVectorInt{1, 0}, PieceTypeSoldierAnt)
		game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeSoldierAnt)
		game.PlaceTile(HexVectorInt{0, 1}, PieceTypeGrasshopper)

		game.MoveTile(HexVectorInt{-2, 0}, HexVectorInt{-1, 1})

		return game
	}

	game := initGame()
	if ok := game.MoveTile(HexVectorInt{0, 1}, HexVectorInt{1, 1}); ok {
		t.Fatalf("Falsely allowed grasshopper to move to adjacent tile")
	}
	if ok := game.MoveTile(HexVectorInt{0, 1}, HexVectorInt{-1, 2}); ok {
		t.Fatalf("Falsely allowed grasshopper to move to adjacent tile")
	}

	if ok := game.MoveTile(HexVectorInt{0, 1}, HexVectorInt{0, -1}); !ok {
		t.Fatalf("Failed to allow to move grasshopper legally")
	}

	game = initGame()
	if ok := game.MoveTile(HexVectorInt{0, 1}, HexVectorInt{2, -1}); !ok {
		t.Fatalf("Failed to allow to move grasshopper legally")
	}

	game = initGame()
	if ok := game.MoveTile(HexVectorInt{0, 1}, HexVectorInt{-2, 1}); !ok {
		t.Fatalf("Failed to allow to move grasshopper legally")
	}
}

func TestMoveLadybug(t *testing.T) {
	exampleFromRulebookP7 := func() HiveGame {
		return HiveGame{
			ColorToMove: ColorWhite,
			Move:        6,
			Tiles: []HiveTile{
				{Color: ColorBlack, Position: HexVectorInt{0, 0}, PieceType: PieceTypeBeetle},
				{Color: ColorBlack, Position: HexVectorInt{0, -2}, PieceType: PieceTypeQueenBee},
				{Color: ColorWhite, Position: HexVectorInt{-1, 0}, PieceType: PieceTypeBeetle},
				{Color: ColorWhite, Position: HexVectorInt{-1, -1}, PieceType: PieceTypeGrasshopper},
				{Color: ColorWhite, Position: HexVectorInt{1, -1}, PieceType: PieceTypeQueenBee},
				{Color: ColorWhite, Position: HexVectorInt{-1, 1}, PieceType: PieceTypeLadybug},
			},
		}
	}

	ladybugPosition := HexVectorInt{-1, 1}
	legalMoves := []HexVectorInt{
		{-2, 0},
		{-2, 1},
		{-2, -1},
		{-1, -2},
		{0, -1},
		{1, -2},
		{2, -2},
		{2, -1},
		{1, 0},
		{0, 1},
	}

	illegalMoves := []HexVectorInt{
		{0, -3},
		{1, -3},
		{-1, 2},
		{-2, 2},
	}

	for _, legalMove := range legalMoves {
		game := exampleFromRulebookP7()

		if ok := game.MoveTile(ladybugPosition, legalMove); !ok {
			t.Fatalf("Did not allow legal move from %v to %v", ladybugPosition, legalMove)
		}
	}

	game := exampleFromRulebookP7()
	for _, illegalMove := range illegalMoves {
		if ok := game.MoveTile(ladybugPosition, illegalMove); ok {
			t.Fatalf("Incorrectly allowed illegal move from %v to %v", ladybugPosition, illegalMove)
		}
	}
}

func TestMoveBeetle(t *testing.T) {
	exampleFromRulebookP4 := func() HiveGame {
		return HiveGame{
			ColorToMove: ColorWhite,
			Move:        6,
			Tiles: []HiveTile{
				{Color: ColorWhite, Position: HexVectorInt{1, -1}, PieceType: PieceTypeBeetle},
				{Color: ColorWhite, Position: HexVectorInt{0, 0}, PieceType: PieceTypeSoldierAnt},
				{Color: ColorWhite, Position: HexVectorInt{0, -1}, PieceType: PieceTypeSpider},
				{Color: ColorWhite, Position: HexVectorInt{-1, 0}, PieceType: PieceTypeQueenBee},
				{Color: ColorBlack, Position: HexVectorInt{0, 1}, PieceType: PieceTypeQueenBee},
				{Color: ColorBlack, Position: HexVectorInt{1, 1}, PieceType: PieceTypeGrasshopper},
			},
		}
	}

	legalMoves := []HexVectorInt{
		{0, 0},
		{1, 0},
		{0, -1},
		{1, -2},
	}

	illegalMoves := []HexVectorInt{
		{2, 0},
		{0, -2},
	}

	beetlePosition := HexVectorInt{1, -1}

	for _, legalMove := range legalMoves {
		game := exampleFromRulebookP4()

		if ok := game.MoveTile(beetlePosition, legalMove); !ok {
			t.Fatalf("did not allow legal move from %v to %v", beetlePosition, legalMove)
		}
	}

	game := exampleFromRulebookP4()
	for _, illegalMove := range illegalMoves {
		if ok := game.MoveTile(beetlePosition, illegalMove); ok {
			t.Fatalf("Incorrectly allowed illegal move from %v to %v", beetlePosition, illegalMove)
		}
	}
}

func TestMoveMosquito(t *testing.T) {
	exampleFromRulebookP8 := func() HiveGame {
		return HiveGame{
			ColorToMove: ColorWhite,
			Move:        6,
			Tiles: []HiveTile{
				{Color: ColorWhite, Position: HexVectorInt{0, 0}, PieceType: PieceTypeBeetle},
				{Color: ColorWhite, Position: HexVectorInt{1, 0}, PieceType: PieceTypeQueenBee},
				{Color: ColorWhite, Position: HexVectorInt{-1, 1}, PieceType: PieceTypeMosquito},
				{Color: ColorBlack, Position: HexVectorInt{-1, 0}, PieceType: PieceTypeSpider},
				{Color: ColorBlack, Position: HexVectorInt{1, -1}, PieceType: PieceTypeQueenBee},
			},
		}
	}

	legalMoves := []HexVectorInt{
		{0, 0},
		{-1, 0},
		{-1, -1},
		{-2, 1},
		{0, 1},
		{2, 0},
	}

	illegalMoves := []HexVectorInt{
		{-2, 0},
		{0, -1},
		{2, -1},
		{1, 0},
	}

	mosquitoPosition := HexVectorInt{-1, 1}

	for _, legalMove := range legalMoves {
		game := exampleFromRulebookP8()

		if ok := game.MoveTile(mosquitoPosition, legalMove); !ok {
			t.Fatalf("Did not allow legal move from %v to %v", mosquitoPosition, legalMove)
		}
	}

	game := exampleFromRulebookP8()
	for _, illegalMove := range illegalMoves {
		if ok := game.MoveTile(mosquitoPosition, illegalMove); ok {
			t.Fatalf("Incorrectly allowed illegal move from %v to %v", mosquitoPosition, illegalMove)
		}
	}
}

func TestBeetleStack(t *testing.T) {
	expectLegal := func(ret bool) {
		if !ret {
			t.Fatalf("Incorrectly failed to make a legal move")
		}
	}

	game := CreateHiveGame()

	expectLegal(game.PlaceTile(HexVectorInt{0, 0}, PieceTypeGrasshopper))
	expectLegal(game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeGrasshopper))
	expectLegal(game.PlaceTile(HexVectorInt{1, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeQueenBee))

	expectLegal(game.PlaceTile(HexVectorInt{1, -1}, PieceTypeBeetle))
	expectLegal(game.PlaceTile(HexVectorInt{-2, 1}, PieceTypeBeetle))

	expectLegal(game.MoveTile(HexVectorInt{1, -1}, HexVectorInt{0, 0}))
	expectLegal(game.MoveTile(HexVectorInt{-2, 1}, HexVectorInt{-1, 0}))

	expectLegal(game.MoveTile(HexVectorInt{0, 0}, HexVectorInt{-1, 0}))

	expectLegal(game.PlaceTile(HexVectorInt{-3, 0}, PieceTypeGrasshopper))
	if ok := game.PlaceTile(HexVectorInt{0, -1}, PieceTypeSpider); !ok {
		t.Fatalf("Failure to place next to a beetle stack with same color on top")
	}

	if ok := game.MoveTile(HexVectorInt{-1, 0}, HexVectorInt{0, 0}); ok {
		t.Fatalf("Tried to move beetle under top of stack")
	}
}

func TestStackHeightsAreUpdated(t *testing.T) {
	game := CreateHiveGame()

	expectLegal := func(ret bool) {
		if !ret {
			t.Fatalf("Incorrectly failed to make a legal move")
		}
	}

	expectLegal(game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{1, 0}, PieceTypeBeetle))
	expectLegal(game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeBeetle))

	game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{0, 0})

	if game.Tiles[2].StackHeight != 1 {
		t.Fatalf("The beetle should have had its stack height updated")
	}

	for i := 0; i < 4; i++ {
		if i != 2 && game.Tiles[i].StackHeight != 0 {
			t.Fatalf("Some other tile at idx. %d has stack height %d", i, game.Tiles[i].StackHeight)
		}
	}
}

func TestStackMosquitos(t *testing.T) {
	game := CreateHiveGame()

	expectLegal := func(ret bool) {
		if !ret {
			t.Fatalf("Incorrectly failed to make a legal move")
		}
	}

	expectLegal(game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{-1, 0}, PieceTypeQueenBee))
	expectLegal(game.PlaceTile(HexVectorInt{1, 0}, PieceTypeBeetle))
	expectLegal(game.PlaceTile(HexVectorInt{-2, 0}, PieceTypeMosquito))

	expectLegal(game.MoveTile(HexVectorInt{1, 0}, HexVectorInt{0, 0}))
	expectLegal(game.MoveTile(HexVectorInt{-2, 0}, HexVectorInt{-1, -1}))
	expectLegal(game.MoveTile(HexVectorInt{0, 0}, HexVectorInt{-1, 0}))

	if ok := game.MoveTile(HexVectorInt{-1, -1}, HexVectorInt{-1, 0}); !ok {
		t.Fatalf("Failed to let mosquito behave like a beetle")
	}

	if game.Tiles[3].StackHeight != 2 {
		t.Fatalf("Did not stack the mosquito on top of the beetle")
	}
}

func TestWinCondition(t *testing.T) { t.Skip() }
