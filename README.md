# Gamut Mask Tools

## Requirement

* go 1.12

## To Compile and Install From Source

To resolve dependency tree, simply `go build`.

```
$ go mod verify
$ go build
```

In order to run you will need an output folder. By default the ``_output`` folder is used which you'll need to create if you want to use the default one:

```
$ mkdir _output
```

You can certainly modify output folder by passing ``--output`` parameter:

```
$ gamutmask.exe --output ./output
```

Once executed the tool monitors the `./_input` folder for images. That can also be changed by passing a value after `--intput`` parameter:

```
$ gamutmask.exe --input ./input
```

