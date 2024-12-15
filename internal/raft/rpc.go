package raft

import (
	"context"
	"fmt"

	"hw2/internal/models"
	"hw2/internal/rpc"
)

func (follower *Node) AppendEntries(ctx context.Context, req *rpc.AppendEntriesIn) (*rpc.AppendEntriesOut, error) {
	follower.Mu.Lock()
	defer follower.Mu.Unlock()

	if follower.Term > req.Term {
		return &rpc.AppendEntriesOut{
			Success: false,
			Term:    follower.Term,
		}, nil
	} else if follower.Term < req.Term {
		fmt.Printf("Updated to term %d and leader %d\n", req.Term, req.LeaderId)
		follower.Role = RoleFollower
		follower.Term = req.Term
		follower.LeaderId = req.LeaderId
		follower.VotedFor = -1
	}

	follower.needToStartElection = false

	// process heartbeat
	if req.PrevLogIdx < 0 {
		return &rpc.AppendEntriesOut{
			Success: follower.commitChanges(req.LeaderCommitIdx, req.PrevLogTerm),
			Term:    follower.Term,
		}, nil
	}

	if int(req.PrevLogIdx) >= len(follower.DataLog) || follower.DataLog[req.PrevLogIdx].Term != req.PrevLogTerm {
		return &rpc.AppendEntriesOut{
			Success: false,
			Term:    follower.Term,
		}, nil
	}

	follower.DataLog = follower.DataLog[:req.PrevLogIdx+1]
	entries := make([]models.LogEntry, 0, len(req.Entries))
	for _, entry := range req.Entries {
		entries = append(entries, *models.LogEntryFromProto(entry))
	}
	follower.DataLog = append(follower.DataLog, entries...)

	commitTo := req.LeaderCommitIdx
	commitTo = min(commitTo, int64(len(follower.DataLog))-1)

	return &rpc.AppendEntriesOut{
		Success: follower.commitChanges(commitTo, follower.DataLog[commitTo].Term),
		Term:    follower.Term,
	}, nil
}

func (elector *Node) RequestVote(ctx context.Context, req *rpc.RequestVoteIn) (*rpc.RequestVoteOut, error) {
	elector.Mu.Lock()
	defer elector.Mu.Unlock()

	fmt.Printf("Vote request from %d in term %d\n", req.CandidateId, req.Term)

	voteGranted := false
	if elector.Term <= req.Term && (elector.VotedFor == -1 || elector.VotedFor == req.CandidateId) && (elector.DataLog[len(elector.DataLog)-1].Term > req.LastLogTerm ||
		elector.DataLog[len(elector.DataLog)-1].Term == req.LastLogTerm && len(elector.DataLog)-1 <= int(req.LastLogIdx)) {
		elector.Term = req.Term
		elector.Role = RoleFollower
		elector.VotedFor = req.CandidateId
		elector.needToStartElection = false
		voteGranted = true
	}

	return &rpc.RequestVoteOut{
		Term:        elector.Term,
		VoteGranted: voteGranted,
	}, nil
}
