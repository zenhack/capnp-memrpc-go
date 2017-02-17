WIP transport for [go-capnproto2][1] using the [memfd][2] package.
Big roadblock right now is that we need a way to tell the rpc package to
use a particular allocator.

Apache 2.0 license.

[1]: https://github.com/zombiezen/go-capnproto2
[2]: https://github.com/justincormack/go-memfd
