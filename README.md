zfsd
====

Simple HTTP interface for ZFS.

Needs a new name as the name "zfsd" is already in us by at least one other project.

After struggling with working with ZFS in various languages, a few of us said "it would be nice if there was an HTTP interface that worked across all supported platform."  That was a few months ago and we never did anything about it. So, I decided to start this in the hope others had the same itch.

## Current Plan ##
* HTTP API for common ZFS tasks, including snapshots and clones.
* endpoints for gathering metrics
* Just listen on a unix socket for now and use file permissions.
* Write it in Go.  I like being able to deploy a single binary.  I don't have strong feelings about this, however.

## Future ##
* Add real AAA
* 


## RPC ##

This branch includes some initial playing using JSON-RPC.  Why
JSON-RPC?  I used this on a project and it was very easy thanks to the
[Gorilla package](http://www.gorillatoolkit.org/pkg/rpc/json).  Also,
ZFS names are "filepath like" so doing them in a traditional REST way
is a little interesting.  Imagine a path like `POST
/my/zfs/snapshot/has/a/long/name@mysnapshot`.  You could do this by
passing query strings or form values, but then you may as well parse
json, and the Gorilla libraries do that very well.  It is easy enough
to add a REST veneer to the calls here, so I'll continue
experiementing using JSON-RPC.  Perhaps we can just do both?


### Examples ###

If you just use the provisioning done in the Vagrantfile:


#### List ####

```
curl --compressed -H "Content-Type: application/json" \
-X POST -sv http://localhost:9373/_zfs_ \
--data-binary '{"id": 1, "params": [], "method": "ZFS.List" }'
```

Output snippet:

```json

  "result": [
    {
      "name": "testing",
      "used": 568040,
      "available": 8389991704,
      "mountpoint": "/testing",
      "compression": "lz4",
      "type": "filesystem",
      "written": 50716
    },
    {
      "name": "testing/A",
      "used": 45808,
      "available": 8389991704,
      "mountpoint": "/testing/A",
      "compression": "lz4",
      "type": "filesystem"
    },
```

You can also "filter" results to a specific prefix or type.  To just get snapshots in `testing/A`, do:

```
curl --compressed -H "Content-Type: application/json" \
-X POST -sv http://localhost:9373/_zfs_ \
--data-binary '{"id": 1, "params": [{"type":"snapshot", "prefix": "testing/A"} ],"method": "ZFS.List" }'
```

#### Create Snapshot ####

```
curl --compressed -H "Content-Type: application/json" \
-X POST -sv http://localhost:9373/_zfs_ \
--data-binary '{"id": 1, "params":[{"name": "testing/A", "snapshot": "987654321"}], "method":"ZFS.Snapshot" }'
```

Output:

```json
{
  "result": {
    "name": "testing/A@987654321",
    "type": "snapshot"
  },
  "error": null,
  "id": 1
}
```

#### Clone Snapshot ####

```
curl --compressed -H "Content-Type: application/json" \
-X POST -sv http://localhost:9373/_zfs_ \
--data-binary '{"id": 1, "params":[{"name": "testing/A", "snapshot": "987654321", "target": "testing/baz"}],"method": "ZFS.Clone" }'
```

Output:

```json
{
  "result": {
    "name": "testing/baz",
    "used": 1636,
    "available": 8389955712,
    "mountpoint": "/testing/baz",
    "compression": "lz4",
    "type": "filesystem",
    "written": 1636,
    "origin": "testing/A@987654321"
  },
  "error": null,
  "id": 1
  }
  ```
  
