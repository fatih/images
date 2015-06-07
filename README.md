# Images [![Build Status](http://img.shields.io/travis/fatih/images.svg?style=flat-square)](https://travis-ci.org/fatih/images) [![Release](https://img.shields.io/github/release/fatih/images.svg?style=flat-square)](https://github.com/fatih/images/releases)

Images is a tool to manage machine images from multiple providers over a single
CLI interface. Its fast(concurrent actions), simple to use and and very
flexible. Think of it as a companion to the popular image creation tool
[Packer](https://packer.io/). You can fetch images from multiple providers,
delete them, change tags or names of multiple images, any many other things.

![images](http://d.pr/i/1lR5f+)
![images](https://github.com/github/fatih/images/blob/master/asset/dia.png)

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

Download the latest release suitable for your system:

=> [Images releases](https://github.com/fatih/images/releases)

Or if you have Go installed, install latest development version:

```bash
go get -u github.com/fatih/images
```

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
$ images list --providers aws --region "us-east-1,eu-west-2" --access-key "..." -secret-key "..."
```

or via environment variable:

```bash
$ IMAGES_PROVIDER=aws IMAGES_AWS_REGION="us-east-1" IMAGES_AWS_ACCESS_KEY=".." images list
```

or via `.imagesrc` file, which can be either in `TOML` or `JSON`. Below is an example for `TOML`:

```toml
providers = "aws"

[aws]
region     = "us-east-1,eu-west-2"
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
`IMAGES_PROVIDER=aws` or with `--providers aws` or added to `.imagesrc` file via
`providers = "aws"`


#### List

List images for a given provider. Examples:

```
$ images list -region "us-east-1"
```

List from all regions (fetches concurrently):

```
$ images list -region "all"
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
$images modify --create-tags "Name=ImagesExample" --ids ami-f465e69d
```

To delete the tags of an image

```
$images modify --delete-tags "Name=ImagesExample" --ids ami-f465e69d
```

The commands also have support for batch action:

```
$images modify --create-tags "Name=Example" --ids ami-f465e69d,ami-c5c237ac,ami-64pgca7e
$images modify --delete-tags "Name=Example" --ids ami-f465e69d,ami-c5c237ac,ami-64pgca7e
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

To build `images` just run:

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
