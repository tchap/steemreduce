package runner

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/go-steem/rpc"
	"gopkg.in/tomb.v2"
)

type BlockMapReducer interface {
	Initialise(client *rpc.Client) (acc interface{}, err error)
	BlockRange() (blockRangeFrom, blockRangeTo uint32)
	Map(client *rpc.Client, emit func(interface{}) error, block *rpc.Block) (err error)
	Reduce(client *rpc.Client, acc, value interface{}) (newAcc interface{}, err error)
	ProcessResults(acc interface{}, nextBlockToProcess uint32) (err error)
}

type Context struct {
	client *rpc.Client

	implementation BlockMapReducer
	acc            interface{}

	blockRangeFrom uint32
	blockRangeTo   uint32

	mapCh              chan *rpc.Block
	reduceCh           chan interface{}
	unprocessedBlockCh chan uint32

	wg sync.WaitGroup
	t  tomb.Tomb
}

func Run(client *rpc.Client, implementation BlockMapReducer) (*Context, error) {
	// Compute how many mappers to start.
	numMappers := runtime.NumCPU() - 1
	if numMappers == 0 {
		numMappers = 1
	}

	// Prepare a new Context object.
	ctx := &Context{
		client:             client,
		implementation:     implementation,
		mapCh:              make(chan *rpc.Block, numMappers*10),
		reduceCh:           make(chan interface{}, 0),
		unprocessedBlockCh: make(chan uint32, 1),
	}

	// Initialise MapReduce.
	fmt.Println("---> Runner: Initialising MapReduce ...")
	acc, err := implementation.Initialise(client)
	if err != nil {
		fmt.Fprintln(os.Stderr, "---> Runner: Failed to initialise MapReduce:", err)
		return nil, err
	}
	ctx.acc = acc

	// Get the block range to process.
	fmt.Println("---> Runner: Getting the block range to process ...")
	from, to := implementation.BlockRange()
	if to == 0 {
		props, err := client.GetDynamicGlobalProperties()
		if err != nil {
			return nil, err
		}
		to = props.LastIrreversibleBlockNum
	}
	ctx.blockRangeFrom = from
	ctx.blockRangeTo = to

	// Start the fetcher and the reducer.
	ctx.t.Go(ctx.blockFetcher)
	ctx.t.Go(ctx.reducer)

	// Close the reduce channel once all mappers are done.
	fmt.Printf("---> Mapper: Spawning %v threads ...\n", numMappers)
	ctx.wg.Add(numMappers)
	go func() {
		ctx.wg.Wait()
		close(ctx.reduceCh)
	}()

	// Start the mappers.
	for i := 0; i < numMappers; i++ {
		ctx.t.Go(ctx.mapper)
	}

	return ctx, nil
}

func (ctx *Context) Interrupt() {
	ctx.t.Kill(nil)
}

func (ctx *Context) Wait() error {
	return ctx.t.Wait()
}

func (ctx *Context) blockFetcher() error {
	// Shortcuts.
	client := ctx.client
	from, to := ctx.blockRangeFrom, ctx.blockRangeTo

	defer client.Close()

	// Make sure we are not doing bullshit.
	if from > to {
		return fmt.Errorf("invalid block range: [%v, %v]", from, to)
	}

	// Progress bar!
	numBlocks := to - from + 1
	bar := pb.New(int(numBlocks))
	bar.Width = 80
	bar.ShowTimeLeft = true
	bar.ShowFinalTime = true
	bar.RefreshRate = 5 * time.Second

	// Fetch all blocks matching the given range.
	next := from
	defer func() {
		ctx.unprocessedBlockCh <- next
		close(ctx.unprocessedBlockCh)
	}()

	fmt.Printf("---> Fetcher: Fetching blocks in range [%v, %v]\n", from, to)
	bar.Start()
	for ; next <= to; next++ {
		block, err := client.GetBlock(next)
		if err != nil {
			return err
		}

		bar.Increment()

		select {
		case ctx.mapCh <- block:
		case <-ctx.t.Dying():
			return nil
		}
	}

	// Signal that all blocks have been enqueued.
	bar.FinishPrint("---> Fetcher: All blocks fetched and enqueued")
	close(ctx.mapCh)
	return nil
}

func (ctx *Context) mapper() error {
	defer ctx.wg.Done()

	for {
		select {
		case block, ok := <-ctx.mapCh:
			if !ok {
				return nil
			}
			if err := ctx.implementation.Map(ctx.client, ctx.emit, block); err != nil {
				if err == tomb.ErrDying {
					return nil
				}
				return err
			}
		case <-ctx.t.Dying():
			return nil
		}
	}
}

func (ctx *Context) emit(v interface{}) error {
	select {
	case ctx.reduceCh <- v:
		return nil
	case <-ctx.t.Dying():
		return tomb.ErrDying
	}
}

func (ctx *Context) reducer() (err error) {
	// Get the initial accumulator value.
	acc := ctx.acc

	// Process the results on exit.
	defer func() {
		fmt.Println("---> Reducer: Processing the results ...")
		ex := ctx.implementation.ProcessResults(acc, <-ctx.unprocessedBlockCh)
		if ex != nil {
			if err == nil {
				err = ex
			} else {
				fmt.Fprintln(os.Stderr, "---> Reducer: Failed to process the results:", ex)
			}
		}
	}()

	fmt.Println("---> Reducer: Starting to process values being emitted ...")
	for {
		select {
		case next, ok := <-ctx.reduceCh:
			if !ok {
				fmt.Println("---> Reducer: We are done, writing the output ...")
				return nil
			}
			var ex error
			acc, ex = ctx.implementation.Reduce(ctx.client, acc, next)
			if ex != nil {
				return ex
			}
		case <-ctx.t.Dying():
			return nil
		}
	}
}
