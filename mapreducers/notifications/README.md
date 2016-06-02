# MapReduce: notifications

This MapReduce is actually not computing any statistics. It is able, though,
to trigger notifications when certain event is detected on the blockchain.

There is no block range config available as of now, this MapReduce always
starts with the last block available on the blockchain and keeps processing
new blocks as they arrive.

## Dependencies

You need to have [steemd](https://steem.io/documentation/how-to-build/) running locally.

## Usage

Fist we need to set the path to the data directory. The default path is
`./steemreduce_data/notifications`, but it can be changed:

```bash
export STEEMREDUCE_PARAMS_DATA_DIR=./data
```

Let's assume we are using the default value for now and let's create
the configuration file in `./steemreduce_data/notifications/config.yml`.
A template can be found in `config.example.yml`. What you can found there is
briefly explained in the following sections of this README.

Once the config file is in place, we are ready to run:

```bash
steemreduce \
	-rpc_endpoint="ws:$(docker-machine ip default):8090"
	-mapreduce_id=notifications
```

This will start `steemreduce` and it will start watching new blocks for matching
operations as specified in the configuration file.

Every time there is an operation match, the configured notifiers are used to
send out a notifications.

## Available Events to Watch

* Story published/edited
* Story vote cast
* Comment published/edited
* Comment vote cast

The events can be filtered of course. The filtering is apparent from the
exemplar configuration file. Everything you need for starters is available, e.g.
you can filter stories created by author or tag.

## Available Notification Modules

You need to enable one or more notification module to make this MapReduce work.

### Command

The `command` module just runs arbitrary executable when an event is emited.
Check `config.example.yml` to see how templating can be used to insert relevant
data into the command.

The example in `config.example.yml` uses `terminal-notifier` to trigger Mac OS X
desktop notifications, but any other command can be invoked.

![Steemit Mac OS X Desktop Notifications](https://ipfs.pics/ipfs/QmP52GSJ9Fpb1MLVor66rWnuG9pYvxwR59g1gsBZNw1cd1)

### Email

The `email` module simply send an email when an event is emitted. As of now it
is not possible to change the email templates, they are hard-coded.

![Steemit Email Notifications](https://ipfs.pics/ipfs/QmfPTZkEyo1VLuDKM7igyQwdYpThWxfJ3x69ndzurJ9GB6)
