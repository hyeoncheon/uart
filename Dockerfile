# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
FROM gobuffalo/buffalo:v0.17.3 as builder

RUN mkdir -p /build
WORKDIR /build

# cache npm and go packages, unless they are changed
COPY package.json yarn.lock ./
RUN yarn install --no-progress
COPY go.mod go.sum ./
RUN go mod download
# then copy source tree and build
ADD . .
RUN HC_ROOT=/ scripts/setup.sh


FROM alpine
RUN apk add --no-cache bash ca-certificates
COPY --from=builder /uart /uart
WORKDIR /uart

# Uncomment to run the binary in "production" mode:
#ENV GO_ENV=production
ENV ADDR=0.0.0.0
EXPOSE 3000

CMD /uart/bin/uart migrate && /uart/bin/uart
