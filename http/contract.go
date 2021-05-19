package http

import (
	"encoding/json"
	"net/http"
	"time"

	"stageai.tech/sunshine/sunshine/contract"
	"stageai.tech/sunshine/sunshine/controller"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type contractHandler struct {
	c *controller.Contract
	s sessions.Store
}

func newContractHandler(env *services.Env) *contractHandler {
	return &contractHandler{
		c: controller.NewContract(env),
		s: env.SessionStore,
	}
}

func (ch *contractHandler) getTable(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	table, err := ch.c.GetTable(r.Context(), id, mux.Vars(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(table)
}

func (ch *contractHandler) downloadEnglishPDF(w http.ResponseWriter, r *http.Request) {
	ch.download(w, r, "english", "pdf")
}
func (ch *contractHandler) downloadNativePDF(w http.ResponseWriter, r *http.Request) {
	ch.download(w, r, "native", "pdf")
}
func (ch *contractHandler) downloadEnglishTeX(w http.ResponseWriter, r *http.Request) {
	ch.download(w, r, "english", "tex")
}
func (ch *contractHandler) downloadNativeTeX(w http.ResponseWriter, r *http.Request) {
	ch.download(w, r, "native", "tex")
}

func (ch *contractHandler) download(w http.ResponseWriter, r *http.Request, language, format string) {
	id := mustExtractUUID(r)

	file, name, err := ch.c.DownloadContract(r.Context(), id, language, format)
	if err != nil {
		writeError(w, r, err)
		return
	}
	defer file.Close()

	w.Header().Del("Content-Type")
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	http.ServeContent(w, r, name, time.Now(), file)
}

func (ch *contractHandler) downloadNativeAgreementPDF(w http.ResponseWriter, r *http.Request) {
	ch.downloadAgreement(w, r, "native", "pdf")
}

func (ch *contractHandler) downloadEnglishAgreementPDF(w http.ResponseWriter, r *http.Request) {
	ch.downloadAgreement(w, r, "english", "pdf")
}

func (ch *contractHandler) downloadNativeAgreementTex(w http.ResponseWriter, r *http.Request) {
	ch.downloadAgreement(w, r, "native", "tex")
}

func (ch *contractHandler) downloadEnglishAgreementTex(w http.ResponseWriter, r *http.Request) {
	ch.downloadAgreement(w, r, "english", "tex")
}

func (ch *contractHandler) downloadAgreement(w http.ResponseWriter, r *http.Request, language, format string) {
	id := mustExtractUUID(r)

	file, name, err := ch.c.DownloadAgreement(r.Context(), id, format, language)
	if err != nil {
		writeError(w, r, err)
		return
	}
	defer file.Close()

	w.Header().Del("Content-Type")
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	http.ServeContent(w, r, name, time.Now(), file)
}

func (ch *contractHandler) updateTable(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)

	var t contract.Table
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, r, err)
		return
	}

	c, err := ch.c.UpdateTable(r.Context(), id, t, mux.Vars(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(*c)
}

func (ch *contractHandler) updateFields(w http.ResponseWriter, r *http.Request) {
	var fields contract.JSONMap
	id := mustExtractUUID(r)

	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		writeError(w, r, err)
		return
	}

	doc, err := ch.c.UpdateFields(r.Context(), id, fields)
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(doc)
}

func (ch *contractHandler) getFields(w http.ResponseWriter, r *http.Request) {
	id := mustExtractUUID(r)
	f, err := ch.c.GetFields(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(f)
}

func (ch *contractHandler) getAgreement(w http.ResponseWriter, r *http.Request) {
	a, err := ch.c.GetAgreement(r.Context(), mustExtractUUID(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(a)
}

func (ch *contractHandler) updateAgreement(w http.ResponseWriter, r *http.Request) {
	var ctr contract.JSONMap
	if err := json.NewDecoder(r.Body).Decode(&ctr); err != nil {
		writeError(w, r, err)
		return
	}
	doc, err := ch.c.UpdateAgreement(r.Context(), mustExtractUUID(r), ctr)
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(doc)
}

func (ch *contractHandler) getMaintenance(w http.ResponseWriter, r *http.Request) {
	m, err := ch.c.GetMaintenance(r.Context(), mustExtractUUID(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(m)
}

func (ch *contractHandler) updateMaintenance(w http.ResponseWriter, r *http.Request) {
	var ctr contract.JSONMap
	if err := json.NewDecoder(r.Body).Decode(&ctr); err != nil {
		writeError(w, r, err)
		return
	}

	doc, err := ch.c.UpdateMaintenance(r.Context(), mustExtractUUID(r), ctr)
	if err != nil {
		writeError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(doc)
}

func (ch *contractHandler) updateMarkdown(w http.ResponseWriter, r *http.Request) {
	m, err := ch.c.UpdateMarkdown(r.Context(), mustExtractUUID(r), r.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Write(m)
}

func (ch *contractHandler) getMarkdown(w http.ResponseWriter, r *http.Request) {
	m, err := ch.c.GetMarkdown(r.Context(), mustExtractUUID(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Write(m)
}
