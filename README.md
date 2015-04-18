# Images

Images is a tool to manage machine images from multiple providers with a single
CLI interface. Think of it's as a companion to the popular image creation tool
[Packer](https://packer.io/).

## Features

- Multiple provider support: `AWS`
- List images from a provider
- Modify image attributes, such as tags or names
- Delete images
- Copy images from one region to another region
- Flexible configuration system. Read from file, environment variables or command line flags.

## Installation

Download pre combiled binaries:


Or if you have Go installed, just do:

```bash
go get github.com/fatih/images
```

## Usage

`images` is a very flexible CLI tool. It can parse the necessary arguments from
either a configuration file, from environment variables or command line flags.
To list all commands just run `images`:

```bash
$ images
usage: images [--version] [--help] <command> [<args>]

Available commands are:
    copy      Copy images to different region
    delete    Delete images
    list      List available images
    modify    Modify image properties
```

To show help message for a specific subcommand, attach the `--help` flag:

```
$ images modify --help
```

Because `images` is built around to support multiple providers, each provider
has a specific set of features. To display the specific provider help message
pass the provider name too:

```
$ images modify --provider aws --help

	or via environment variable:


$ IMAGES_PROVIDER=aws images modify --help
```

#### List

```
$ images list
```

#### Delete
```
$ images delete
```

#### Modify
```
$ images modify
```

#### Copy
```
$ images copy
```

## License

The MIT License (MIT) - see LICENSE.md for more details

