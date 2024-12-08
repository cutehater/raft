package handlers

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"

	"hw2/internal/raft"
)

func getID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return 0, errInvalidID
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, errInvalidID
	}

	return idInt, nil
}

func replicateToReplicas(v *raft.Node, logEntryIdx int64) {
	agreeCount := 1
	cond := sync.NewCond(&sync.Mutex{})

	cond.L.Lock()
	defer cond.L.Unlock()

	for i := range v.HTTPNodesAddress {
		if int64(i) == v.Id {
			continue
		}

		go func() {
			v.UpdateFollower(i, logEntryIdx)
			cond.L.Lock()
			agreeCount++
			if agreeCount > len(v.HTTPNodesAddress)/2 {
				cond.Signal()
			}
			cond.L.Unlock()
		}()
	}

	cond.Wait()
}
