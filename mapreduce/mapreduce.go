package mapreduce

import (
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/go-steem/rpc"
)

const Author = "void"

type Story struct {
	BlockNum      uint32
	Title         string
	Permlink      string
	PendingPayout float64
}

type Accumulator struct {
	Stories            []*Story
	ProcessedStories   map[string]struct{}
	TotalPendingPayout float64
}

// NewAccumulator returns an empty accumulator.
// No stories in the map, pending payout total set to 0.
func NewAccumulator(client *rpc.Client) (*Accumulator, error) {
	return &Accumulator{
		Stories:          make([]*Story, 0, 100),
		ProcessedStories: make(map[string]struct{}, 100),
	}, nil
}

// Map in this case emits a value for every story operation by the given author.
func Map(client *rpc.Client, emit func(interface{}) error, block *rpc.Block) error {
	for _, tx := range block.Transactions {
		for _, op := range tx.Operations {
			switch body := op.Body.(type) {
			case *rpc.CommentOperation:
				// Not interested in other authors.
				if body.Author != Author {
					continue
				}

				// Not interested in comments.
				if !body.IsStoryOperation() {
					continue
				}

				// Assemble the value and emit it.
				value := &Story{
					BlockNum: block.Number,
					Title:    body.Title,
					Permlink: body.Permlink,
				}
				if err := emit(value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Reduce stores the story in the map in case it is a new story operation
// and adds the story pending payout to the sum of all pending payouts.
func Reduce(client *rpc.Client, _acc, _next interface{}) error {
	// We need to do type assertions here.
	acc := _acc.(*Accumulator)
	story := _next.(*Story)

	// In case we have already seen the story, we are done here.
	if _, ok := acc.ProcessedStories[story.Title]; ok {
		return nil
	}

	// Get current pending payout.
	content, err := client.GetContent(Author, story.Permlink)
	if err != nil {
		return err
	}

	// Convert to float64.
	payout, err := steemToFloat64(content.PendingPayoutValue)
	if err != nil {
		return err
	}

	// Store the payout value in the story object.
	story.PendingPayout = payout

	// Store the story in the map and mark it as processed.
	acc.Stories = append(acc.Stories, story)
	acc.ProcessedStories[story.Permlink] = struct{}{}

	// Add the pending payout.
	acc.TotalPendingPayout += story.PendingPayout

	// Done.
	return nil
}

// WriteResults uses a text/tabwriter to format the output.
func WriteResults(_acc interface{}, writer io.Writer) error {
	// We need to do type assertions here.
	acc := _acc.(*Accumulator)

	// Format and write.
	tw := tabwriter.NewWriter(writer, 0, 8, 0, '\t', 0)
	fmt.Fprintln(tw)
	fmt.Fprint(tw, "Block\tTitle\tPending Payout\n")
	fmt.Fprint(tw, "=====\t=====\t==============\n")
	for _, story := range acc.Stories {
		fmt.Fprintf(tw, "%v\t%v\t%v\n", story.BlockNum, story.Title, story.PendingPayout)
	}
	fmt.Fprintf(tw, "\nTotal pending payout: %v\n\n", acc.TotalPendingPayout)

	// Flush the buffer.
	return tw.Flush()
}

func steemToFloat64(value string) (float64, error) {
	return strconv.ParseFloat(value[:len(value)-6], 64)
}
