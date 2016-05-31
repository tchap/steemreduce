# MapReduce: account\_pending\_payout

This MapReduce collects pending payouts for all stories by the given author.

**KoolAid: This MapReduce supports incremental updates.**

## Usage

Fist, set the path to the data directory:

```bash
export STEEMREDUCE_PARAMS_DATA_DIR=./data
```

Now, create the MapReduce JSON file `./data/account_pending_payout/mapreduce.json`:

```json
{
  "config": {
    "author": "void"
  }
}
```

Now you are ready to run MapReduce:

```bash
steemreduce \
	-rpc_endpoint="ws:$(docker-machine ip default):8090"
	-mapreduce_id=account_pending_payout
```

The output will be located in the `./data/account_pending_payout/output.txt`:

```
Block   Title                                               Pending Payout
=====   =====                                               ==============
1268905	The Body Knows                                      2.4
1325197	Let it Go, Let it Happen                            418.7
1473920	So Much Energy is Wasted                            0
1548838	Accessing steemd RPC endpoint using Golang          0.2
1592040	The Evolution of Consciousness                      20.5
1605795	Nobody Has to Be Anybody!                           0
1728062	We Are Taught, Yet We Do Not Understand             20.3
1786011	Dockerfile for steemd, Tuned for Performance        33.7
1851608	Evaluating Martial Arts                             45.6
1872747	go-steem/rpc: Golang RPC Client Library for Steem   672
1907611	steemreduce: You Personal MapReduce for Steem       1060.5

Total pending payout: 2273.9
```

The context will be stored in `mapreduce.json` created before:

```json
{
  "config": {
    "author": "void"
  },
  "state": {
    "next_block": 1933756
  },
  "accumulator": {
    "stories": [
      {
        "block_number": 1268905,
        "title": "The Body Knows",
        "permlink": "the-body-knows",
        "pending_payout": 2.4
      },
      {
        "block_number": 1325197,
        "title": "Let it Go, Let it Happen",
        "permlink": "let-it-go-let-it-happen",
        "pending_payout": 418.7
      },
      {
        "block_number": 1473920,
        "title": "So Much Energy is Wasted",
        "permlink": "so-much-energy-is-wasted",
        "pending_payout": 0
      },
      {
        "block_number": 1548838,
        "title": "Accessing steemd RPC endpoint using Golang",
        "permlink": "accessing-steemd-rpc-endpoint-using-golang",
        "pending_payout": 0.2
      },
      {
        "block_number": 1592040,
        "title": "The Evolution of Consciousness",
        "permlink": "the-evolution-of-consciousness",
        "pending_payout": 20.5
      },
      {
        "block_number": 1605795,
        "title": "Nobody Has to Be Anybody!",
        "permlink": "nobody-has-to-be-anybody",
        "pending_payout": 0
      },
      {
        "block_number": 1728062,
        "title": "We Are Taught, Yet We Do Not Understand",
        "permlink": "we-are-taught-yet-we-do-not-understand",
        "pending_payout": 20.3
      },
      {
        "block_number": 1786011,
        "title": "Dockerfile for steemd, Tuned for Performance",
        "permlink": "dockerfile-for-steemd-tuned-for-performance",
        "pending_payout": 33.7
      },
      {
        "block_number": 1851608,
        "title": "Evaluating Martial Arts",
        "permlink": "evaluating-martial-arts",
        "pending_payout": 45.6
      },
      {
        "block_number": 1872747,
        "title": "go-steem/rpc: Golang RPC Client Library for Steem",
        "permlink": "go-steemrpc-golang-rpc-client-library-for-steem",
        "pending_payout": 672
      },
      {
        "block_number": 1907611,
        "title": "steemreduce: You Personal MapReduce for Steem",
        "permlink": "steemreduce-you-personal-mapreduce-for-steem",
        "pending_payout": 1060.5
      }
    ],
    "total_pending_payout": 2273.9
  }
}
```

Next time you run the the same command again, MapReduce will start at
`next_block` as stored in `mapreduce.json`, only processing new blocks,
which can save massive amount of time.
