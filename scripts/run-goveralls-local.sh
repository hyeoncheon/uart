#!/bin/sh
#
# manually migrate development database to test database and run goveralls.
#
#  repository: https://github.com/mattn/goveralls

buffalo pop drop -e test
buffalo pop create -e test
buffalo pop migrate -e test
# bad way while using on test machine:
# buffalo pop schema dump -e development && buffalo pop schema load -e test

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
