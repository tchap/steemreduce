package notifications

import (
	"fmt"

	"github.com/go-steem/rpc"
)

type WatchCommentVotesConfig struct {
	Authors []string `yaml:"authors"`
	Voters  []string `yaml:"voters"`
}

type CommentVoteEvent struct {
	Op      *rpc.VoteOperation
	Content *rpc.Content
}

type CommentVotesEventMiner struct {
	authors StringSet
	voters  StringSet
}

func newCommentVotesEventMiner(config *WatchCommentVotesConfig) *CommentVotesEventMiner {
	return &CommentVotesEventMiner{
		authors: MakeStringSet(config.Authors),
		voters:  MakeStringSet(config.Voters),
	}
}

func (miner *CommentVotesEventMiner) MineEvent(operation *rpc.Operation, content *rpc.Content) interface{} {
	// Do nothing in case this is a story event.
	if !content.IsStory() {
		return nil
	}

	// Make sure this is a VoteOperation.
	op, ok := operation.Body.(*rpc.VoteOperation)
	if !ok {
		return nil
	}

	// Match.
	var match bool
	match = match || miner.authors.Contains(op.Author)
	match = match || miner.voters.Contains(op.Voter)
	if !match {
		return nil
	}

	// Create the event.
	fmt.Println(" ---> Emitting CommentVote event ...")
	return &CommentVoteEvent{op, content}
}
