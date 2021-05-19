package http

import (
	"encoding/json"
	"net/http"

	"stageai.tech/sunshine/sunshine/controller"
)

func (ch *contractHandler) getIndoorClima(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)

	doc, err := ch.c.GetIndoorClima(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(doc)
}

func (ch *contractHandler) updateIndoorClima(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	decode := controller.MarshalJSON(r.Body)

	doc, err := ch.c.UpdateIndoorClima(r.Context(), id, decode)
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(doc)
}
