# Copyright 2022 Carol Hsu
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG CGROUP_VERSION=2

FROM golang:1.27-alpine AS builder
ARG VERSION=0.1
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ARG CGROUP_VERSION

# build
WORKDIR /coli-build/
COPY . .
RUN GO111MODULE=on go mod download

RUN if [ "$CGROUP_VERSION" = "2" ] ; then \
        go build -o colibri scraperv2.go ; \
    elif [ "$CGROUP_VERSION" = "1" ] ; then \
        go build -o colibri scraper.go ; \
    else \
        echo "Please indicate proper CGROUP_VERSION based on your OS." ; \
    fi

# runtime image
FROM gcr.io/google_containers/ubuntu-slim:0.14
ARG CGROUP_VERSION

COPY --from=builder /coli-build/colibri /usr/bin/
ENTRYPOINT ["/usr/bin/colibri"]
