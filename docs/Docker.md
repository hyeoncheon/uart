# Ship It on the Cloud: Docker

If you want to deploy UART on the cloud, the easiest way to ship it on the
cloud is using a docker container.



## Build Your Own Docker Image

UART provides a `Dockerfile` which uses a multi-stage build. UART was built
on top of gobuffalo and buffalo provides a builder image for this purpose.

Run `docker build` command within the UART source tree to build docker image.
It will get builder image as below (step 1/16), and will start building UART
image in multiple stages.

Step 2 to 7 are running on the first stage and the result of those steps
(layers) is reusable. If you run the same build command again, it will use
the cached layers that already contain the result of `go mod download` and
`yarn install` which is relatively heavy task.

Step 8 to 9 are UART specific steps which build the app and preparing the
directory structure of the app.

Since we uses alpine image as a base image, we build the app as a static
binary with the following command:

```
buffalo build --static --tags netgo --clean-assets -o bin/uart
```

Step 10 to 16 are running as the second stage and it composites the final
image with already built binary and directory structure.

While build the application binary, it uses a full size builder image with
many tools. However, for the final smaller image, the second stage uses
alpine image for it.

The whole output of the build process is:

```console
$ sudo docker build -t uart .
Sending build context to Docker daemon  4.656MB
Step 1/16 : FROM gobuffalo/buffalo:v0.17.3 as builder
v0.17.3: Pulling from gobuffalo/buffalo
4c25b3090c26: Pull complete
1acf565088aa: Pull complete
b95c0dd0dc0d: Pull complete
5cf06daf6561: Pull complete
fcb5bda771c7: Pull complete
7b6a93dab1a5: Pull complete
a777b06e31ba: Pull complete
ea56e830e9ef: Pull complete
ee3b7ecbf8f9: Pull complete
302eea8e5726: Pull complete
e73552ef8838: Pull complete
bf91b247da7b: Pull complete
a8fe9da0602e: Pull complete
875f5a18c5e1: Pull complete
be413e35f3dd: Downloading  56.91MB/60.25MB
37e8ae41921e: Verifying Checksum
4f4fb700ef54: Download complete
Digest: sha256:753aef1488dd17a496edcf59de3b6041515b293d8e9671274e377c1faca62adc
Status: Downloaded newer image for gobuffalo/buffalo:v0.17.3
 ---> 57cdd3aa52d0
Step 2/16 : RUN mkdir -p /build
 ---> Running in d860e68d3df8
Removing intermediate container d860e68d3df8
 ---> 680e1dddf35f
Step 3/16 : WORKDIR /build
 ---> Running in 932458d177c8
Removing intermediate container 932458d177c8
 ---> b17691325ed4
Step 4/16 : COPY package.json yarn.lock ./
 ---> c8426f3aff6d
Step 5/16 : RUN yarn install --no-progress
 ---> Running in 454dce28e38d
yarn install v1.22.11
[1/4] Resolving packages...
[2/4] Fetching packages...
info fsevents@2.3.2: The platform "linux" is incompatible with this module.
info "fsevents@2.3.2" is an optional dependency and failed compatibility check. Excluding it from installation.
[3/4] Linking dependencies...
[4/4] Building fresh packages...
Done in 13.76s.
Removing intermediate container 454dce28e38d
 ---> ce8b3f7785b6
Step 6/16 : COPY go.mod go.sum ./
 ---> 6f1b207c8334
Step 7/16 : RUN go mod download
 ---> Running in b9861b8e4f06
Removing intermediate container b9861b8e4f06
 ---> 84f57808374e
Step 8/16 : ADD . .
 ---> 19d2c2016f33
Step 9/16 : RUN HC_ROOT=/ scripts/setup.sh
 ---> Running in fa9af6eaf26f
'database.yml' does not exists. create default.
+ go mod tidy
+ yarn install --no-progress
yarn install v1.22.11
[1/4] Resolving packages...
success Already up-to-date.
Done in 0.40s.
+ buffalo build --static --tags netgo --clean-assets -o bin/uart
# github.com/hyeoncheon/uart
/usr/bin/ld: /tmp/go-link-258082567/000002.o: in function `mygetgrouplist':
/usr/local/go/src/os/user/getgrouplist_unix.go:18: warning: Using 'getgrouplist' in statically linked applications requires at runtime the shared libraries from the glibc version used for linking
/usr/bin/ld: /tmp/go-link-258082567/000001.o: in function `mygetgrgid_r':
/usr/local/go/src/os/user/cgo_lookup_unix.go:40: warning: Using 'getgrgid_r' in statically linked applications requires at runtime the shared libraries from the glibc version used for linking
/usr/bin/ld: /tmp/go-link-258082567/000001.o: in function `mygetgrnam_r':
/usr/local/go/src/os/user/cgo_lookup_unix.go:45: warning: Using 'getgrnam_r' in statically linked applications requires at runtime the shared libraries from the glibc version used for linking
/usr/bin/ld: /tmp/go-link-258082567/000001.o: in function `mygetpwnam_r':
/usr/local/go/src/os/user/cgo_lookup_unix.go:35: warning: Using 'getpwnam_r' in statically linked applications requires at runtime the shared libraries from the glibc version used for linking
/usr/bin/ld: /tmp/go-link-258082567/000001.o: in function `mygetpwuid_r':
/usr/local/go/src/os/user/cgo_lookup_unix.go:30: warning: Using 'getpwuid_r' in statically linked applications requires at runtime the shared libraries from the glibc version used for linking
+ mkdir -p //uart/bin
+ scripts/keygen.sh
Generating RSA private key, 1024 bit long modulus (2 primes)
.........+++++
..................+++++
e is 65537 (0x010001)
writing RSA key
+ install bin/uart //uart/bin
+ cp -a files messages //uart
+ cp -a supports/uart.service //uart
+ '[' -f uart.conf ']'
+ cp -a supports/uart.conf.dist //uart/uart.conf
Removing intermediate container fa9af6eaf26f
 ---> 8517b39c497f
Step 10/16 : FROM alpine
 ---> 14119a10abf4
Step 11/16 : RUN apk add --no-cache bash ca-certificates
 ---> Using cache
 ---> 054c26d5be33
Step 12/16 : COPY --from=builder /uart /uart
 ---> 8b1b53318116
Step 13/16 : WORKDIR /uart
 ---> Running in 190e2969c550
Removing intermediate container 190e2969c550
 ---> f22ee24f2c9a
Step 14/16 : ENV ADDR=0.0.0.0
 ---> Running in 7391437d6d14
Removing intermediate container 7391437d6d14
 ---> 465e3d418102
Step 15/16 : EXPOSE 3000
 ---> Running in a1e25faf3a19
Removing intermediate container a1e25faf3a19
 ---> 034642aef6c3
Step 16/16 : CMD /uart/bin/uart migrate && /uart/bin/uart
 ---> Running in 94a7e409cf6f
Removing intermediate container 94a7e409cf6f
 ---> cfa60e42f2e1
Successfully built cfa60e42f2e1
Successfully tagged uart:latest
$ 
```



### Run a Container

Once you have an image, you can run it on any docker host.


#### Preparing Database

First, prepare the database. In this step, we need the `buffalo` command and
the `buffalo-pop` plugin.

```console
$ DATABASE_URL='mysql://username:password@(database.example.com:3306)/hc_prod?parseTime=true&multiStatements=true&readTimeout=10s' buffalo pop create -e production
v5.3.1

[POP] 2021/10/13 21:12:11 info - create hc_prod (username:password@(database.example.com:3306)/hc_prod?parseTime=true&multiStatements=true&readTimeout=10s)
[POP] 2021/10/13 21:12:14 info - created database hc_prod
$ 
```

If you need to clean up the database, you can also use the same approach with
`drop` sub-command instead of `create`.


#### Run a Container

Now run the container with the UART image. The image has the default local
database configuration but has no OAuth2 specific defaults. You need to
prepare them as your own values, and should pass them as a form of
environment variable.

```console
$ sudo docker run -it \
> -p 3000:3000 \
> --env DATABASE_URL='mysql://username:password@(database.example.com:3306)/hc_prod?parseTime=true&multiStatements=true&readTimeout=10s' \
> --env GO_ENV='production' \
> --env SESSION_SECRET='fdb3...55b9' \
> --env SESSION_NAME='_uart_session' \
> --env HOST='http://localhost:3000' \
> --env GPLUS_KEY='8730...e5i6.apps.googleusercontent.com' \
> --env GPLUS_SECRET='c4m1...vwTu' \
> --env GITHUB_KEY='50d4...b4ab' \
> --env GITHUB_SECRET='1cf2...9ba5' \
> --env FACEBOOK_KEY='4231...3146' \
> --env FACEBOOK_SECRET='d4ed...b40f' \
> --env MG_API_KEY='key-7...cf53' \
> --env MG_DOMAIN='mg.example.com' \
> --env MG_URL='https://api.mailgun.net/v3' \
> --env MAIL_SENDER='C-3PO <c3po@example.com>' \
> uart
INFO[2021-10-13T12:14:33Z] UART executed as uart (in production mode)...
INFO[2021-10-13T12:14:33Z] UART Home is /uart
DEBU[2021-10-13T12:14:33Z] invoking RegisterMessaging... category=worker
INFO[2021-10-13T12:14:33Z] messaging: set C-3PO <c3po@example.com> as mail sender category=worker
INFO[2021-10-13T12:14:40Z] new background job handler worker.Messaging 0/0 registered category=worker
DEBU[2021-10-13T12:14:40Z] invoking RegisterTimer... category=worker
INFO[2021-10-13T12:14:40Z] new background job handler worker.Timer 0/0 registered category=worker
INFO[2021-10-13T12:14:40Z] jobs registration completed! (2 handlers) category=worker
INFO[2021-10-13T12:14:40Z] models initialized category=models
INFO[2021-10-13T12:14:40Z] oauth2 provider with jwt support initialized!
INFO[2021-10-13T12:14:49Z] UART executed as uart (in production mode)...
INFO[2021-10-13T12:14:49Z] UART Home is /uart
DEBU[2021-10-13T12:14:49Z] invoking RegisterMessaging... category=worker
INFO[2021-10-13T12:14:49Z] messaging: set C-3PO <c3po@example.com> as mail sender category=worker
INFO[2021-10-13T12:14:52Z] new background job handler worker.Messaging 0/0 registered category=worker
DEBU[2021-10-13T12:14:52Z] invoking RegisterTimer... category=worker
INFO[2021-10-13T12:14:52Z] new background job handler worker.Timer 0/0 registered category=worker
INFO[2021-10-13T12:14:52Z] jobs registration completed! (2 handlers) category=worker
INFO[2021-10-13T12:14:52Z] models initialized category=models
INFO[2021-10-13T12:14:52Z] oauth2 provider with jwt support initialized!
INFO[2021-10-13T12:14:52Z] Starting application at http://0.0.0.0:3000
INFO[2021-10-13T12:14:52Z] Starting Simple Background Worker
INFO[2021-10-13T12:16:26Z] / content_type=text/html db=0s duration=1.813886087s human_size="3.6 kB" method=GET path=/ render=4.867255ms request_id=857947e33d26b8aa4eb1-ad88f58e808970147725 size=3613 status=200
INFO[2021-10-13T12:16:30Z] /login/ content_type=text/html db=0s duration=658.444717ms human_size="4.0 kB" method=GET path=/login/ render=4.301321ms request_id=857947e33d26b8aa4eb1-49a381811742e60eaa80 size=3991 status=200
INFO[2021-10-13T12:16:33Z] /auth/gplus/ content_type=text/html db=0s duration=422.437945ms human_size="414 B" method=GET path=/auth/gplus/ request_id=857947e33d26b8aa4eb1-47f522fabd719dfd03c3 size=414 status=307
<...>
INFO[2021-10-13T12:16:56Z] FIRST FLIGHT! register my self UART category=models
<...>
```

That's it!
