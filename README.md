# Gamut Mask Tools

## Purpose

Gamutmask once executed currently runs as a daemon, looking for new and modified images in one specidifed `input` folder and outputs generated gamutmask images to the `output` folder.

One can certainly simply use `internal/lib/lib.go` containing the core function itself inside his own project:

```
func GenerateGamutMask(img image.Image, maskWidth, maskHeight int) (wheel *image.RGBA64)
```

The resulted images look like this:

![gamut example](https://user-images.githubusercontent.com/8169082/55894480-2b33b680-5b88-11e9-9709-432a05dd416f.jpg)

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

Once executed the tool monitors the `./_input` folder for images. That can also be changed by passing a value after `--intput` parameter:

```
$ gamutmask.exe --input ./input
```

