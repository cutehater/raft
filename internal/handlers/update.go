package handlers

import (
	"encoding/json"
	"net/http"

	"hw2/internal/models"
	"hw2/internal/raft"
)

func MakeUpdateHandler(v *raft.Node) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			http.Error(w, errInvalidMethod.Error(), http.StatusMethodNotAllowed)
			return
		}

		var resource models.Resource
		if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
			http.Error(w, errInvalidData.Error(), http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodDelete && resource.Value != "" || r.Method != http.MethodDelete && resource.Value == "" {
			http.Error(w, errInvalidData.Error(), http.StatusBadRequest)
			return
		}

		v.Mu.Lock()
		if _, ok := v.Data[resource.Id]; !ok {
			v.Mu.Unlock()
			http.Error(w, errNotFound.Error(), http.StatusBadRequest)
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

		models.JSONResponse(w, http.StatusOK, resource)
	}
}
