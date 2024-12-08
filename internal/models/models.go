package models

import (
	"encoding/json"
	"net/http"

	"hw2/internal/rpc"
)

type Id int64
type Value string

type Resource struct {
	Id    Id    `json:"id"`
	Value Value `json:"value"`
}

type LogEntry struct {
	Id    Id    `json:"id"`
	Value Value `json:"value"`
	Term  int64 `json:"term"`
}

func JSONResponse(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
}

func (e *LogEntry) ToProto() *rpc.Entry {
	return &rpc.Entry{
		Id:    int64(e.Id),
		Value: string(e.Value),
		Term:  e.Term,
	}
}

func LogEntryFromProto(e *rpc.Entry) *LogEntry {
	return &LogEntry{
		Id:    Id(e.Id),
		Value: Value(e.Value),
		Term:  e.Term,
	}
}
