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

	t tomb.Tomb
}

func NewContext(client *rpc.Client, fromBlockNum, toBlockNum uint32) *Context {
	ctx := &Context{
		client:       client,
		fromBlockNum: fromBlockNum,
		toBlockNum:   toBlockNum,
	}

	ctx.t.Go(ctx.blockFetcher)
	ctx.t.Go(ctx.reducer)

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
		case ctx.blockCh <- block:
		case ctx.t.Dying:
			return nil
		}
	}
}

func (ctx *Context) emit(value interface{}) error {
	select {
	case ctx.emitCh <- value:
	case <-ctx.t.Dying:
		return errors.New("terminating")
	}
}
