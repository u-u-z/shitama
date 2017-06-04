
# Building

Tested on Fedora 25, with Go 1.7.5 and Qt 5.7.1.

```
# docker run -ti fedora bash
$ dnf install git golang mingw32-qt5-qmake
$ mkdir -p ~/go/src/github.com/evshiron/
$ cd ~/go/src/github.com/evshiron/
$ git clone https://github.com/evshiron/shitama.git
$ cd shitama
$ bash scripts/build.sh
$ bash scripts/dist.sh
// $ bash scripts/release.sh
```
