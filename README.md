# It's UART

[![Build Status](https://travis-ci.org/hyeoncheon/uart.svg?branch=master)](https://travis-ci.org/hyeoncheon/uart)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyeoncheon/uart)](https://goreportcard.com/report/github.com/hyeoncheon/uart)
[![Code Climate](https://codeclimate.com/github/hyeoncheon/uart/badges/gpa.svg)](https://codeclimate.com/github/hyeoncheon/uart)
[![Coverage Status](https://coveralls.io/repos/github/hyeoncheon/uart/badge.svg?branch=master)](https://coveralls.io/github/hyeoncheon/uart?branch=master)

UART is an Universal Authorizaion, Role and Team management service software.

UART was developed to succeed my old SiSO project, the original SSO service
for Hyeoncheon Project. (which was developed with Ruby on Rails framework
with well known Devise, OmniAuth and other open source components.)

UART is written in Go Language and also is built upon many open source
software modules including
[OSIN OAuth2 server library](https://github.com/RangelReale/osin)
and powered by open source
[Buffalo Go web development eco-system](https://github.com/gochigo/buffalo).

## Feature

The main features are below:

* Support sign on/in with social network accounts
  * currently Google, Facebook, and Github accounts are allowed.
* (Future Plan) Email address based local authentication will be added soon.
  * This will be used as One-Time-Password option for other authentication.
* Work as OAuth2 Provider to act as SSO authenticator for family projects.
* OAuth2 Client App management with optional role based authorization.
  * Role management per each apps.
* Support standard OAuth2 authorization process.
  * The format of Access Token is JWT(JSON Web Token).
  * Also provide `/userinfo` API endpoint.
* Member management and per App roles.

## Install

Installation procedure for Ubuntu Linux.

### Requirement

#### Essential Build Environment

```console
$ sudo apt-get update
$ sudo apt-get install build-essential
$ 
```

#### Install Golang

```console
$ sudo mkdir -p /opt/google
$ cd /opt/google/
$ wget -nv https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz -O - |sudo tar zx
$ sudo mv go go-1.8.3
$ sudo ln -s go-1.8.3 go
$ cat >> ~/.bashrc <<EOF
> 
> ## GOLANG
> export GOPATH="\$HOME/go"
> export GOROOT="/opt/google/go"
> export PATH="\$PATH:\$GOPATH/bin:\$GOROOT/bin"
> 
> EOF
$ 
$ # source bashrc or restart the shell
$ mkdir $GOPATH
$ cd $GOPATH
$ 
```

#### Install Node.js with nvm

```console
$ curl -o- https://raw.githubusercontent.com/creationix/nvm/v0.33.2/install.sh | bash
$ 
$ # source bashrc or restart the shell
$ nvm --version
0.33.2
$ nvm ls-remote --lts |tail -2
        v6.11.1   (LTS: Boron)
        v6.11.2   (Latest LTS: Boron)
$ nvm install lts/boron
$ node --version
v6.11.2
$ npm --version
3.10.10
$ 
```


### Install and Build UART

```console
$ mkdir -p $GOPATH/src/github.com/hyeoncheon
$ cd $GOPATH/src/github.com/hyeoncheon
$ git clone https://github.com/hyeoncheon/uart.git
$ cd uart
$ go get -t -v ./...
$ go get -u github.com/gobuffalo/buffalo/buffalo
$ # buffalo setup
$ npm install --no-progress
$ buffalo build --static
$ ls bin/uart
$ 
```

## Usage


### Configure


### Run


## TODO

## Author

Yonghwan SO https://github.com/sio4

## Copyright (GNU General Public License v3.0)

Copyright 2016 Yonghwan SO

This program is free software; you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation; either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program; if not, write to the Free Software Foundation, Inc., 51
Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA

