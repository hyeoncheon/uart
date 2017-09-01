#!/bin/sh
#
# manually migrate development database to test database and run goveralls.
#
#  repository: https://github.com/mattn/goveralls

buffalo db drop -e test
buffalo db create -e test
buffalo db schema dump -e development
buffalo db schema load -e test

./scripts/keygen.sh 

export GO_ENV="test"
export UART_HOME=`pwd`
goveralls -repotoken $COVERALLS_TOKEN

exit
###
### this script stops here!
###

### manual testing with package loop.
### `buffalo test` does not support -coverprofile options.
for p in `go list ./...`; do
	echo $p |grep "/vendor/" && continue
	pname=`basename $p`
	go test -p 1 -cover -coverprofile cover.$pname.out $p
done
