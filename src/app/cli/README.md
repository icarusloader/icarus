# ![](https://wsrv.nl?url=raw.githubusercontent.com/icarusloader/.github/416d97ada95d88f771b4542b41769d8241d99bad/media/icarus.svg&height=32) icarusloader config.kdl & meta.kdl

config.kdl has the following format:

```kdl
enabled {
    interal.foo
    interal.bar // bar gets applied last and takes priority
}
flags {
    example {
        key value
    }
}
```

icarusloader defines a simple API for a mod's meta.kdl:

```kdl
domain "interal.example"
version "1.2.3"
modules basic // the modules it uses, order doen't matter
```

Each mod's specific module data is inside a folder with the same name as the module. For example:

```
example
 └ basic
    └ content
      ...
 └ meta.kdl
```

For more info on how a specific modules flags or mods folder works, refer to their README.md.
