#!/usr/bin/env bash

go build . && xk6 build latest --with xk6-k9-amqp=.
