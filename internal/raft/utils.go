package raft

import (
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"hw2/internal/rpc"
)

const (
	heartbeatInterval = 10 * time.Millisecond
)

// call under v.Mu.Lock() only
func (v *Node) CommitChanges(lastCommitIdx int64, lastCommitTerm int64) bool {
	if lastCommitIdx < 0 || int(lastCommitIdx) >= len(v.DataLog) ||
		v.DataLog[lastCommitIdx].Term != lastCommitTerm {
		return false
	}

	for i := v.LastCommittedIdx + 1; i <= lastCommitIdx; i++ {
		if v.DataLog[i].Value == "" {
			delete(v.Data, v.DataLog[i].Id)
		} else {
			v.Data[v.DataLog[i].Id] = v.DataLog[i].Value
		}
	}
	v.LastCommittedIdx = lastCommitIdx

	return true
}

func (leader *Node) sendHeartbeats() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			leader.Mu.Lock()
			if leader.Role != RoleLeader {
				leader.Mu.Unlock()
				return
			}
			for i := range leader.HTTPNodesAddress {
				go leader.UpdateFollower(i, -1)
			}
			leader.Mu.Unlock()
		}
	}
}

func (v *Node) createGRPCClient(id int, address string) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to node %d at %s: %v", id, address, err)
	}

	v.Mu.Lock()
	defer v.Mu.Unlock()
	v.GrpcClients[id] = rpc.NewRaftNodeClient(conn)
}
