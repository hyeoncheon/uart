#!/bin/bash

# mkdir -p $GOPATH/src/github.com/hyeoncheon
# cd $GOPATH/src/github.com/hyeoncheon
# git clone https://github.com/hyeoncheon/uart.git
# cd uart
go get -t -v ./...
go get -u github.com/gobuffalo/buffalo/buffalo
# buffalo setup
npm install --no-progress
buffalo build --static
ls bin/uart
