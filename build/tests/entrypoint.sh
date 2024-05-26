#!/bin/bash

bash -c "/opt/simda/simda -c /etc/simda/config.yml &" && \
go test -v -race /app/tests/...
