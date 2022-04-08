# ez

Easy to use cross-platform p2p file transfer tool for your local network.

## Getting started

There are 4 types of entities that participate in the transfer of a file:

- tracker(binary executable named `ezt`)
- seeder(binary executable named `ezs`)
- seeder client(binary executable named `ezl`)
- client(binary executable named `ez`)

There needs to be one tracker and at least one seeder in order to start downloading files.

The tracker, as the name suggests, is responsible to track where the files are(i.e. it stores information about the seeders for each file)
and to respond to queries from the client about the files.

The seeder has a list of files which it makes available for download and serves file chunks when requested by clients.

The seeder client manages the set of files which the seeder is responsible for.

The client lists the available files and downloads them if requested.


### Install

- [install golang](https://golang.org/doc/install)
- download binaries:

```
go get -ldflags="-s -w" github.com/aburdulescu/ez/cmd/...
```

### Simple setup

You need at least 2 machines(containers, VMs, native).

One will be the client and the other will be the server(tracker+seeder).

#### Machine1(server)

- setup the tracker

This is simple, just run the binary `ezt`.

- setup the seeder

The seeder needs some config flags when it's started:

```
Usage of ezs:
  -dbpath string
        path where the database is stored (default "./seeder.db")
  -disable-log
        disable logging
  -seedaddr string
        address to used by peers
  -trackeraddr string
        tracker address
```

We need to provide `seedaddr` and `trackeraddr`, the rest will be left as default:

```
ezs -seedaddr "here goes the IP/name that is available to other machines" -trackeraddr "the IP/name of the tracker"
```

If, for example, Machine1 can be reached by the IP 192.168.0.10, the seedaddr will be 192.168.0.10 and
the trackeraddr can be 192.168.0.10 or localhost because the tracker is on the same machine.

```
ezs -seedaddr 192.168.0.10 -trackeraddr localhost
```

- add files to seeder

`ezl` is used for adding files.

You can add a file:

`ezl add path_to_file`

And then list the available files:

`ezl ls`

This is the whole setup needed on the server machine.

#### Machine2(client)

First, you need to setup the tracker address:

`ez tracker tracker_address`

Now you can list the available files:

`ez ls`

And download one of them:

`ez get file_id`

### A more real setup

When serving multiple big files, it would be best to have the tracker and the seeders on separate machines.

Also there needs to be more than one seeder for each file, otherwise there is no point in using this tool.
