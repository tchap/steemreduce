package accountpendingpayout

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cheggaaa/pb"
	"github.com/go-steem/rpc"
)

const ID = "account_pending_payout"

const DataDirectoryEnvironmentKey = "STEEMREDUCE_PARAMS_DATA_DIR"

type Story struct {
	BlockNum      uint32  `json:"block_number"`
	Title         string  `json:"title"`
	Permlink      string  `json:"permlink"`
	PendingPayout float64 `json:"pending_payout"`
}

type Accumulator struct {
	Stories            []*Story          `json:"stories"`
	ProcessedStories   map[string]*Story `json:"-"`
	TotalPendingPayout float64           `json:"total_pending_payout"`
}

// BlockMapReducer implements runner.BlockMapReducer interface.
type BlockMapReducer struct {
	data              *Data
	dataDirectoryPath string
}

func NewBlockMapReducer() *BlockMapReducer {
	return &BlockMapReducer{}
}

func (reducer *BlockMapReducer) Initialise(client *rpc.Client) (interface{}, error) {
	// Get params from the environment.
	dataDirectoryPath := os.Getenv(DataDirectoryEnvironmentKey)
	if dataDirectoryPath == "" {
		return nil, errors.New("environment variable is not set:" + DataDirectoryEnvironmentKey)
	}

	// Load the data.
	data, err := loadData(dataDirectoryPath)
	if err != nil {
		return nil, err
	}
	reducer.data = data
	reducer.dataDirectoryPath = dataDirectoryPath

	// Update existing data.
	if len(data.Acc.Stories) != 0 {
		if err := reducer.updateData(client); err != nil {
			return nil, err
		}
	}

	// Return the accumulator.
	return reducer.data.Acc.Accumulator, nil
}

func (reducer *BlockMapReducer) updateData(client *rpc.Client) error {
	author := reducer.data.Config.Author
	acc := reducer.data.Acc.Accumulator
	acc.TotalPendingPayout = 0

	fmt.Println("---> MapReduce: Updating known stories ...")

	bar := pb.New(len(acc.Stories))
	bar.Width = 80
	bar.ShowTimeLeft = true
	bar.ShowFinalTime = true
	bar.Start()

	for _, story := range acc.Stories {
		content, err := client.GetContent(author, story.Permlink)
		if err != nil {
			return err
		}

		story.Title = content.Title

		payout, err := steemToFloat64(content.PendingPayoutValue)
		if err != nil {
			return err
		}

		story.PendingPayout = payout
		acc.TotalPendingPayout += payout

		bar.Increment()
	}

	bar.FinishPrint("---> MapReduce: All known stories updated")
	return nil
}

func (reducer *BlockMapReducer) BlockRange() (from, to uint32) {
	state := reducer.data.State

	// FROM
	if state.NextBlockToProcess != 0 {
		from = state.NextBlockToProcess
	} else {
		from = state.BlockRangeFrom
	}

	// TO
	to = state.BlockRangeTo
	return
}

// Map in this case emits a value for every story operation by the given author.
func (reducer *BlockMapReducer) Map(client *rpc.Client, emit func(interface{}) error, block *rpc.Block) error {
	for _, tx := range block.Transactions {
		for _, op := range tx.Operations {
			switch body := op.Body.(type) {
			case *rpc.CommentOperation:
				// Not interested in other authors.
				if body.Author != reducer.data.Config.Author {
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
func (reducer *BlockMapReducer) Reduce(client *rpc.Client, _acc, _next interface{}) (interface{}, error) {
	// We need to do type assertions here.
	acc := _acc.(*Accumulator)
	story := _next.(*Story)

	// In case we have already seen the story, we are done here.
	if storedStory, ok := acc.ProcessedStories[story.Permlink]; ok {
		// Update the title, which might have changed.
		storedStory.Title = story.Title
		return acc, nil
	}

	// Get current pending payout.
	content, err := client.GetContent(reducer.data.Config.Author, story.Permlink)
	if err != nil {
		return acc, err
	}

	// Convert to float64.
	payout, err := steemToFloat64(content.PendingPayoutValue)
	if err != nil {
		return acc, err
	}

	// Store the payout value in the story object.
	story.PendingPayout = payout

	// Store the story in the map and mark it as processed.
	acc.Stories = append(acc.Stories, story)
	acc.ProcessedStories[story.Permlink] = story

	// Add the pending payout.
	acc.TotalPendingPayout += story.PendingPayout

	// Done.
	return acc, nil
}

// WriteResults is used to generate output for the resulting accumulator.
// This implementation uses a text/tabwriter to format the output.
func (reducer *BlockMapReducer) ProcessResults(_acc interface{}, nextBlockToProcess uint32) error {
	// We need to do type assertions here.
	acc := _acc.(*Accumulator)
	reducer.data.State.NextBlockToProcess = nextBlockToProcess
	reducer.data.Acc.Accumulator = acc
	return storeData(reducer.dataDirectoryPath, reducer.data)
}

func steemToFloat64(value string) (float64, error) {
	return strconv.ParseFloat(value[:len(value)-6], 64)
}
