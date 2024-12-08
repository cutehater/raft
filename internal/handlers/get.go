package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"hw2/internal/models"
	"hw2/internal/raft"
)

func MakeMasterReadHandler(v *raft.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		replica := v.HTTPNodesAddress[rand.Int()%len(v.HTTPNodesAddress)]

		v.Mu.RLock()
		newRequestBody := fmt.Sprintf(`{"lastCommittedIdx": %d}`, v.LastCommittedIdx)
		v.Mu.RUnlock()

		r.Body = io.NopCloser(strings.NewReader(newRequestBody))
		r.ContentLength = int64(len(newRequestBody))
		r.Header.Set("Content-Type", "application/json")

		http.Redirect(w, r, replica+"/read"+r.URL.Path, http.StatusFound)
	}
}

func MakeReplicaReadHandler(v *raft.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ID, err := getID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		type reqBody struct {
			LastCommittedIdx int64 `json:"lastCommittedIdx"`
		}
		var req reqBody
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Failed to parse helper data", http.StatusInternalServerError)
			return
		}

		v.Mu.Lock()
		defer v.Mu.Unlock()
		if req.LastCommittedIdx > v.LastCommittedIdx {
			http.Error(w, "Stale replica", http.StatusInternalServerError)
			return
		}
		resource, ok := v.Data[models.Id(ID)]

		if !ok {
			http.Error(w, errNotFound.Error(), http.StatusNotFound)
			return
		}

		models.JSONResponse(w, http.StatusOK, resource)
	}
}
