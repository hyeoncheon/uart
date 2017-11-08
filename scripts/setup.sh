#!/bin/bash
#
# vim: set ts=4 sw=4:

set -xe

HC_ROOT=/opt/hyeoncheon
UART_HOME=$HC_ROOT/uart

if [ ! -f "database.yml" ]; then
	echo "'database.yml' not exists. please create this file before run."
	echo "'database.yml.dist' is good start point for this."
	exit
fi

if [ ! -f "uart.conf" ]; then
	echo "'uart.conf' not exists. please create this file before run."
	echo "'supports/uart.conf.dist' is good start point for this."
	exit
fi

# ensure package dependancy, vendoring
go get -u github.com/golang/dep/cmd/dep
dep ensure

# setup buffalo environment and build
npm install --no-progress
go get -u github.com/gobuffalo/buffalo/buffalo
buffalo build --static

# setup database
buffalo db create && buffalo db migrate
GO_ENV=production buffalo db create && GO_ENV=production buffalo db migrate

# install files
scripts/keygen.sh
mkdir -p $UART_HOME
install bin/uart $UART_HOME
cp -a messages files locales templates $UART_HOME
cp -a supports/uart.service $UART_HOME
cp -a uart.conf $UART_HOME

# setup systemd
sudo ln -s /opt/hyeoncheon/uart/uart.service /etc/systemd/system/

