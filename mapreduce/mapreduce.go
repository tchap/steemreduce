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
	PendingPayout float64
}

type Accumulator struct {
	Stories            map[string]*Story
	PendingPayoutTotal float64
}

// NewAccumulator returns an empty accumulator.
// No stories in the map, pending payout total set to 0.
func NewAccumulator(client *rpc.Client) (*Accumulator, error) {
	return &Accumulator{
		Stories: make(map[string]*Story, 100),
	}, nil
}

// Map in this case emits a value for every story operation by the given author.
func Map(client *rpc.Client, block *rpc.Block, emit func(interface{}) error) error {
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

				// Get content metadata.
				content, err := client.GetContent(body.Author, body.Permlink)
				if err != nil {
					return err
				}

				// Convert the pending payout string to float64.
				payout, err := steemToFloat64(content.PendingPayoutValue)
				if err != nil {
					return err
				}

				// Assemble the value and emit it.
				value := &Story{
					BlockNum:      block.Number,
					Title:         content.Title,
					PendingPayout: payout,
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
func Reduce(_acc, _next interface{}) error {
	// We need to do type assertions here.
	acc := _acc.(*Accumulator)
	story := _next.(*Story)

	// We have already seen the story.
	if _, ok := acc.Stories[story.Title]; ok {
		return nil
	}

	// Store the story in the map.
	acc.Stories[story.Title] = story

	// Add the pending payout.
	acc.PendingPayoutTotal += story.PendingPayout

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
	fmt.Fprintf(tw, "\nTotal pending payout: %v\n\n", acc.PendingPayoutTotal)

	// Flush the buffer.
	return tw.Flush()
}

func steemToFloat64(value string) (float64, error) {
	return strconv.ParseFloat(value[:len(value)-6], 64)
}
