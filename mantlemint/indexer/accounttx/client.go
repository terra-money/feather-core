package accounttx

import (
	tmdb "github.com/cometbft/cometbft-db"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/gorilla/mux"
	"github.com/terra-money/feather-core/mantlemint/indexer"
	"net/http"
	"strconv"
)

var (
	defaultLimit          = uint64(100)
	defaultOffset         = uint64(0)
	EndpointGETAccountTxs = "/index/account/{account}"
)

var RegisterRESTRoute = indexer.CreateRESTRoute(func(router *mux.Router, indexerDB tmdb.DB) {
	router.HandleFunc(EndpointGETAccountTxs, func(w http.ResponseWriter, r *http.Request) {
		account, ok := mux.Vars(r)["account"]
		if !ok {
			http.Error(w, "invalid request: account not found", 400)
			return
		}
		queries := r.URL.Query()
		offset, err := strconv.ParseUint(queries.Get("offset"), 10, 64)
		if err != nil {
			offset = defaultOffset
		}
		limit, err := strconv.ParseUint(queries.Get("offset"), 10, 64)
		if err != nil {
			limit = defaultLimit
		}
		txs, err := getTxnsByAccount(indexerDB, account, offset, limit)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		rxRes := &GetAccountTxsResponse{
			Limit:  limit,
			Offset: offset,
			Txs:    txs,
		}
		res, err := tmjson.Marshal(rxRes)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.WriteHeader(200)
		w.Write(res)
	})
})
