package notifications

import (
	"fmt"
	"github.com/go-steem/rpc"
)

type WatchCommentsConfig struct {
	Authors       []string `yaml:"authors"`
	ParentAuthors []string `yaml:"parent_authors"`
}

type CommentEvent struct {
	Op      *rpc.CommentOperation
	Content *rpc.Content
}

type CommentsEventMiner struct {
	authors       StringSet
	parentAuthors StringSet
}

func newCommentsEventMiner(config *WatchCommentsConfig) *CommentsEventMiner {
	return &CommentsEventMiner{
		authors:       MakeStringSet(config.Authors),
		parentAuthors: MakeStringSet(config.ParentAuthors),
	}
}

func (miner *CommentsEventMiner) MineEvent(operation *rpc.Operation, content *rpc.Content) interface{} {
	if content.IsStory() {
		return nil
	}

	op, ok := operation.Body.(*rpc.CommentOperation)
	if !ok {
		return nil
	}

	var match bool
	match = match || miner.authors.Contains(content.Author)
	match = match || miner.parentAuthors.Contains(content.ParentAuthor)
	if !match {
		return nil
	}

	fmt.Println(" ---> Emitting Comment event ...")
	return &CommentEvent{op, content}
}
