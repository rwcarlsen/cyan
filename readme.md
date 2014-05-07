
# CyAn - Cyclus Analysis Tools

To install, you need the Go toolchain.  You can get it from
http://golang.org/doc/install or you can use your favorite linux
distribution's package manager:

```
# debain-based distros
apt-get install golang

# archlinux
pacman -S go
```

You should make a directory to use as your GOPATH and set the GOPATH
environment variable to it.  The Go tool will install packages into this
directory.  For convenience, you should also add `$GOPATH/bin` to your PATH so
that binaries from fetched packages are directly accessible on the command
line.

There are two binary tools:

* `cycpost` - for post processing a Cyclus sqlite output database.  Creates an
  "Inventories" table and combines the "AgentEntry" and "AgentExit" table into
  the "Agents" table.

* `metric` - perform various queries on a Cyclus sqlite output database.  Some
  queries require `cycpost` to be run on the database first

Both commands have various flags and subcommands that can be viewed with the
`-h` flag:

```
cycpost -h
metric -h
metric -db foo.sqlite [subcmd] -h
```
