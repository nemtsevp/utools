# utools

## Installation

Install [Go](https://golang.org/) and run:

    $ go get -v github.com/multicharts/utools/...
    $ export PATH="$PATH:$GOPATH/bin"

## ucat

`ucat` implements portable unbounded pipe.

It reads data from stdin and writes it to stdout just like `cat`, but uses unbounded memory buffer, so that writing to `ucat` stdin is near to nonblocking, if you have enough memory.

Usage:

    producer | ucat | consumer

## uwatch

`uwatch` is a kind of transparent watchdog. It runs child process and kills it when it receives death signal or its parent dies. It exits with exit status of its child.

Usage:

    uwatch command -with -arguments
