FROM golang:1.21-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev \
        git make \
        mingw-w64

WORKDIR /src
COPY go.mod go.sum ./

RUN go mod download -x
