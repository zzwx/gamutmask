# Gamut Mask Tools

## To Compile and Install From Source

To resolve dependency tree, `dep` should be installed:

```
$ brew install dep
$ brew upgrade dep
```

"vendor" folder will be populated with necessary dependencies by running
the following from the project's root:

```
$ dep ensure
```

## ``godep``-enabling Log

* https://github.com/golang/dep
* https://github.com/golang/dep/blob/master/docs/FAQ.md

Initialization from project's root directory:

```
$ dep init
```

Useful commands:

```
$ dep status
$ dep ensure -update
```
