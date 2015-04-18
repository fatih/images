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

## Documentation

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

Modify modifies certain attributes of an image(s). The `modify` command might
expose different usages based on the provider. This is mainly due the
difference between each providers reflected API's. To get the usage for a
provider call it with the `--provider NAME --help` flag:

```
$ images modify --provider aws --help
```

##### AWS

`images` allows to change the tags of AWS images. The following commands are
supposed to be executed with `IMAGES_PROVIDER=aws` or with `--provider aws`:

To create or override a image tag:

```
$images modify --create-tags "Name=ImagesExample" --image-ids ami-f465e69d
```

To delete the tags of an image

```
$images modify --delete-tags "Name=ImagesExample" --image-ids ami-f465e69d
```

The commands also have support for batch action:

```
$images modify --create-tags "Name=Example" --image-ids ami-f465e69d,ami-c5c237ac,ami-64pgca7e
$images modify --delete-tags "Name=Example" --image-ids ami-f465e69d,ami-c5c237ac,ami-64pgca7e
```

#### Copy
```
$ images copy
```

## License

The MIT License (MIT) - see LICENSE.md for more details

