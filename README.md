# hac-cli

home access center command line interface

## usage
- clone the repo and build the executable: `go build` or just run `go install github.com/tildezero/hac@main` to install it to `$GOPATH/bin`
- run the executable: `hac`
- setup credentials: `hac setup`
- view overview of classes (ids and averages): `hac overview`
- view all grades for a class: `hac class [id]` where id is from `hac overview`
