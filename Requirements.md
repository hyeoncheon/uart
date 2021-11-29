# Requirements to Build/Run UART

To build UART, you need a golang development environment, node.js, and
gobuffalo. Also database like MySQL is required to run UART.



## Essential Build Environment

If you want to install gobuffalo manually with sqlite3 support or in similar
situation, you may need `build-essential` package.

```console
$ sudo apt update
$ sudo apt install build-essential
```



## Install Golang

The following is my standard way to install golang (version 1.16.7 which is
newest version at this moment). However, you can do the same thing in your
favorite way.  The goal is just having a golang environment.

```console
$ sudo mkdir -p /opt/google
$ cd /opt/google/
$ rm -f go
$ rm -rf go-1.16.7
$ wget -nv https://golang.org/dl/go1.16.7.linux-amd64.tar.gz -O - |sudo tar zx
$ sudo mv go go-1.16.7
$ sudo ln -s go-1.16.7 go
$ cat >> ~/.bashrc <<EOF
>
> ## GOLANG
> export GOPATH="\$HOME/go"
> export GOROOT="/opt/google/go"
> export PATH="\$PATH:\$GOPATH/bin:\$GOROOT/bin"
>
> EOF
$ # source .bashrc or restart the shell
$ mkdir $GOPATH
$ cd $GOPATH
$
```

Note that some of above steps could be modified as your prefer way.

Note that your installation is under directory `/opt/google/go-1.16.7` so
you can easily remove them by running `sudo rm -rf /opt/google/go-1.16.7`.



## Install Node.js with nvm

This is also my favorite way of installing node.js via nvm as a normal user.
This method does not require a root privilege (sudo).

```console
$ curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | bash
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 14926  100 14926    0     0  68783      0 --:--:-- --:--:-- --:--:-- 68783
=> Downloading nvm from git to '/home/sio4/.nvm'
=> Cloning into '/home/sio4/.nvm'...
remote: Enumerating objects: 348, done.
remote: Counting objects: 100% (348/348), done.
remote: Compressing objects: 100% (297/297), done.
remote: Total 348 (delta 39), reused 158 (delta 26), pack-reused 0
Receiving objects: 100% (348/348), 199.85 KiB | 7.40 MiB/s, done.
Resolving deltas: 100% (39/39), done.
* (HEAD detached at FETCH_HEAD)
  master
=> Compressing and cleaning up git repository

=> nvm source string already in /home/sio4/.bashrc
=> bash_completion source string already in /home/sio4/.bashrc
=> Close and reopen your terminal to start using nvm or run the following to use it now:

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
$
$ # source bashrc or restart the shell
$ nvm --version
0.38.0
$ nvm ls-remote --lts |grep Latest
         v4.9.1   (Latest LTS: Argon)
        v6.17.1   (Latest LTS: Boron)
        v8.17.0   (Latest LTS: Carbon)
       v10.24.1   (Latest LTS: Dubnium)
       v12.22.6   (Latest LTS: Erbium)
       v14.17.6   (Latest LTS: Fermium)
$ nvm install --lts # or nvm install lts/fermium
Installing latest LTS version.
Downloading and installing node v14.17.6...
Downloading https://nodejs.org/dist/v14.17.6/node-v14.17.6-linux-x64.tar.xz...
######################################################################### 100.0%
Computing checksum with sha256sum
Checksums matched!
Now using node v14.17.6 (npm v6.14.15)
Creating default alias: default -> lts/* (-> v14.17.6)
$ node --version
v14.17.6
$ npm --version
6.14.15
$ npm install --global yarn

> yarn@1.22.11 preinstall /home/sio4/.nvm/versions/node/v14.17.6/lib/node_modules/yarn
> :; (node ./preinstall.js > /dev/null 2>&1 || true)

/home/sio4/.nvm/versions/node/v14.17.6/bin/yarnpkg -> /home/sio4/.nvm/versions/node/v14.17.6/lib/node_modules/yarn/bin/yarn.js
/home/sio4/.nvm/versions/node/v14.17.6/bin/yarn -> /home/sio4/.nvm/versions/node/v14.17.6/lib/node_modules/yarn/bin/yarn.js
+ yarn@1.22.11
added 1 package in 0.452s
$ 
$ yarn --version
1.22.11
$
$ which node
/home/sio4/.nvm/versions/node/v14.17.6/bin/node
$ which npm
/home/sio4/.nvm/versions/node/v14.17.6/bin/npm
$ which yarn
/home/sio4/.nvm/versions/node/v14.17.6/bin/yarn
$ 
```

Finally, you have working commands of `node`, `npm`, and `yarn` in your home.
Since this method installs all packages in your home, you don't need a root
privilege and you can easily remove them by running command
`rm -rf ~/.nvm` or specific version directory under `~/.nvm/versions`



## Install Gobuffalo

Buffalo provides several ways to install and configure it. Since I want
a sqlite support, my choice is as follow.

```console
$ go install -v -tags sqlite github.com/gobuffalo/cli/cmd/buffalo@v0.18.1
go: downloading github.com/gobuffalo/cli v0.18.1
<...>
github.com/gobuffalo/cli/cmd/buffalo
$ 
```

Now you have `buffalo` command.

```console
$ buffalo version
INFO[0000] Buffalo version is: v0.18.1
$ which buffalo
/home/sio4/go/bin/buffalo
$ 
```

Run `buffalo info` in the application root then you will get more information
about the buffalo environment and the application configuration.

Buffalo has plugins and `buffalo-pop` is one of them when the application
needs database access via pop.

```console
$ go get -v -tags sqlite github.com/gobuffalo/buffalo-pop/v3
go: downloading github.com/spf13/cobra v1.2.1
go: downloading github.com/gobuffalo/pop/v6 v6.0.0
go: downloading github.com/gobuffalo/flect v0.2.4
<...>
github.com/gobuffalo/pop/v6/soda/cmd
github.com/gobuffalo/buffalo-pop/v3/cmd
github.com/gobuffalo/buffalo-pop/v3
$ 
```

`buffalo-pop` will be automatically installed with the same command if you
create or build an application using pop anyway.

With this method, buffalo cli and related packages will be installed on your
`GO_HOME` so you can easily remove them by runing `rm $GOPATH/bin/buffalo*`
and `sudo rm -rf $GO_HOME/pkg/*`.



## MySQL as a database management system

This is not a mandatory step if you want use existing database you have.
However, this procedure could be referenced when you configure your own.

```console
$ sudo apt install mysql-server
Reading package lists... Done
Building dependency tree
Reading state information... Done
The following additional packages will be installed:
  libaio1 libcgi-fast-perl libcgi-pm-perl libevent-core-2.1-7
  libevent-pthreads-2.1-7 libfcgi-perl libhtml-template-perl libmecab2
  mecab-ipadic mecab-ipadic-utf8 mecab-utils mysql-client-8.0
  mysql-client-core-8.0 mysql-server-8.0 mysql-server-core-8.0
Suggested packages:
  libipc-sharedcache-perl mailx tinyca
The following NEW packages will be installed:
  libaio1 libcgi-fast-perl libcgi-pm-perl libevent-core-2.1-7
  libevent-pthreads-2.1-7 libfcgi-perl libhtml-template-perl libmecab2
  mecab-ipadic mecab-ipadic-utf8 mecab-utils mysql-client-8.0
  mysql-client-core-8.0 mysql-server mysql-server-8.0 mysql-server-core-8.0
0 upgraded, 16 newly installed, 0 to remove and 0 not upgraded.
Need to get 31.4 MB of archives.
After this operation, 261 MB of additional disk space will be used.
<...>
```

and configure default user for Hyeoncheon project:

```console
$ sudo mysql -u root << EOF
> CREATE USER hyeoncheon@localhost IDENTIFIED BY 'password';
> GRANT ALL PRIVILEGES ON \`hc_%\`.* TO hyeoncheon@localhost;
> EOF
$
$ sudo mysql -u root -e 'SHOW GRANTS FOR hyeoncheon@localhost;'
+--------------------------------------------------------------+
| Grants for hyeoncheon@localhost                              |
+--------------------------------------------------------------+
| GRANT USAGE ON *.* TO `hyeoncheon`@`localhost`               |
| GRANT ALL PRIVILEGES ON `hc_%`.* TO `hyeoncheon`@`localhost` |
+--------------------------------------------------------------+
$
```



Ready to Go!
