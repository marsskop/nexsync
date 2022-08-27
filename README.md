# NexSync

**NexSync**, a Go tool to synchronize repositories and artifacts to and from Nexus2/Nexus3 Artifact Servers. Able to synchronize Maven2 artifacts and artifact versions between repositories.

Currently capable of:

- synchronizing artifacts between repos in Nexus3
- synchronizing artifact versions from repos in Nexus2/Nexus3 to repos in Nexus3

Tool was originally created for automatic artifact mirroring.

Inspired by [N3DR](https://github.com/030/n3dr).

## Installation

Build from source.

```bash
git clone https://github.com/marsskop/nexsync.git
cd nexsync
go build -o ./ ./..
```

## Usage

Help.

```bash
./nexsync -h
nexsync is a tool to synchronise Maven repositories and artifacts from and to Nexus2/Nexus3

Usage:
  nexsync [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  help         Help about any command
  sync         Synchronize repositories
  syncartifact Synchronize artifact

Flags:
      --config string      Path to config file; default is ~/.nexsync
  -d, --debug              enable debug
      --dir string         Directory to store artifacts in during sync (default "/tmp")
  -h, --help               help for nexsync
      --nexus2from         nexusFrom is Nexus2
      --nexus2to           nexusTo is Nexus2
      --nexusfrom string   Nexus endpoint to sync from (default "http://localhost:8080")
      --nexusto string     Nexus endpoint to sync to (default "http://localhost:8080")
      --passfrom string    Password for Nexus user in nexusFrom
      --passto string      Password for Nexus user in nexusTo
      --repofrom string    Repository to sync from in nexusFrom
      --repoto string      Repository to sync to in nexusTo
      --userfrom string    Nexus user to authenticate with in nexusFrom
      --userto string      Nexus user to authenticate with in nexusTo

Use "nexsync [command] --help" for more information about a command.
```

Sync repos between Nexus3 servers.

```bash
./nexsync sync --urlfrom="<repo URL to sync from>" --repofrom="<repo to sync from>" \
    --userfrom="<user with download permissions>" --passfrom='<password for userFrom>' --tmpdir="./tmp" \
    --urlto="<repo URL to sync to>" --repoto="<repo to sync to>" \
    --userto="<user with upload permissions>" --passto='<password for userTo>'
```

Sync artifact versions fron Nexus2 to Nexus3 server.

```bash
./nexsync syncartifact --urlfrom="<repo URL to sync from>" --repofrom="<repo to sync from>" \
    --userfrom="<user with download permissions>" --passfrom='<password for userFrom>' --tmpdir="./tmp" \
    --urlto="<repo URL to sync to>" --repoto="<repo to sync to>" \
    --userto="<user with upload permissions>" --passto='<password for userTo>' \
    --artifact="<groupID>/<artifactID>" --fromnexus2
```

## Why?

No tools that I could find ([N3DR](https://github.com/030/n3dr), [nexus_cli](https://github.com/RiotGamesMinions/nexus_cli/), [repositorytools](https://github.com/packagemgmt/repositorytools/), [nexus3_cli](https://github.com/thiagofigueiro/nexus3-cli/)) are capable of mirroring artifacts between Nexus Artifact Servers of different versions.
