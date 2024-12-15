package raft

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"hw2/internal/rpc"
)

func (follower *Node) StartElectionTimeout() {
	ticker := time.NewTicker(getElectionTimeout())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			follower.Mu.Lock()
			if follower.Role != RoleLeader && follower.needToStartElection {
				follower.startElection()
			} else {
				follower.needToStartElection = true
				follower.Mu.Unlock()
			}
			ticker.Reset(getElectionTimeout())
		}
	}
}

func (candidate *Node) startElection() {
	log.Printf("%d starting election term %d\n", candidate.Id, candidate.Term+1)

	candidate.Term++
	candidate.Role = RoleCandidate
	candidate.VotedFor = candidate.Id
	candidate.Mu.Unlock()

	votes := 1

	for i, v := range candidate.GrpcClients {
		if int64(i) == candidate.Id {
			continue
		}

		candidate.Mu.RLock()
		req := &rpc.RequestVoteIn{
			Term:        candidate.Term,
			CandidateId: candidate.Id,
			LastLogIdx:  int64(len(candidate.DataLog) - 1),
			LastLogTerm: candidate.DataLog[len(candidate.DataLog)-1].Term,
		}
		candidate.Mu.RUnlock()
		go func() {
			resp, err := v.RequestVote(context.Background(), req)

			if err == nil {
				candidate.Mu.Lock()
				defer candidate.Mu.Unlock()

				if resp.VoteGranted {
					votes++
					fmt.Printf("Vote granted from %d in term %d\n", i, candidate.Term)
					if votes == len(candidate.HTTPNodesAddress)/2+1 {
						log.Printf("%d won Election and becoming leader\n", candidate.Id)
						candidate.Role = RoleLeader
						candidate.LeaderId = candidate.Id
						go candidate.sendHeartbeats()
					}
				} else if resp.Term > candidate.Term {
					log.Printf("%d lost Election and returning to follower\n", candidate.Id)
					candidate.Role = RoleFollower
					candidate.Term = resp.Term
					candidate.VotedFor = -1
				}
			}
		}()
	}
}

func getElectionTimeout() time.Duration {
	maxMs, minMs := 3000, 1000
	randomMs := rand.Intn(maxMs-minMs+1) + minMs
	return time.Duration(randomMs) * time.Millisecond
}
