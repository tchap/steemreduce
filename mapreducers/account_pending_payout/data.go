package accountpendingpayout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
)

const (
	StateFilename  = "mapreduce.json"
	OutputFilename = "output.txt"
)

type Config struct {
	Author string `json:"author"`
}

type Data struct {
	Config *Config          `json:"config,omitempty"`
	State  *State           `json:"state,omitempty"`
	Acc    *AccumulatorData `json:"accumulator,omitempty"`
}

func (data *Data) WriteOutput(writer io.Writer) error {
	acc := data.Acc.Accumulator

	// Format and write.
	tw := tabwriter.NewWriter(writer, 0, 1, 4, ' ', 0)
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

type State struct {
	BlockRangeFrom     uint32 `json:"block_range_from,omitempty"`
	BlockRangeTo       uint32 `json:"block_range_to,omitempty"`
	NextBlockToProcess uint32 `json:"next_block,omitempty"`
}

type AccumulatorData struct {
	*Accumulator
}

func (accData *AccumulatorData) UnmarshalJSON(data []byte) error {
	var acc Accumulator
	if err := json.Unmarshal(data, &acc); err != nil {
		return err
	}

	acc.ProcessedStories = make(map[string]*Story, len(acc.Stories))
	for _, story := range acc.Stories {
		acc.ProcessedStories[story.Permlink] = story
	}

	accData.Accumulator = &acc
	return nil
}

func loadData(dataDirectoryPath string) (*Data, error) {
	// Open the state file.
	stateFilePath := filepath.Join(dataDirectoryPath, StateFilename)
	fd, err := os.Open(stateFilePath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	// Unmarshall the state data.
	var data Data
	if err := json.NewDecoder(fd).Decode(&data); err != nil {
		return nil, err
	}

	// Make sure the data object is filled with non-nil values.
	if data.Config == nil || data.Config.Author == "" {
		return nil, fmt.Errorf("%v: key not set: config.author", stateFilePath)
	}
	if data.State == nil {
		data.State = &State{}
	}
	if data.Acc == nil {
		data.Acc = &AccumulatorData{
			Accumulator: &Accumulator{
				Stories:          make([]*Story, 0, 100),
				ProcessedStories: make(map[string]*Story, 100),
			},
		}
	}

	// Return the data object.
	return &data, nil
}

func storeData(dataDirectoryPath string, data *Data) error {
	// Make sure the directory exists.
	if err := os.MkdirAll(dataDirectoryPath, 0750); err != nil {
		return err
	}

	// Store the state.
	statePath := filepath.Join(dataDirectoryPath, StateFilename)
	stateFile, err := os.OpenFile(statePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer stateFile.Close()

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if _, err := io.Copy(stateFile, bytes.NewReader(content)); err != nil {
		return err
	}

	// Store the human-readable output.
	outputPath := filepath.Join(dataDirectoryPath, OutputFilename)
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return data.WriteOutput(outputFile)
}
