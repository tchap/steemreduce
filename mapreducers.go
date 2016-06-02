package main

import (
	app "github.com/tchap/steemreduce/mapreducers/account_pending_payout"
	notif "github.com/tchap/steemreduce/mapreducers/notifications"
	"github.com/tchap/steemreduce/runner"
)

var (
	availableMapReducerIDs = make([]string, 0)
	availableMapReducers   = make(map[string]runner.BlockMapReducer)
)

func MustRegisterMapReducer(id string, implementation runner.BlockMapReducer) {
	if _, ok := availableMapReducers[id]; ok {
		panic("MapReduce implementation already registered: " + id)
	}

	availableMapReducerIDs = append(availableMapReducerIDs, id)
	availableMapReducers[id] = implementation
}

func init() {
	MustRegisterMapReducer(app.Id, app.NewBlockMapReducer())
	MustRegisterMapReducer(notif.Id, notif.NewBlockMapReducer())
}
