# steemreduce

[![CircleCI](https://circleci.com/gh/tchap/steemreduce.svg?style=svg)](https://circleci.com/gh/tchap/steemreduce)

Your personal MapReduce for Steem - https://steem.io

## About

`steemreduce` does just one thing - it runs MapReduce over the given range of blocks,
thus allowing you to collect data and generate your own statistics. Cool, right?

## Usage

`steemreduce` is available for you already compiled on CircleCI, so

1. Go to [CircleCI](https://circleci.com/gh/tchap/steemreduce/tree/master).
2. Choose the build you want, probably the latest one.
3. Append `#artifacts` to the URL and press Enter. This will reload the page
   and expand the Artifacts tab so that you can choose the executable
   for your plarform and download it.
4. There are some MapReduce implementations already included in `steemreduce`.
   Please check the respective `README` files in `mapreducers/<mapreduce_id>`
   to see how to configure the desired MapReduce.

## More Handy MapReduce Implementations

In case there is a MapReduce you would like to have implemented, send me a
private message on Steem Slack, I am there as `@void`. I will write your
requested MapReduce and publish it at https://steemit.com as a standalone story
which you can upvote. Or you can send me some STEEM directly if you so decide.

## Issues

In case you find any issue, please file a report in the issue for this
repository.

## Contributions

Feel free to send a pull request. Use `develop` branch as the base.

## License

`MIT`, see the `LICENSE` file.
