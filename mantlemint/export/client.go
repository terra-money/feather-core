package export

import (
	"net/http"

	"github.com/gorilla/mux"
	app "github.com/terra-money/feather-core/app"
)

func RegisterRESTRoutes(router *mux.Router, a *app.App) {
	router.Handle("/export/accounts", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		err := ExportAllAccounts(a)
		if err != nil {
			writer.WriteHeader(http.StatusConflict)
			writer.Write([]byte(err.Error()))
		}
		writer.WriteHeader(http.StatusOK)
	})).Methods("POST")

	router.Handle("/export/circulating_supply", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cs, err := ExportCirculatingSupply(a)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cs.String()))
	}))
}
