registry-compare
===

Compares a Docker V1 registry to a V2 registry and lets you know what's missing from the V2 registry.

Why?
---
When migrating a very large registry from V1 to V2, it's useful to know what images are missing and
still need to be migrated.

It would also be fairly simple to extend this tool to allow comparing two different V2 registries.
This could help when migrating from one backing store to another.

Usage
--
```
$ go build .
$ ./registry-compare <V1 URI> <V2 URI>
```

Limitations
---
We only take into account the first 99,999 images and first 99,999 tags per image in the V2 registry.
This is because I was too lazy to implement proper pagination support.
