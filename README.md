# Images [![Build Status](http://img.shields.io/travis/fatih/images.svg?style=flat-square)](https://travis-ci.org/fatih/images)

Images is a tool to manage machine images from multiple providers over a single
CLI interface. Its fast(concurrent actions), simple to use and very
flexible. Think of it as a companion to the popular image creation tool
[Packer](https://packer.io/). You can fetch images from multiple providers,
delete them, change tags or names of multiple images, any many other things.

[![images](https://github.com/fatih/images/blob/master/asset/dia.png)](https://github.com/fatih/images)

## Features

- Multiple provider backend support: `AWS`, `DigitalOcean`, `GCE`, etc...
- Multi region support
- List images from multiple regions/providers
- Modify image attributes, such as tags, names, states
- Delete images
- Copy images from one region to another region
- Commands are executed concurrently (delete, list, etc..).
- Flexible configuration system. Read from file, environment variables or
  command line flags.

## Installation

Images is still under development. Any feedback/contribution is welcome!
Download the latest release suitable for your system:

[Images releases](https://github.com/fatih/images/releases)

(Want to build & develop the source? Check out the
[Build&Develop](https://github.com/fatih/images#build--development) section!)

## Intro

To list all commands just run `images`:

```bash
$ images
usage: images [--version] [--help] <command> [<args>]

Available commands are:
    copy        Copy images to regions
    delete      Delete available images
    list        List available images
    modify      Modify image properties
    version     Prints the Images version

...
```

Because `images` is built around to support multiple providers, each provider
has a specific set of features. To display the specific provider help message
pass the `-providers name -help` flags at any time, where `name` is the provider
name, such as "aws".

Current supported providers are:

* `aws`
* `do`
* `gce`

Coming soon:

* `virtualbox`
* `docker`

## Configuration

`images` is a very flexible CLI tool. It can parse the necessary configuration from
either a file, from environment variables or command line flags. Examples:

```bash
$ images list --providers aws --regions "us-east-1,eu-west-2" --access-key "..." -secret-key "..."
```

or via environment variable:

```bash
$ IMAGES_PROVIDERS=aws IMAGES_AWS_REGIONS="us-east-1" IMAGES_AWS_ACCESS_KEY=".." images list
```

or via `.imagesrc` file, which can be either in `TOML` or `JSON`. Below is an example for `TOML`:

```toml
providers = ["aws"]
no_color  = true

[aws]
regions    = ["us-east-1","eu-west-2"]
access_key = "..."
secret_key = "..."
```
and execute simply:

```bash
$ images list
```

## Usage

`images` has multi provider support. The following examples are for the
provider "aws".  The commands are supposed to be executed with
`IMAGES_PROVIDERS=aws` or with `--providers aws` or added to `.imagesrc` file
via `providers = "aws"`


#### List

List images for a given region. Examples:

```
$ images list -regions "us-east-1"
```

List from all regions (fetches concurrently):

```
$ images list -regions "all"
```

List from multiple providers (fetches concurrently):

```
$ images list -providers "aws,do"
```

List from all supported providers

```
$ images list -providers "all"
```

Change output mode to json

```
$ images list -output json
```

#### Delete

Delete images from the given provider. Examples:

```
$ images delete -ids "ami-1ec4d766,ami-c3h207b4,ami-26f1d9r37"
```

Note that you don't need to specify a region if you define multiple ids.
`images` is automatically matching the correct region and deletes it. Plus they
all are deleted concurrently.

#### Modify

`images` allows to change the tags of AWS images for the provider "aws".

To create or override a image tag:

```
$ images modify --create-tags "Name=ImagesExample" --ids ami-f465e69d
```

To delete the tags of an image

```
$ images modify --delete-tags "Name=ImagesExample" --ids ami-f465e69d
```

The commands also have support for batch action:

```
$ images modify --create-tags "Name=Example" --ids ami-f465e69d,ami-c5c237ac,ami-64pgca7e
$ images modify --delete-tags "Name=Example" --ids ami-f465e69d,ami-c5c237ac,ami-64pgca7e
```

Just like for the `delete` command, all you need to give is the ami ids.
`images` will automatically match the region for the given id. You don't need
to define any region information.


#### Copy

Copy supports copying an AMI to the same or different regions. Below is a simple example:

```
$ images copy -image "ami-530ay345" -to "us-east-1"
```

Copy supports concurrent copying to multiple regions.:

```
$ images copy -image "ami-530ay345" -to "us-east-1,ap-southeast-1,eu-central-1"
```

Description can be given too (optional):

```
$ images copy -image "ami-530ay345" -to "us-east-1"  -desc "My new AMI"
```

## Build & Development

To build `images` just run ([gb](http://getgb.io) needs to be available on the
system):

```sh
$ make build
```

This will put the `images` binary in the bin folder. Development builds doesn't
have a version, so if called with "--version" it'll output `dev`:

```sh
$ .bin/images --version
dev
```

For creating `release` binaries run ([goxc](https://github.com/laher/goxc) required):

```sh
IMAGES_VERSION="0.1.0" make release
```

## License

The BSD 3-Clause License - see `LICENSE` for more details
