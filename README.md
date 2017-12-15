# Gamut Mask Tools

## Using Glide (as if using npm)

* https://glide.sh/
* http://glide.readthedocs.io/en/latest/vendor/

Glide should be installed first.

Resolve The Dependency Tree, install them into "vendor":
```
$ glide update
$ glide install
```

Adding dependency (for ex. ``$ go get -u github.com/kataras/iris/iris``):
```
$ glide get github.com/kataras/iris/iris
```

