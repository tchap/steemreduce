package notifications

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-steem/rpc"
)

// BlockMapReducer implements runner.BlockMapReducer interface.
type BlockMapReducer struct {
	config *Config

	eventMiners []EventMiner
	notifiers   []Notifier

	blockRangeFrom uint32
}

func NewBlockMapReducer() *BlockMapReducer {
	return &BlockMapReducer{}
}

func (reducer *BlockMapReducer) Initialise(client *rpc.Client) (interface{}, error) {
	// Load config.
	fmt.Println("---> MapReduce: Loading configuration ...")
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	reducer.config = config

	// In case the command is set, parse the template.
	var notifiers []Notifier
	for _, v := range config.EnabledNotifications {
		fmt.Printf("---> MapReduce: Configuring notifier: %v ...\n", v)
		var (
			notifier Notifier
			err      error
		)
		switch v {
		case "command":
			notifier, err = NewCommandNotifier(config.Command)
		case "email":
			notifier, err = NewEmailNotifier(config.Email)
		case "slack":
			notifier, err = NewSlackNotifier(config.Slack)
		}
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, notifier)
	}
	reducer.notifiers = notifiers

	// Get the last block on the blockchain.
	fmt.Println("---> MapReduce: Getting the block number to start with ...")
	props, err := client.GetDynamicGlobalProperties()
	if err != nil {
		return nil, err
	}
	reducer.blockRangeFrom = props.LastIrreversibleBlockNum
	fmt.Println(" ---> Got", reducer.blockRangeFrom)

	// Set up event miners.
	reducer.eventMiners = []EventMiner{
		newStoriesEventMiner(&config.Watch.Stories),
		newStoryVotesEventMiner(&config.Watch.StoryVotes),
		newCommentsEventMiner(&config.Watch.Comments),
		newCommentVotesEventMiner(&config.Watch.CommentVotes),
	}

	// Return a new accumulator.
	fmt.Println("---> MapReduce: Ready to go!")
	return nil, nil
}

func (reducer *BlockMapReducer) BlockRange() (from, to uint32) {
	return reducer.blockRangeFrom, 0
}

// Map in this case emits a value for every story operation by the given author.
func (reducer *BlockMapReducer) Map(client *rpc.Client, emit func(interface{}) error, block *rpc.Block) error {
	for _, tx := range block.Transactions {
	OpLoop:
		for _, op := range tx.Operations {
			fmt.Printf(
				"---> MapReduce: Processing operation: %v (block %v)\n", op.Type, block.Number)

			// Fetch the associated content.
			var (
				content *rpc.Content
				err     error
			)
			switch body := op.Body.(type) {
			case *rpc.CommentOperation:
				content, err = client.GetContent(body.Author, body.Permlink)
			case *rpc.VoteOperation:
				content, err = client.GetContent(body.Author, body.Permlink)
			default:
				fmt.Println(" ---> No action taken")
				continue OpLoop
			}
			if err != nil {
				return err
			}

			// Mine events.
			for _, eventMiner := range reducer.eventMiners {
				if event := eventMiner.MineEvent(op, content); event != nil {
					if err := emit(event); err != nil {
						return err
					}
					// For now we continue the loop.
					// This is so that comment events are not sent when
					// a story event is sent already.
					continue OpLoop
				}
			}
			fmt.Println(" ---> No action taken")
		}
	}

	return nil
}

func (reducer *BlockMapReducer) Reduce(client *rpc.Client, _acc, _next interface{}) (interface{}, error) {
	var wg sync.WaitGroup

	wg.Add(len(reducer.notifiers))
	for _, notifier := range reducer.notifiers {
		go func(notifier Notifier) {
			defer wg.Done()
			if err := notifier.DispatchNotification(_next); err != nil {
				fmt.Fprintln(os.Stderr, "---> MapReduce: Reduce failed:", err)
			}
		}(notifier)
	}

	wg.Wait()
	return nil, nil
}

func (reducer *BlockMapReducer) ProcessResults(_acc interface{}, nextBlockToProcess uint32) error {
	return nil
}
