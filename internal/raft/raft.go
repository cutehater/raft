package raft

import (
	"context"
	"log"
	"sync"

	"hw2/internal/models"
	"hw2/internal/rpc"
)

func NewNode(id int64, GRPCNodesAddress []string, HTTPNodesAddress []string) *Node {
	node := &Node{
		Mu:               sync.RWMutex{},
		Role:             RoleFollower,
		Id:               id,
		VotedFor:         -1,
		Term:             1,
		HTTPNodesAddress: HTTPNodesAddress,
		LeaderId:         0,
		Data:             make(map[models.Id]models.Value),
		DataLog:          make([]models.LogEntry, 1),

		appendEntriesMu: make([]sync.Mutex, len(HTTPNodesAddress)),
	}

	if id == 0 {
		node.Role = RoleLeader
	}

	wg := sync.WaitGroup{}
	for i, address := range GRPCNodesAddress {
		if int64(i) == id {
			continue
		}

		go func() {
			defer wg.Done()
			wg.Add(1)
			node.createGRPCClient(i, address)
		}()
	}

	wg.Wait()

	go node.StartElectionTimeout()
	return node
}

func (leader *Node) UpdateFollower(followerIdx int, logEntryIdx int64) {
	// for not resending long stale logs history in several UpdateFollower calls
	leader.appendEntriesMu[followerIdx].Lock()
	defer leader.appendEntriesMu[followerIdx].Unlock()

	leader.Mu.RLock()
	if leader.Role != RoleLeader || followerIdx == int(leader.Id) {
		leader.Mu.RUnlock()
		return
	}

	var entries []*rpc.Entry
	if logEntryIdx != -1 {
		entries = append(entries, leader.DataLog[logEntryIdx].ToProto())
	}
	leader.Mu.RUnlock()

	for {
		leader.Mu.RLock()
		req := &rpc.AppendEntriesIn{
			Term:            leader.Term,
			LeaderId:        leader.Id,
			LeaderCommitIdx: leader.LastCommittedIdx,
			PrevLogIdx:      logEntryIdx - 1,
			Entries:         entries,
		}
		if req.PrevLogIdx >= 0 {
			req.PrevLogTerm = leader.DataLog[req.PrevLogIdx].Term
		}
		leader.Mu.RUnlock()

		resp, err := leader.GrpcClients[followerIdx].AppendEntries(context.Background(), req)

		if err != nil || resp == nil {
			log.Printf("AppendEntries RPC call failed: %v", err)
		} else if !resp.Success {
			leader.Mu.Lock()
			if resp.Term > leader.Term {
				log.Printf("Stale leader: %v", leader.Id)
				leader.Term = resp.Term
				leader.Role = RoleFollower
				leader.Mu.Unlock()
				break
			} else if logEntryIdx > 0 {
				logEntryIdx--
				req.Entries = append([]*rpc.Entry{leader.DataLog[logEntryIdx].ToProto()}, req.Entries...)
			}
			leader.Mu.Unlock()
		} else {
			break
		}

		// do not repeat heartbeats
		if logEntryIdx == -1 {
			break
		}
	}
}
