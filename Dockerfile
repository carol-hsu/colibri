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

FROM golang:1.18-alpine as builder
ARG VERSION=0.1
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# build
WORKDIR /coli-build/
COPY . .
RUN GO111MODULE=on go mod download
RUN go build -o colibri scraper.go request.go util.go path_finder.go
RUN go build -o colibri_v2 scraper_v2.go request.go util.go path_finder.go

# runtime image
FROM gcr.io/google_containers/ubuntu-slim:0.14
COPY --from=builder /coli-build/colibri /usr/bin/colibri
COPY --from=builder /coli-build/colibri_v2 /usr/bin/colibri_v2
CMD ["colibri_v2"]
