package models

import (
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
