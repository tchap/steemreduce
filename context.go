package main

import (
	"errors"

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

	ctx.t.Go(ctx.blockReader)
	ctx.t.Go(ctx.reducer)

	return ctx
}

func (ctx *Context) Interrupt() {
	ctx.t.Kill(nil)
}

func (ctx *Context) Wait() error {
	ctx.t.Wait()
}

func (ctx *Context) blockReader() {
	// Get config.
	log.Println("---> GetConfig()")
	config, err := client.GetConfig()
	if err != nil {
		return err
	}

	// Keep processing incoming blocks forever.
	log.Printf("---> Entering the block processing loop (last block = %v)\n", lastBlock)
	for {
		// Get current properties.
		props, err := client.GetDynamicGlobalProperties()
		if err != nil {
			return err
		}

		// Process new blocks.
		for props.LastIrreversibleBlockNum-lastBlock > 0 {
			block, err := client.GetBlock(lastBlock)
			if err != nil {
				return err
			}

			ctx.ProcessBlock()

			lastBlock++
		}

		// Sleep for STEEMIT_BLOCK_INTERVAL seconds before the next iteration.
		time.Sleep(time.Duration(config.SteemitBlockInterval) * time.Second)
	}
}

func (ctx *Context) emit(value interface{}) error {
	select {
	case ctx.emitCh <- value:
	case <-ctx.t.Dying:
		return errors.New("terminating")
	}
}
