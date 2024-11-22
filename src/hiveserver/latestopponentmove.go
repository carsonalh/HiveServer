package main

import "net/http"

type latestOpponentMoveHandler struct{}

func (h *latestOpponentMoveHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
}
