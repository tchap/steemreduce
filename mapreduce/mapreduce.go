package mapreduce

import (
	"fmt"
	"io"
	"strconv"

	"github.com/go-steem/rpc"
)

const Author = "void"

type Value struct {
	URL           string
	PendingPayout float64
}

type Accumulator struct {
	PendingPayout float64
}

func NewAccumulator(client *rpc.Client) (*Accumulator, error) {
	return &Accumulator{}, nil
}

func Map(client *rpc.Client, block *rpc.Block, emit func(interface{}) error) error {
	for _, tx := range block.Transactions {
		for _, op := range tx.Operations {
			switch body := op.Body.(type) {
			case *rpc.CommentOperation:
				if body.Author == Author && body.IsNewStory() {
					content, err := client.GetContent(body.Author, body.Permlink)
					if err != nil {
						return err
					}

					// Drop trailing " STEEM".
					payoutString := content.PendingPayoutValue[:len(content.PendingPayoutValue)-6]

					// Convert to float64.
					payout, err := strconv.ParseFloat(payoutString, 64)
					if err != nil {
						return err
					}

					v := &Value{content.URL, payout}
					fmt.Println(v)
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
	next := _next.(*Value)

	acc.PendingPayout += next.PendingPayout
	return nil
}

func WriteResults(_acc interface{}, writer io.Writer) error {
	acc := _acc.(*Accumulator)
	_, err := fmt.Fprintln(writer, acc.PendingPayout)
	return err
}