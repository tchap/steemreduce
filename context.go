package main

import (
	"errors"
	"fmt"

	"github.com/go-steem/rpc"
	"gopkg.in/tomb.v2"
)

type Context struct {
	client *rpc.Client

	fromBlockNum uint32
	toBlockNum   uint32

	mappersGroup sync.WaitGroup
	t            tomb.Tomb
}

func NewContext(client *rpc.Client, fromBlockNum, toBlockNum uint32) *Context {
	ctx := &Context{
		client:       client,
		fromBlockNum: fromBlockNum,
		toBlockNum:   toBlockNum,
	}

	// Start the fetcher and the reducer.
	ctx.t.Go(ctx.blockFetcher)
	ctx.t.Go(ctx.reducer)

	// Compute how many mappers to start.
	numMappers := runtime.NumCPU() - 1
	if numMappers == 0 {
		numMappers = 1
	}

	// Close the reduce channel once all mappers are done.
	ctx.mappersGroup.Add(numMappers)
	go func() {
		ctx.mappersGroup.Wait()
		close(ctx.reduceCh)
	}()

	// Start the mappers.
	for i := 0; i < numMappers; i++ {
		ctx.t.Go(ctx.mapper)
	}

	return ctx
}

func (ctx *Context) Interrupt() {
	ctx.t.Kill(nil)
}

func (ctx *Context) Wait() error {
	ctx.t.Wait()
}

func (ctx *Context) blockFetcher() error {
	// Shortcuts.
	var (
		client = ctx.client
		from   = ctx.rangeFrom
		to     = ctx.rangeTo
	)

	// Make sure we are not doing bullshit.
	if from > to {
		return fmt.Errorf("invalid block range: [%v, %v]", from, to)
	}

	// Fetch all blocks matching the given range.
	fmt.Printf("---> BlockFetcher: Fetching blocks in range [%v, %v]\n", from, to)
	for next := from; next <= to; next++ {
		block, err := client.GetBlock(next)
		if err != nil {
			return err
		}

		select {
		case ctx.mapCh <- block:
		case ctx.t.Dying():
			return nil
		}
	}

	// Signal that all blocks have been enqueued.
	close(ctx.mapCh)
	return nil
}

func (ctx *Context) mapper() error {
	defer ctx.mappers.Done()

	for {
		select {
		case block := <-ctx.mapCh:
			if err := mapreduce.Map(ctx.client, block, ctx.emit); err != nil {
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
	case <-ctx.t.Dying():
		return tomp.ErrDying
	}
}

func (ctx *Context) reducer() error {
	fmt.Println("---> Reducer: Getting the initial value ...")
	acc, err := mapreduce.NewAccumulator(ctx.client)
	if err != nil {
		return err
	}

	fmt.Println("---> Reducer: Starting to process incoming blocks ...")
	for {
		select {
		case next, ok := <-ctx.reduceCh:
			if !ok {
				return ctx.dump(acc)
			}

			if err := mapreduce.Reduce(acc, next); err != nil {
				return err
			}
		case ctx.t.Dying():
			return nil
		}
	}
}

func (ctx *Context) dump(value interface{}) error {
	dst, err := os.OpenFile("output.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	return mapreduce.WriteResults(value, dst)
}
