//go:build js && wasm

package main

import (
	"HiveServer/src/hivegame"
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
			"pieceType": tile.PieceType,
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

func JsValueToHiveGame(value js.Value) (hivegame.HiveGame, bool) {
	if value.Type() != js.TypeObject {
		return hivegame.HiveGame{}, false
	}

	colorToMove, ok := HiveColorFromJsValue(value.Get("colorToMove"))

	if !ok {
		return hivegame.HiveGame{}, false
	}

	move, ok := JsValueToInt(value.Get("move"))

	if !ok {
		return hivegame.HiveGame{}, false
	}

	tiles := value.Get("tiles")

	if tiles.Type() != js.TypeObject || !tiles.InstanceOf(js.Global().Get("Array")) {
		return hivegame.HiveGame{}, false
	}

	tilesLength := tiles.Get("length").Int()

	parsedTiles := make([]hivegame.HiveTile, tilesLength)

	for i := range tilesLength {
		//ok := parsedTiles[i].FromJsValue(tiles.Index(i))
		if tile, ok := JsValueToHiveTile(tiles.Index(i)); !ok {
			return hivegame.HiveGame{}, false
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
			panic("cannot lookup string for piece value")
		}

		blackReserve[pieceType] = blackReserveJsValue.Get(pieceString).Int()
	}

	whiteReserve := make(map[hivegame.HivePieceType]int)
	whiteReserveJsValue := value.Get("whiteReserve")
	for _, pieceType := range pieceTypes {
		pieceString, ok := pieceTypeStrings[pieceType]

		if !ok {
			panic("cannot lookup string for piece value")
		}

		whiteReserve[pieceType] = whiteReserveJsValue.Get(pieceString).Int()
	}

	game := hivegame.HiveGame{}

	game.ColorToMove = colorToMove
	game.Move = move
	game.Tiles = parsedTiles
	game.BlackReserve = blackReserve
	game.WhiteReserve = whiteReserve

	return game, true
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

func createHiveGame(this js.Value, args []js.Value) interface{} {
	game := hivegame.CreateHiveGame()
	return HiveGameToJsValue(game)
}

func placeTile(this js.Value, args []js.Value) interface{} {
	if len(args) != 3 {
		panic("placeTile function expects 3 arguments : game, piece type, position")
	}

	game, ok := JsValueToHiveGame(args[0])

	if !ok {
		panic("could not parse js value to hive game")
	}

	pieceType, ok := JsValueToHivePieceType(args[1])

	if !ok {
		panic("could not parse js value to hive piece type")
	}

	position, ok := JsValueToHexVectorInt(args[2])

	if !ok {
		panic("could not parse js value to hex vector position")
	}

	game.PlaceTile(position, pieceType)

	return HiveGameToJsValue(game)
}

func main() {
	hiveModule := js.Global().Get("Object").New()
	hiveModule.Set("createHiveGame", js.FuncOf(createHiveGame))
	hiveModule.Set("placeTile", js.FuncOf(placeTile))
	ExportEnumConstants(hiveModule)
	js.Global().Set("hive", hiveModule)
	select {}
}
