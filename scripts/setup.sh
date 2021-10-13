#!/bin/bash
#
# vim: set ts=4 sw=4:

[ "$HC_ROOT" = "" ] && HC_ROOT=/opt/hyeoncheon
UART_HOME=$HC_ROOT/uart

if [ ! -f "database.yml" ]; then
	echo "'database.yml' does not exists. create default."
	cp database.yml.dist database.yml
fi

set -xe

# ensure package dependancy, yarn, and build
go mod tidy
yarn install --no-progress
buffalo build --static --tags netgo --clean-assets -o bin/uart

# setup database
#buffalo db create && buffalo db migrate
#GO_ENV=production buffalo db create && GO_ENV=production buffalo db migrate

# install files
mkdir -p $UART_HOME/bin
scripts/keygen.sh
install bin/uart $UART_HOME/bin
cp -a files messages $UART_HOME
cp -a supports/uart.service $UART_HOME
if [ -f "uart.conf" ]; then
	cp -a uart.conf $UART_HOME/
else
	cp -a supports/uart.conf.dist $UART_HOME/uart.conf
fi

# setup systemd
#sudo ln -s $UART_HOME/uart.service /etc/systemd/system/

