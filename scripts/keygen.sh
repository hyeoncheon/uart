#!/bin/sh
#

prefix=files/jwt

openssl genrsa -out $prefix.private.pem 1024 && \
openssl rsa -in $prefix.private.pem -outform PEM -pubout -out $prefix.public.pem

