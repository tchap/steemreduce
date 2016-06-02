package notifications

import (
	"fmt"

	"github.com/go-steem/rpc"
)

type WatchStoryVotesConfig struct {
	Authors []string `yaml:"authors"`
	Voters  []string `yaml:"voters"`
}

type StoryVoteEvent struct {
	Op      *rpc.VoteOperation
	Content *rpc.Content
}

type StoryVotesEventMiner struct {
	authors StringSet
	voters  StringSet
}

func newStoryVotesEventMiner(config *WatchStoryVotesConfig) *StoryVotesEventMiner {
	return &StoryVotesEventMiner{
		authors: MakeStringSet(config.Authors),
		voters:  MakeStringSet(config.Voters),
	}
}

func (miner *StoryVotesEventMiner) MineEvent(operation *rpc.Operation, content *rpc.Content) interface{} {
	if !content.IsStory() {
		return nil
	}

	op, ok := operation.Body.(*rpc.VoteOperation)
	if !ok {
		return nil
	}

	var match bool
	match = match || miner.authors.Contains(op.Author)
	match = match || miner.voters.Contains(op.Voter)
	if !match {
		return nil
	}

	fmt.Println(" ---> Emitting StoryVote event ...")
	return &StoryVoteEvent{op, content}
}
