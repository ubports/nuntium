# nuntium

`nuntium` is a golang written component that is supposed to be a drop in
replacement for `mmsd`, interacting between `ofono`'s push client and
`telepathy-ofono`.

This is not a full `mmsd` implementation and only covers most of the feature
requirements that satisfy an Ubuntu Phone (aka Ubuntu Touch).

## Documentation

* [Architecture](docs/architecture.md)
* [Testing](docs/testing.md)

Addtional information:

* [mmsd documentaion](https://kernel.googlesource.com/pub/scm/network/ofono/mmsd/+/master/doc/)
* This project is a continuation of [the ubuntu-phonedations nuntium project](https://github.com/ubuntu-phonedations/nuntium/). 

Crossbuilder:

```bash
crossbuilder inst-foreign dh-golang \
    golang-1.6-doc \
    golang-1.6 \
    golang-go-dbus-dev \
    golang-go-flags-dev \
    golang-go-xdg-dev \
    golang-gocheck-dev\
    golang-udm-dev
```
