
# CyAn - Cyclus Analysis Tools

CyAn contains command line tools for post processing and analyzing Cyclus
simulation databases (http://fuelcycle.org). It is still experimental and not
all features have been validated/verified to be correct.

## Installation

To install, you need the Go toolchain.  You can get it from
http://golang.org/doc/install or you can use your favorite linux
distribution's package manager:

```
# debain-based distros
apt-get install golang

# archlinux
pacman -S go

# macports
port install go
```

You should make a directory to use as your GOPATH and set the GOPATH
environment variable to it.  The Go tool will install packages into this
directory.  For convenience, you should also add `$GOPATH/bin` to your PATH so
that binaries from fetched packages are directly accessible on the command
line.  When you are ready, run:

```
go get github.com/rwcarlsen/cyan/...

```

## Usage

There are two binary tools:

* `cycpost` - for post processing a Cyclus sqlite database.  Creates an
  "Inventories" table and combines the "AgentEntry" and "AgentExit" table into
  the "Agents" table.

* `metric` - perform various queries on a Cyclus sqlite database.  Some
  queries require `cycpost` to be run on the database first.

Both commands have various flags and subcommands that can be viewed with the
`-h` flag:

```
cycpost -h
metric -h
metric -db foo.sqlite [subcmd] -h
```

Some quick examples:

```
# post process the db
cycpost cyclus.sqlite

# output a png graph of the flow of all material between agents t=2 to t=7
metric -db cyclus.sqlite flowgraph -t1=2 -t2=7 > flow.dot
dot -Tpng -o flow.png flow.dot

# output a time series of active deployments for all AP1000
metric -db cyclus.sqlite deployseries AP1000
```
