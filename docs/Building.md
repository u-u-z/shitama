
# Building

At the moment, Shitama should be able to build with Go 1.6.x and later.

## Holder, Shard and Client

```bash
mkdir -p ~/go/src/github.com/evshiron/
cd ~/go/src/github.com/evshiron/
git clone https://github.com/evshiron/shitama
cd shitama
bash ./scripts/build.sh
# ./build/holder/holder or ./builder/shard/shard
```

## Client UI

Currently the only UI `client-ui-qt` is Qt-based. As the target platform is Windows, cross-compiling on dockerized Fedora was used, but the generated binaries wouldn't work, so at the moment the compiling is automatically done by AppVeyor workers.

Read `/.appveyor.yml` and `/scripts/build_client_dist.bat` for a general understanding how the distributables are built.
