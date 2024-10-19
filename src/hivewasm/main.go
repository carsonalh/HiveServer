//go:build js && wasm

package main

import (
	"syscall/js"
)

type HexMatrixInt struct {
	A00, A01, A10, A11 int
}

type HexVectorInt struct {
	Q, R int
}

func (v HexVectorInt) Add(u HexVectorInt) HexVectorInt {
	return HexVectorInt{Q: v.Q + u.Q, R: v.R + u.R}
}

func (tile *HiveTile) FromJsValue(value js.Value) bool {
	if value.Type() != js.TypeObject {
		return false
	}

	color, ok := HiveColorFromJsValue(value.Get("color"))
	if !ok {
		return false
	}

	pieceType, ok := HivePieceTypeFromJsValue(value.Get("pieceType"))
	if !ok {
		return false
	}

	ok = tile.position.FromJsValue(value.Get("position"))
	if !ok {
		return false
	}

	tile.color = color
	tile.pieceType = pieceType

	return true
}

func (v HexVectorInt) Subtract(u HexVectorInt) HexVectorInt {
	return HexVectorInt{Q: v.Q - u.Q, R: v.R - u.R}
}

func (m HexMatrixInt) Transform(v HexVectorInt) HexVectorInt {
	return HexVectorInt{
		Q: m.A00*v.Q + m.A01*v.R,
		R: m.A10*v.Q + m.A11*v.R,
	}
}

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

func (tile *HiveTile) FromJsValue(value js.Value) bool {
	if value.Type() != js.TypeObject {
		return false
	}

	color, ok := HiveColorFromJsValue(value.Get("color"))
	if !ok {
		return false
	}

	pieceType, ok := HivePieceTypeFromJsValue(value.Get("pieceType"))
	if !ok {
		return false
	}

	ok = tile.position.FromJsValue(value.Get("position"))
	if !ok {
		return false
	}

	tile.color = color
	tile.pieceType = pieceType

	return true
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

func (game *HiveGame) ToJsValue() js.Value {
	tilesAsJsInterfaceSlice := make([]interface{}, 0, len(game.tiles))
	for _, tile := range game.tiles {
		tilesAsJsInterfaceSlice = append(tilesAsJsInterfaceSlice, map[string]interface{}{
			"color": tile.color,
			"position": map[string]interface{}{
				"q": tile.position.Q,
				"r": tile.position.R,
			},
			"pieceType": tile.pieceType,
		})
	}
	return js.ValueOf(map[string]interface{}{
		"colorToMove": game.colorToMove,
		"move":        game.move,
		"tiles":       tilesAsJsInterfaceSlice,
	})
}

func JsValueToInt(value js.Value) (int, bool) {
	if value.Type() != js.TypeNumber {
		return 0, false
	}

	isInteger := js.Global().Get("Number").Call("isInteger", value).Bool()
	if !isInteger {
		return 0, false
	}

	return value.Int(), true
}

func HiveColorFromJsValue(value js.Value) (HiveColor, bool) {
	x, ok := JsValueToInt(value)

	if !ok {
		return 0, false
	}

	switch x {
	case ColorBlack:
		fallthrough
	case ColorWhite:
		return x, true
	}

	return 0, false
}

func HivePieceTypeFromJsValue(value js.Value) (HivePieceType, bool) {
	x, ok := JsValueToInt(value)

	if !ok {
		return 0, false
	}

	switch x {
	case PieceTypeQueenBee:
	case PieceTypeSoldierAnt:
	case PieceTypeGrasshopper:
	case PieceTypeSpider:
	case PieceTypeBeetle:
	case PieceTypeLadybug:
	case PieceTypeMosquito:
		return x, true
	}

	return 0, false
}

func (v *HexVectorInt) FromJsValue(value js.Value) bool {
	rawQ := value.Get("q")
	rawR := value.Get("r")

	if rawQ.Type() != js.TypeNumber || rawR.Type() != js.TypeNumber {
		return false
	}

	v.Q = rawQ.Int()
	v.R = rawR.Int()

	return true
}

func (game *HiveGame) FromJsValue(value js.Value) bool {
	if value.Type() != js.TypeObject {
		return false
	}

	colorToMove, ok := HiveColorFromJsValue(value.Get("colorToMove"))

	if !ok {
		return false
	}

	move, ok := JsValueToInt(value.Get("move"))

	if !ok {
		return false
	}

	tiles := value.Get("tiles")

	if tiles.Type() != js.TypeObject || !tiles.InstanceOf(js.Global().Get("Array")) {
		return false
	}

	tilesLength := tiles.Get("length").Int()

	parsedTiles := make([]HiveTile, tilesLength)

	for i := range tilesLength {
		ok := parsedTiles[i].FromJsValue(tiles.Index(i))
		if !ok {
			return false
		}
	}

	game.colorToMove = colorToMove
	game.move = move
	game.tiles = parsedTiles

	return true
}

func createHiveGame(this js.Value, args []js.Value) interface{} {
	game := CreateHiveGame()
	return game.ToJsValue()
}

func placeTile(this js.Value, args []js.Value) interface{} {
	game := HiveGame{}

	ok := game.FromJsValue(args[0])

	if !ok {
		panic("could not parse js value to hive game")
	}

	game.PlaceTile(HexVectorInt{}, PieceTypeQueenBee)

	return game.ToJsValue()
}

func main() {
	hiveModule := js.Global().Get("Object").New()
	hiveModule.Set("createHiveGame", js.FuncOf(createHiveGame))
	hiveModule.Set("placeTile", js.FuncOf(placeTile))
	js.Global().Set("hive", hiveModule)
	select {}
}
