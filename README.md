# steemreduce

Your personal MapReduce for Steem - https://steem.io

## About

`steemreduce` does just one thing - it runs MapReduce over the given range of blocks,
thus allowing you to collect data and generate your own statistics. Cool, right?

## Example

```
$ ./steemreduce -rpc_endpoint="ws://$(docker-machine ip default):8090"
---> Mapper: Spawning 3 threads ...
---> Fetcher: Fetching blocks in range [0, 1909012]
---> Reducer: Getting the initial value ...
---> Reducer: Starting to process incoming blocks ...
 1909013 / 1909012 [============================================] 100.00% 51m21s
---> Fetcher: All blocks fetched and enqueued
---> Reducer: We are done, writing the output ...
$ cat output.txt 

Block   Title                                               Pending Payout
=====   =====                                               ==============
1268905 The Body Knows                                      2.6
1325197 Let it Go, Let it Happen                            444.2
1473920 So Much Energy is Wasted                            0
1548838 Accessing steemd RPC endpoint using Golang          0.2
1592040 The Evolution of Consciousness                      21.7
1605795 Nobody Has to Be Anybody!                           0
1728062 We Are Taught, Yet We Do Not Understand             21.5
1786011 Dockerfile for steemd, tuned for performance        35.7
1851608 Evaluating Martial Arts                             48.3
1872747 go-steem/rpc: Golang RPC Client Library for Steem   711.8
1907611 steemreduce: You personal MapReduce for Steem       836.6

Total pending payout: 2122.6
```

## How to Implement Your MapReduce

Before even starting, you need to install [Go](https://golang.org/dl/).

Then you need to set up a [Go workspace](https://golang.org/doc/code.html#Workspaces)
and clone this package into the right directory:

```
$ git clone https://github.com/tchap/steemreduce "$WORKSPACE/src/github.com/tchap/steemreduce"
```

And now, you can finally start implementing your MapReduce.

```
cd "$WORKSPACE/src/github.com/tchap/steemreduce"
```

Now, go to check the `mapreduce` directory. When you open `mapreduce.go`, you will
see a few exported functions there. These functions form the implementation of
MapReduce. The `mapreduce` package as implemented here is just an example that
collects and prints the pending payout for all stories published by the given
user. In case you want to do something else, you just need to change the
implementation to do what you want:

```
cp -R mapreduce mapreduce_old
```

Now you are ready to start hacking. You can consult `mapreduce_old` any time
something is not entirely clear. When you are done, just run

```
$ go build
```

This will place `steemreduce` into the current working directory.
All that is left is to run the executable with the right flags. Please see

```
$ ./steemreduce -h
```

for that.

## Roadmap

1. Make it possible to start where finished last time.

## License

`MIT`, see the `LICENSE` file.
