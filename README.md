# Gamut Mask Tools

## Purpose

Generating gamutmask images from existing image files like this (width/height=250, paddingX/paddingY=2)

![(Van Gogh Starry Night)](https://user-images.githubusercontent.com/8169082/56171963-5c135180-5fb5-11e9-9b77-b50144c41fac.png)

> Van Gogh Starry Night

Gamutmask tool once executed generates gamutmask images for every image found in the `input` folder. By default it runs until stopped thus monitoring for updated or added files.

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

Command line also supports the following parameters:
* `width`
* `height`
* `paddingX`
* `paddingY`

## Full Help

```
  -fresh
        Start fresh by deleting all images from output
  -height int
        Height of the resulting gamut image (default 250)
  -help
        Print this help
  -input string
        Folder name where input files are located (default "./_input")
  -monitor
        Monitor input folder for new and updated files (default true)
  -once
        Shortuct to monitor=false
  -output string
        Folder name where output files should be saved (default "./_output")
  -paddingX int
        Widgth of the resulting gamut image (default 2)
  -paddingY int
        Widgth of the resulting gamut image (default 2)
  -width int
        Widgth of the resulting gamut image (default 250)
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
