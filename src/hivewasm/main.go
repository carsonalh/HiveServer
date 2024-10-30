//go:build js && wasm

package main

import (
	"HiveServer/src/hivegame"
	"fmt"
	"syscall/js"
)

var pieceTypeStrings = map[hivegame.HivePieceType]string{
	hivegame.PieceTypeQueenBee:    "QUEEN_BEE",
	hivegame.PieceTypeSoldierAnt:  "SOLDIER_ANT",
	hivegame.PieceTypeGrasshopper: "GRASSHOPPER",
	hivegame.PieceTypeSpider:      "SPIDER",
	hivegame.PieceTypeBeetle:      "BEETLE",
	hivegame.PieceTypeLadybug:     "LADYBUG",
	hivegame.PieceTypeMosquito:    "MOSQUITO",
}

func HiveGameToJsValue(game hivegame.HiveGame) js.Value {
	tilesAsJsInterfaceSlice := make([]interface{}, 0, len(game.Tiles))
	for _, tile := range game.Tiles {
		tilesAsJsInterfaceSlice = append(tilesAsJsInterfaceSlice, map[string]interface{}{
			"color": tile.Color,
			"position": map[string]interface{}{
				"q": tile.Position.Q,
				"r": tile.Position.R,
			},
			"pieceType":   tile.PieceType,
			"stackHeight": tile.StackHeight,
		})
	}

	jsifiedBlackReserve := make(map[string]interface{})
	for piece, count := range game.BlackReserve {
		jsifiedBlackReserve[pieceTypeStrings[piece]] = count
	}

	jsifiedWhiteReserve := make(map[string]interface{})
	for piece, count := range game.WhiteReserve {
		jsifiedWhiteReserve[pieceTypeStrings[piece]] = count
	}

	return js.ValueOf(map[string]interface{}{
		"colorToMove":  game.ColorToMove,
		"move":         game.Move,
		"tiles":        tilesAsJsInterfaceSlice,
		"blackReserve": jsifiedBlackReserve,
		"whiteReserve": jsifiedWhiteReserve,
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

func HiveColorFromJsValue(value js.Value) (hivegame.HiveColor, bool) {
	x, ok := JsValueToInt(value)

	if !ok {
		return 0, false
	}

	switch x {
	case hivegame.ColorBlack:
		fallthrough
	case hivegame.ColorWhite:
		return x, true
	}

	return 0, false
}

func JsValueToHexVectorInt(value js.Value) (hivegame.HexVectorInt, bool) {
	rawQ := value.Get("q")
	rawR := value.Get("r")

	if rawQ.Type() != js.TypeNumber || rawR.Type() != js.TypeNumber {
		return hivegame.HexVectorInt{}, false
	}

	v := hivegame.HexVectorInt{}

	v.Q = rawQ.Int()
	v.R = rawR.Int()

	return v, true
}

func JsValueToHivePieceType(value js.Value) (hivegame.HivePieceType, bool) {
	if value.Type() != js.TypeNumber {
		return 0, false
	}

	isInteger := js.Global().Get("Number").Call("isInteger", value).Bool()
	if !isInteger {
		return 0, false
	}

	switch asInteger := value.Int(); asInteger {
	case hivegame.PieceTypeQueenBee:
		fallthrough
	case hivegame.PieceTypeSoldierAnt:
		fallthrough
	case hivegame.PieceTypeGrasshopper:
		fallthrough
	case hivegame.PieceTypeSpider:
		fallthrough
	case hivegame.PieceTypeBeetle:
		fallthrough
	case hivegame.PieceTypeLadybug:
		fallthrough
	case hivegame.PieceTypeMosquito:
		return asInteger, true
	}

	return 0, false
}

func JsValueToHiveGame(value js.Value) (hivegame.HiveGame, error) {
	if value.Type() != js.TypeObject {
		return hivegame.HiveGame{}, fmt.Errorf("tried to parse non-object type")
	}

	colorToMove, ok := HiveColorFromJsValue(value.Get("colorToMove"))

	if !ok {
		return hivegame.HiveGame{}, fmt.Errorf("failed to parse hive color to move")
	}

	move, ok := JsValueToInt(value.Get("move"))

	if !ok {
		return hivegame.HiveGame{}, fmt.Errorf("failed to parse hive move")
	}

	tiles := value.Get("tiles")

	if tiles.Type() != js.TypeObject || !tiles.InstanceOf(js.Global().Get("Array")) {
		return hivegame.HiveGame{}, fmt.Errorf("failed to parse tiles array")
	}

	tilesLength := tiles.Get("length").Int()

	parsedTiles := make([]hivegame.HiveTile, tilesLength)

	for i := range tilesLength {
		//ok := parsedTiles[i].FromJsValue(tiles.Index(i))
		if tile, ok := JsValueToHiveTile(tiles.Index(i)); !ok {
			return hivegame.HiveGame{}, fmt.Errorf("failed to parse tile index %d", i)
		} else {
			parsedTiles[i] = tile
		}
	}

	pieceTypes := []hivegame.HivePieceType{
		hivegame.PieceTypeQueenBee,
		hivegame.PieceTypeSoldierAnt,
		hivegame.PieceTypeGrasshopper,
		hivegame.PieceTypeSpider,
		hivegame.PieceTypeBeetle,
		hivegame.PieceTypeLadybug,
		hivegame.PieceTypeMosquito,
	}

	blackReserve := make(map[hivegame.HivePieceType]int)
	blackReserveJsValue := value.Get("blackReserve")
	for _, pieceType := range pieceTypes {
		pieceString, ok := pieceTypeStrings[pieceType]

		if !ok {
			return hivegame.HiveGame{}, fmt.Errorf("failed to lookup piece name for enum const %d (black)", pieceType)
		}

		blackReserve[pieceType] = blackReserveJsValue.Get(pieceString).Int()
	}

	whiteReserve := make(map[hivegame.HivePieceType]int)
	whiteReserveJsValue := value.Get("whiteReserve")
	for _, pieceType := range pieceTypes {
		pieceString, ok := pieceTypeStrings[pieceType]

		if !ok {
			return hivegame.HiveGame{}, fmt.Errorf("failed to lookup piece name for enum const %d (white)", pieceType)
		}

		whiteReserve[pieceType] = whiteReserveJsValue.Get(pieceString).Int()
	}

	game := hivegame.HiveGame{}

	game.ColorToMove = colorToMove
	game.Move = move
	game.Tiles = parsedTiles
	game.BlackReserve = blackReserve
	game.WhiteReserve = whiteReserve

	return game, nil
}

func JsValueToHiveTile(value js.Value) (hivegame.HiveTile, bool) {
	if value.Type() != js.TypeObject {
		return hivegame.HiveTile{}, false
	}

	tile := hivegame.HiveTile{}

	if color, ok := HiveColorFromJsValue(value.Get("color")); !ok {
		return hivegame.HiveTile{}, false
	} else {
		tile.Color = color
	}

	if position, ok := JsValueToHexVectorInt(value.Get("position")); !ok {
		return hivegame.HiveTile{}, false
	} else {
		tile.Position = position
	}

	if pieceType, ok := JsValueToHivePieceType(value.Get("pieceType")); !ok {
		return hivegame.HiveTile{}, false
	} else {
		tile.PieceType = pieceType
	}

	return tile, true
}

func ExportEnumConstants(object js.Value) {
	object.Set("COLOR_BLACK", hivegame.ColorBlack)
	object.Set("COLOR_WHITE", hivegame.ColorWhite)

	object.Set("PIECE_TYPE_QUEEN_BEE", hivegame.PieceTypeQueenBee)
	object.Set("PIECE_TYPE_SOLDIER_ANT", hivegame.PieceTypeSoldierAnt)
	object.Set("PIECE_TYPE_SPIDER", hivegame.PieceTypeSpider)
	object.Set("PIECE_TYPE_GRASSHOPPER", hivegame.PieceTypeGrasshopper)
	object.Set("PIECE_TYPE_BEETLE", hivegame.PieceTypeBeetle)
	object.Set("PIECE_TYPE_LADYBUG", hivegame.PieceTypeLadybug)
	object.Set("PIECE_TYPE_MOSQUITO", hivegame.PieceTypeMosquito)
}

func createHiveGame(_ js.Value, _ []js.Value) interface{} {
	game := hivegame.CreateHiveGame()
	return HiveGameToJsValue(game)
}

func placeTile(_ js.Value, args []js.Value) interface{} {
	if len(args) != 3 {
		panic("placeTile function expects 3 arguments : game, piece type, position")
	}

	game, err := JsValueToHiveGame(args[0])

	if err != nil {
		panic(err)
	}

	pieceType, ok := JsValueToHivePieceType(args[1])

	if !ok {
		panic("could not parse js value to hive piece type")
	}

	position, ok := JsValueToHexVectorInt(args[2])

	if !ok {
		panic("could not parse js value to hex vector position")
	}

	ok = game.PlaceTile(position, pieceType)

	ret := js.Global().Get("Array").New()
	ret.Call("push", HiveGameToJsValue(game))
	ret.Call("push", js.ValueOf(ok))

	return ret
}

func moveTile(_ js.Value, args []js.Value) interface{} {
	if len(args) != 3 {
		panic("placeTile function expects 3 arguments : game, from position, to position")
	}

	game, err := JsValueToHiveGame(args[0])

	if err != nil {
		panic(err)
	}

	fromPosition, ok := JsValueToHexVectorInt(args[1])

	if !ok {
		panic("could not parse js value to hex vector fromPosition")
	}

	toPosition, ok := JsValueToHexVectorInt(args[2])

	if !ok {
		panic("could not parse js value to hex vector toPosition")
	}

	ok = game.MoveTile(fromPosition, toPosition)

	ret := js.Global().Get("Array").New()
	ret.Call("push", HiveGameToJsValue(game))
	ret.Call("push", js.ValueOf(ok))

	return ret
}

func legalMoves(_ js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		panic("placeTile function expects 2 arguments : game, position")
	}

	game, err := JsValueToHiveGame(args[0])

	if err != nil {
		panic(err)
	}

	position, ok := JsValueToHexVectorInt(args[1])

	if !ok {
		panic("could not parse js value to hex vector position")
	}

	moves := game.LegalMoves(position)

	jsableMoves := make([]map[string]interface{}, 0)

	for _, move := range moves {
		jsableMoves = append(jsableMoves, map[string]interface{}{
			"q": move.Q,
			"r": move.R,
		})
	}

	jsMoves := js.Global().Get("Array").New()

	for _, move := range jsableMoves {
		jsMoves.Call("push", js.ValueOf(move))
	}

	return jsMoves
}

func tiles(_ js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		panic("tiles function expects 1 argument : game")
	}

	game, err := JsValueToHiveGame(args[0])

	if err != nil {
		panic(err)
	}

	jsTiles := js.Global().Get("Array").New()

	for _, tile := range game.Tiles {
		jsTiles.Call("push", js.ValueOf(map[string]interface{}{
			"color": tile.Color,
			"position": js.ValueOf(map[string]interface{}{
				"q": tile.Position.Q,
				"r": tile.Position.R,
			}),
			"stackHeight": tile.StackHeight,
			"pieceType":   tile.PieceType,
		}),
		)
	}

	return jsTiles
}

func colorToMove(_ js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		panic("colorToMove function expects 1 argument : game")
	}

	game, err := JsValueToHiveGame(args[0])

	if err != nil {
		panic(err)
	}

	return js.ValueOf(game.ColorToMove)
}

func main() {
	hiveModule := js.Global().Get("Object").New()
	hiveModule.Set("createHiveGame", js.FuncOf(createHiveGame))
	hiveModule.Set("placeTile", js.FuncOf(placeTile))
	hiveModule.Set("moveTile", js.FuncOf(moveTile))
	hiveModule.Set("legalMoves", js.FuncOf(legalMoves))
	hiveModule.Set("tiles", js.FuncOf(tiles))
	hiveModule.Set("colorToMove", js.FuncOf(colorToMove))
	ExportEnumConstants(hiveModule)
	js.Global().Set("hive", hiveModule)
	select {}
}
