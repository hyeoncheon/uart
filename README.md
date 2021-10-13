# It's UART

[![Build Status](https://app.travis-ci.com/hyeoncheon/uart.svg?branch=master)](https://app.travis-ci.com/hyeoncheon/uart)
[![Go Report Card](https://goreportcard.com/badge/github.com/hyeoncheon/uart)](https://goreportcard.com/report/github.com/hyeoncheon/uart)
[![Maintainability](https://api.codeclimate.com/v1/badges/912df6609e6cb8da3576/maintainability)](https://codeclimate.com/github/hyeoncheon/uart/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/912df6609e6cb8da3576/test_coverage)](https://codeclimate.com/github/hyeoncheon/uart/test_coverage)
[![Coverage Status](https://coveralls.io/repos/github/hyeoncheon/uart/badge.svg?branch=master)](https://coveralls.io/github/hyeoncheon/uart?branch=master)
[![codecov](https://codecov.io/gh/hyeoncheon/uart/branch/master/graph/badge.svg?token=F95PxOlGec)](https://codecov.io/gh/hyeoncheon/uart)

UART is a management service for Universal Authorization, Authentication,
Roles, and Teams for the Hyeoncheon project.

UART was developed as a successor of my old SiSO project, the original SSO
service for the Hyeoncheon project. (which was developed with Ruby on Rails
on top of well-known Devise, OmniAuth, and other opensource components.)

UART is written in Go (golang) and also is built upon many open source
software modules including
[OSIN OAuth2 server library](https://github.com/RangelReale/osin)
and powered by open source
[Buffalo Go web development eco-system](https://github.com/gochigo/buffalo).



## Feature

The main features are:

* Supports sign on/in with social network accounts
  * currently Google, Facebook, and Github accounts are supported.
* (Future Plan) Email address based local authentication will be added.
  * This will be used as a One-Time-Password option for other authentication.
* Works as OAuth2 Provider to provide SSO service for family projects.
* OAuth2 Client App management with optional role based authorization.
  * Role management per each application.
* Supports standard OAuth2 authorization process.
  * The format of Access Token is JWT(JSON Web Token).
  * Also provide `/userinfo` API endpoint.
* Member management and per App roles.



## Install

Installation procedure for Ubuntu Linux.


### Requirement

To build UART, you need a golang development environment, node.js, and
gobuffalo. Also, a database like MySQL is required to run UART.

The separated document
[Requirements to Build/Run UART](Requirements.md) could be a good reference
if you are not prepared with the environment and need a reference.


### Get and Build UART

Clone this repository first.

```console
$ cd $YOUR_WORKSPACE
$ git clone https://github.com/hyeoncheon/uart.git
$ cd uart
```

Then run the following commands to get related packages.

```console
$ go mod tidy
warning: ignoring symlink /home/sio4/git/hyeoncheon/uart/assets/themes/admin
go: downloading github.com/golang-jwt/jwt v3.2.2+incompatible
go: downloading golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
<...>
$ yarn
yarn install v1.22.11
[1/4] Resolving packages...
[2/4] Fetching packages...
info fsevents@2.3.2: The platform "linux" is incompatible with this module.
info "fsevents@2.3.2" is an optional dependency and failed compatibility check. Excluding it from installation.
[3/4] Linking dependencies...
[4/4] Building fresh packages...
Done in 10.32s.
$ 
```

Prepare `database.yml`.

```console
$ cp database.yml.dist database.yml
$ $EDIT database.yml
```

and build the binary.

```console
$ buffalo build
warning: ignoring symlink /home/sio4/git/hyeoncheon/uart/assets/themes/admin
warning: ignoring symlink /home/sio4/git/hyeoncheon/uart/assets/themes/admin
$ ls -lh bin/uart
-rwxrwxr-x 1 sio4 sio4 27M  9월  5 19:13 bin/uart
$ ls -sh public/assets/*.*
132K public/assets/23f19bb08961f37aaf692ff943823453.eot
 36K public/assets/77206a6bb316fa0aded5083cc57f92b9.eot
200K public/assets/9bbb245e67a133f6e486d8d2545e14a5.eot
2.8M public/assets/application.8fe89f055dac055f617c.js
4.0K public/assets/application.8fe89f055dac055f617c.js.LICENSE.txt
3.5M public/assets/application.fbee4f0dc0b49d557c81.css
4.0K public/assets/hyeoncheon.866b2e6102f939c332e1.css
   0 public/assets/hyeoncheon.8fe89f055dac055f617c.js
4.0K public/assets/manifest.json
$ 
```


### Install Files

UART has assets to be installed with it. Configure `$UART_HOME`, generate
SSL keys, then install the binary and all assets into `$UART_HOME`.

`uart.conf` and `database.yml` are provided as a sample. Please modify them
with your conditions.

```console
$ export UART_HOME=/opt/hyeoncheon/uart
$ mkdir -p $UART_HOME/bin
$ scripts/keygen.sh
$ install bin/uart $UART_HOME/bin
$ cp -a files messages $UART_HOME
$ cp supports/uart.service $UART_HOME
$ cp supports/uart.conf.dist $UART_HOME/uart.conf
$ cp database.yml $UART_HOME/database.yml
$ $EDITOR $UART_HOME/uart.conf
$ 
```

Assets also include a service description. Register it as a system service.

```console
$ sudo systemctl enable $UART_HOME/uart.service
$ sudo systemctl is-enabled uart
enabled
$ 
```



## Setup and Run

Mostly done. But UART needs some more preparation to be ready to run.

### Configure Database

UART is backed by a database. You need to configure the database before
running it.

(Please make sure if you configure `database.yml` before running the
commands.)

For development, run the following command. The default environment is
`development` so we can omit the configuration. The output can be different
for each database engine. The following is for MySQL.

```console
$ buffalo pop create && buffalo pop migrate
v5.3.0

[POP] 2021/09/05 20:01:41 info - create hc_uart_development (hyeoncheon:hyeoncheon@(localhost:3306)/hc_uart_development?collation=utf8mb4_general_ci&multiStatements=true&readTimeout=10s&parseTime=true)
[POP] 2021/09/05 20:01:41 info - created database hc_uart_development
v5.3.0

[POP] 2021/09/05 20:01:42 info - > uart
[POP] 2021/09/05 20:01:42 info - > messaging
[POP] 2021/09/05 20:01:42 info - > docs
[POP] 2021/09/05 20:01:42 info - Successfully applied 3 migrations.
[POP] 2021/09/05 20:01:42 info - 0.9432 seconds
mysqldump: [Warning] Using a password on the command line interface can be insecure.
mysqldump: Error: 'Access denied; you need (at least one of) the PROCESS privilege(s) for this operation' when trying to dump tablespaces
[POP] 2021/09/05 20:01:42 info - dumped schema for hc_uart_development
$ 
```

For production mode, you can run the following command. (or you can use the
same command above if you already exported the `GO_ENV` environment variable.)

```console
$ GO_ENV=production buffalo db create && GO_ENV=production buffalo db migrate
$ 
```


### Preparing Social Logins

Currently, UART supports login via Google, Facebook, and Github. Before using
them, you need to configure them from their own websites.

* https://console.cloud.google.com/apis/credentials
* https://developers.facebook.com/apps/
* https://github.com/organizations/YOUR-ORG/settings/applications

Then configure environment variables for them

```
export GPLUS_KEY="..."
export GPLUS_SECRET="..."
export FACEBOOK_KEY="..."
export FACEBOOK_SECRET="..."
export GITHUB_KEY="..."
export GITHUB_SECRET="..."
```

Note: UART does not support enabling/disabling selectively for now. You need
to configure them all, otherwise, users will see errors when they click on
unconfigured login link.


### Configure Mailgun

The only supported email sending feature, for now, is using www.mailgun.com.
Not sure they still provide Free Plan but please check and configure it.


### Run

Wow! Such a long configuration steps! but now we are ready to run!

```console
$ sudo systemctl start uart
$ sudo systemctl status uart
● uart.service - UART server
   Loaded: loaded (/opt/hyeoncheon/uart/uart.service; linked; vendor preset: enabled)
   Active: active (running) since Wed 2017-11-08 19:03:54 KST; 30min ago
 Main PID: 15264 (uart)
    Tasks: 8
   Memory: 7.7M
      CPU: 352ms
   CGroup: /system.slice/uart.service
           └─15264 /opt/hyeoncheon/uart/uart

<...>
$ 
```


### Run in Development Mode

Well, we still need a test. The following script is what I used for dev mode
execution.

```bash
#!/bin/bash
# environment for uart
# vim: set nowrap syntax=sh:

export GO_ENV='development'
export SESSION_SECRET='fdb3...55b9'
export SESSION_NAME='_uart_session'
export HOST='http://localhost:3000'

export GPLUS_KEY='8730....apps.googleusercontent.com'
export GPLUS_SECRET='c4m1...vwTu'
export GITHUB_KEY='50d4...b4ab'
export GITHUB_SECRET='1cf2...9ba5'
export FACEBOOK_KEY='4231...3146'
export FACEBOOK_SECRET='d4ed...b40f'
#export FACEBOOK_KEY='3201...5981'
#export FACEBOOK_SECRET='6bd6...ee96'
export CF_KEY='b8B2...Hcwv'
export CF_SECRET='Vt7D...lCbT'

export MG_API_KEY='key-78...cf53'
export MG_DOMAIN='mg.example.com'
export MG_URL='https://api.mailgun.net/v3'
export MAIL_SENDER='C-3PO <c3po@example.com>'

buffalo dev
```



## OK, Show Me the Shots

#### Login Screen

![UART Login](docs/uart-login.png)

#### Register New App

Each family app should be registered here as the same as we registered UART
on Google, Facebook, and Github. By doing this, users of UART will be able
to login to those family apps.

![UART New App](docs/uart-new-app.png)

#### Registered Apps

![UART Apps](docs/uart-apps.png)

#### App Details

Application managers can configure their own application's OAuth2 settings
and its own roles.

![UART App Details](docs/uart-app-details.png)

#### Membership

Users can see their registered applications as a user, Messengers, Teams,
and Credentials. Also, they can request roles for each application. E.g.
A user can be a user of App-A, a manager of App-B, while they all are
basically a user of UART itself.

![UART Membership](docs/uart-membership.png)


## TODO

* Team support
* Email login

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

