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
	Stories            []*Story
	PendingPayoutTotal float64
}

func NewAccumulator(client *rpc.Client) (*Accumulator, error) {
	return &Accumulator{
		Stories: make([]*Story, 0, 100),
	}, nil
}

func Map(client *rpc.Client, block *rpc.Block, emit func(interface{}) error) error {
	for _, tx := range block.Transactions {
		for _, op := range tx.Operations {
			switch body := op.Body.(type) {
			case *rpc.CommentOperation:
				if body.Author == Author && body.IsStoryOperation() {
					content, err := client.GetContent(body.Author, body.Permlink)
					if err != nil {
						return err
					}

					// We are done in case this is just an edit.
					if !content.IsNewStory() {
						return nil
					}

					// Drop trailing " STEEM".
					payoutString := content.PendingPayoutValue[:len(content.PendingPayoutValue)-6]

					// Convert to float64.
					payout, err := strconv.ParseFloat(payoutString, 64)
					if err != nil {
						return err
					}

					v := &Story{
						BlockNum:      block.Number,
						Title:         content.Title,
						PendingPayout: payout,
					}
					if err := emit(v); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func Reduce(_acc, _next interface{}) error {
	acc := _acc.(*Accumulator)
	story := _next.(*Story)

	acc.Stories = append(acc.Stories, story)
	acc.PendingPayoutTotal += story.PendingPayout
	return nil
}

func WriteResults(_acc interface{}, writer io.Writer) error {
	acc := _acc.(*Accumulator)

	tw := tabwriter.NewWriter(writer, 0, 8, 0, '\t', 0)
	fmt.Fprintln(tw)
	fmt.Fprint(tw, "Block\tTitle\tPending Payout\n")
	fmt.Fprint(tw, "=====\t=====\t==============\n")
	for _, story := range acc.Stories {
		fmt.Fprintf(tw, "%v\t%v\t%v\n", story.BlockNum, story.Title, story.PendingPayout)
	}
	fmt.Fprintf(tw, "\nTotal pending payout: %v\n\n", acc.PendingPayoutTotal)

	return tw.Flush()
}
