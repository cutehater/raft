package handlers

import (
	"encoding/json"
	"net/http"

	"hw2/internal/models"
	"hw2/internal/raft"
)

func MakePostHandler(v *raft.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errInvalidMethod.Error(), http.StatusMethodNotAllowed)
			return
		}

		var resource models.Resource
		if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
			http.Error(w, errInvalidData.Error(), http.StatusBadRequest)
			return
		}

		v.Mu.Lock()
		if v.Role != raft.RoleLeader {
			http.Error(w, "Invalid request: node is not master", http.StatusBadRequest)
			return
		}
		if _, ok := v.Data[resource.Id]; ok {
			v.Mu.Unlock()
			http.Error(w, errResourceExists.Error(), http.StatusBadRequest)
			return
		}
		logEntryIdx := len(v.DataLog)
		v.DataLog = append(v.DataLog, models.LogEntry{
			Id:    resource.Id,
			Value: resource.Value,
			Term:  v.Term,
		})
		v.Mu.Unlock()

		replicateToReplicas(v, int64(logEntryIdx))
		v.Mu.Lock()
		v.Data[resource.Id] = resource.Value
		v.LastCommittedIdx++
		v.Mu.Unlock()

		w.WriteHeader(http.StatusCreated)
	}
}
