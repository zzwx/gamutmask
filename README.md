# Gamut Mask Tools

## Purpose

Generating gamutmask images from existing image files like this:

![gamut example](https://user-images.githubusercontent.com/8169082/56142824-54c85580-5f6d-11e9-9efb-3ba2007bd253.png)

Gamutmask tool once executed generates gamutmask images for every image found in the `input` folder. By default it runs until stopped.

To run only once:
```
$ gamutmask -once
```

To delete all output images before processing:
```
$ gamutmask -fresh
```

To modify output folder by passing ``-output`` parameter:

```
$ gamutmask -output ./output
```

The tool monitors the `./_input` folder for images unless requested to do it `once`. The folder can also be changed by passing a value for `-intput` parameter:

```
$ gamutmask -input ./input
```

For help issue
```
$ gamutmask help
```


## Usage as a Library

One can certainly simply use `internal/lib/lib.go` containing the core function itself inside his own project:

```
func GenerateGamutMask(img image.Image, maskWidth, maskHeight int) (wheel *image.RGBA64)
```


## Requirement

* go 1.12

## To Build and Install From Source

To resolve dependency tree, simply:

```
$ go mod verify
$ go build
```

To install to golang path bin:

```
$ go install
```
