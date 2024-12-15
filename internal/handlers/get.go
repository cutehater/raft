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
		v.Mu.RLock()
		defer v.Mu.RUnlock()
		if v.Role != raft.RoleLeader {
			http.Error(w, "Invalid request: node is not master", http.StatusBadRequest)
			return
		}

		newRequestBody := fmt.Sprintf(`{"lastCommittedIdx": %d}`, v.LastCommittedIdx)
		r.Body = io.NopCloser(strings.NewReader(newRequestBody))
		r.ContentLength = int64(len(newRequestBody))
		r.Header.Set("Content-Type", "application/json")

		replicaIdx := int(v.Id)
		for replicaIdx == int(v.Id) {
			replicaIdx = rand.Int() % len(v.HTTPNodesAddress)
		}
		replica := v.HTTPNodesAddress[replicaIdx]

		http.Redirect(w, r, replica+"/readReplica/{id}", http.StatusFound)
	}
}

func MakeReplicaReadHandler(v *raft.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v.Mu.Lock()
		defer v.Mu.Unlock()
		fmt.Println("Got a read request")

		if v.Role != raft.RoleFollower {
			http.Error(w, "Invalid request: node is not replica", http.StatusBadRequest)
			return
		}

		ID, err := getID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resource, ok := v.Data[models.Id(ID)]

		if !ok {
			http.Error(w, errNotFound.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resource); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
