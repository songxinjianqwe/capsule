# Capsule

# A Simplified OCI(Open Containers Initiative) Implementation, just like runC
[![Travis-CI](https://travis-ci.org/songxinjianqwe/capsule.svg)](https://travis-ci.org/songxinjianqwe/capsule)
[![GoDoc](https://godoc.org/github.com/songxinjianqwe/capsule?status.svg)](http://godoc.org/github.com/songxinjianqwe/capsule)
[![codecov](https://codecov.io/github/songxinjianqwe/capsule/coverage.svg)](https://codecov.io/gh/songxinjianqwe/capsule)
[![Report card](https://goreportcard.com/badge/github.com/songxinjianqwe/capsule)](https://goreportcard.com/report/github.com/songxinjianqwe/capsule)

[https://github.com/songxinjianqwe/capsule](https://github.com/songxinjianqwe/capsule)

## Project Structure
`Capsule` containers a cli and a library, which providers atomic operations of container

## Features
containers created by capsule could provide
- namespace support, including uts,pid,mount,network namespaces.
- control group(linux cgroups) support, including cpu and memory.
- rootfs provided by user, and pivot root.
- network, including container-to-container and container-to-host(todo)
- various container CLI operations, including `list`,`state`,`create`,`run`,`start`,`kill`,`delete`,`exec`,`ps`,`log` and `spec`.

Note: Union FS and image support will be supported in `capsule-daemon`.

## Install
1. `go get "github.com/songxinjianqwe/capsule"`
2. `go install $GOPATH/src/github.com/songxinjianqwe/capsule`
3. `cd $GOPATH/bin`
4.`./capsule`

## Usage

### list
#### `use`
`./capsule list`
#### example


