package hivegame

import "testing"

func TestCreateHiveGame(t *testing.T) {
	game := CreateHiveGame()

	if game.move != 1 {
		t.Fatalf("CreateHiveGame failed. Expected to initialise with 1 move, got %d move(s)", game.move)
	}

	if len(game.tiles) != 0 {
		t.Fatalf("Game must be initialized with zero tiles")
	}

	if game.colorToMove != ColorBlack {
		t.Fatalf("A hive game always has black to move first")
	}
}

func TestPlaceTile(t *testing.T) {
	game := CreateHiveGame()

	game.PlaceTile(HexVectorInt{0, 0}, PieceTypeQueenBee)

	if len(game.tiles) != 1 {
		t.Fatalf("Expected to have one tile that was successfully placed into the game")
	}
}
