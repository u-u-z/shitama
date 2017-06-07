
# Linking

Shitama provides UDP port forwarding service. The primarily supported game is Hisouten, but games host with a single UDP port should benefit too.

## Hisouten Watching

When a watcher connects to the host while having battles, the host will redirect the watcher to the guest, and the guest will provide the needed data for watching. If the watcher can't connect to the guest, e.g. the guest is behind a NAT, the watching will fail.

Shitama uses only a single port to communicate between shards and clients, and replay the data in loopback ports. This design is to support port-restricted-cone and symmetric NATs, but the problems are:

* The guest address told by the host to the watcher will be a loopback address, and the watcher won't be able to connect to it.
* Even if we modify that address to a shard, if the guest is behind a port-restricted-cone or symmetric NAT, the shard has to use a single port per guest, and the guest won't be able to distinguish the data with the original `th123` client.
* If the guest isn't behind a NAT, some extra work will be done to track peers of peers, which might not be a resonable solution.

We can modify `configex123.ini:clientport` to `10800` on guests, and serve that unique port as a host. I'm considering allowing 2 links per client, one for hosting and one for watching, which will require the host to start another `th123` client and watch himself.

## TCP Port Forwarding

Should be as easy as copy-and-paste, but who will use it, and will that consume more traffic for me?
