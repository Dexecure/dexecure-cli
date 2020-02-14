# Dexecure-cli

Tested on Linux and Mac.

## Installation

- Install go
- go get github.com/dexecure/dexecure-cli
- go install github.com/dexecure/dexecure-cli

To use the CLI from anywhere on your terminal, make sure that the bin folder of go is added to your \$PATH variable

## To update the CLI

go get -u github.com/dexecure/dexecure-cli

## To enable autocomplete

PROG=dexecure-cli source $GOPATH/src/gopkg.in/urfave/cli.v1/autocomplete/bash_autocomplete # for bash  
PROG=dexecure-cli source $GOPATH/src/gopkg.in/urfave/cli.v1/autocomplete/zsh_autocomplete # for zsh

## Commands available

dexecure-cli configure

dexecure-cli usage

dexecure-cli domain add

dexecure-cli domain ls  
dexecure-cli domain ls id your-domain-uuid  
dexecure-cli domain ls all  
dexecure-cli domain ls website your-website-uuid

dexecure-cli domain clear your-domain-uuid  
dexecure-cli domain rm your-domain-uuid

dexecure-cli website add  
dexecure-cli website ls id your-website-uuid  
dexecure-cli website rm your-website-uuid
