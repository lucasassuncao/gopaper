# Gopaper

`gopaper` is a command-line tool (CLI) in Go designed to change the desktop wallpaper. It works as follows:

## Configuration:

The tool depends on a configuration file named gopaper.yaml file that defines the program's behavior, containing log settings and a list of categories. Each category specifies a source directory containing images, the display mode (e.g., crop), and whether it is enabled.

You can use `./gopaper baseconfig` to generate the initial configuration file in the correct format.

<!-- gomarkdoc:embed:start -->

<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# gopaper

```go
import "gopaper"
```

## Index



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)


<!-- gomarkdoc:embed:end -->