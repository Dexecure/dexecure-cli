# Dexecure-cli

Tested on Linux and Mac. 

## Installation
- Install go  
- go get github.com/dexecure/dexecure-cli  
- go install github.com/dexecure/dexecure-cli  

## To update
go get -u github.com/dexecure/dexecure-cli

## To enable autocomplete
PROG=dexecure-cli source $GOPATH/src/gopkg.in/urfave/cli.v1/autocomplete/bash_autocomplete # for bash  
PROG=dexecure-cli source $GOPATH/src/gopkg.in/urfave/cli.v1/autocomplete/zsh_autocomplete # for zsh  

## Commands available
dexecure-cli login  
dexecure-cli logout
dexecure-cli usage

dexecure-cli distribution add

dexecure-cli distribution ls  
dexecure-cli distribution ls your-distribution-uuid

dexecure-cli distribution enable your-distribution-uuid  
dexecure-cli distribution disable your-distribution-uuid  
dexecure-cli distribution clear your-distribution-uuid  
dexecure-cli distribution rm your-distribution-uuid  
