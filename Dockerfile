# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM gobuffalo/buffalo:v0.18.7 as builder

ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /src/github.com/hyeoncheon/uart
WORKDIR /src/github.com/hyeoncheon/uart

# this will cache the npm install step, unless package.json changes
COPY package.json .
COPY yarn.lock .
COPY .yarn* .
RUN yarn install
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .
#RUN buffalo build --static -o /bin/app
RUN HC_ROOT=/ scripts/setup.sh

### check and update package version here: https://pkgs.alpinelinux.org/packages
FROM alpine:latest
RUN apk add --no-cache 'bash=~5.1' 'ca-certificates>=20211220-r0'

WORKDIR /uart/

COPY --from=builder /uart /uart

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD ["/uart/bin/uart"]
