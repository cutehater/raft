package raft

import (
	"sync"

	"hw2/internal/models"
	"hw2/internal/rpc"
)

type NodeRole string

const (
	RoleFollower  NodeRole = "Follower"
	RoleCandidate NodeRole = "Candidate"
	RoleLeader    NodeRole = "Leader"
)

type Node struct {
	rpc.UnimplementedRaftNodeServer

	Mu               sync.RWMutex
	Role             NodeRole
	Id               int64
	VotedFor         int64
	Term             int64
	HTTPNodesAddress []string
	LeaderId         int64
	Data             map[models.Id]models.Value
	DataLog          []models.LogEntry
	LastCommittedIdx int64
	GrpcClients      []rpc.RaftNodeClient

	appendEntriesMu []sync.Mutex
}
