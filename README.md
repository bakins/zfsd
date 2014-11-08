zfsd
====

Simple HTTP interface for ZFS.

Needs a new name as the name "zfsd" is already in us by at least on other project.

After struggling with working with ZFS in various languages, a few of us said "it would be nice if there was an HTTP interface that worked across all supported platform."  That was a few months ago and we never did anythign about it. So, I decided to start this in the hope others had the same itch.

## Current Plan ##
* REST-ish API for common ZFS tasks, including snapshots and clones.
* endpoints for gathering metrics
* Just listen on a unix socket for now and use file permissions.
* Write it in Go.  I like being able to deploy a single binary.  I don't have strong feelings about this, however.

## Future ##
* Add real AAA
* 
