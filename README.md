# Images

Images is a tool to manage machine images from multiple providers with a single
CLI interface. Think of it's as a companion to the popular image creation tool
[Packer](https://packer.io/). I'm developing it on my spare times, so feedback
and contributions are welcome!

## Features

- Multiple provider backend support: `AWS`, `DO`, etc...
- Multi region support
- List images from a provider
- Modify image attributes, such as tags or names
- Delete images
- Copy images from one region to another region
- Commands are executed concurrently (delete, list, etc..).
- Flexible configuration system. Read from file, environment variables or command line flags.

## Installation

Download pre combiled binaries:


Or if you have Go installed, just do (note that it fetches not vendored
dependencies, so you might end with a different binary.):

```bash
go get github.com/fatih/images
```

## Intro

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
pass the "--provider name --help" flags at any time, where `name` is the
provider name, such as "aws".

## Configuration

`images` is a very flexible CLI tool. It can parse the necessary configuration from
either a file, from environment variables or command line flags. Examples:

```bash
$ images list --provider aws --region "us-east-1,eu-west-2" --access-key "..." -secret-key "..."
```

or via environment variable:

```bash
$ IMAGES_PROVIDER=aws IMAGES_AWS_REGION="us-east-1" IMAGES_AWS_ACCESS_KEY=".." images list
```

or via `.imagesrc` file, which can be either in `TOML` or `JSON`. Below is an example for `TOML`:

```toml
provider = "aws"

[aws]
region     = "us-east-1,eu-west-2"
access_key = "..."
secret_key = "..."
```
and execute simply:

```bash
$ images list
```

## Examples

`images` has multi provider support. The following examples are for the
provider "aws".  The commands are supposed to be executed with
`IMAGES_PROVIDER=aws` or with `--provider aws` or added to `.imagesrc` file via
`provider = "aws"`


#### List

List images for a given provider. Examples:

```
$ images list -region "us-east-1"
```

List from all regions (fetches concurrently):

```
$ images list -region "all"
```

#### Delete

Delete images from the given provider. Examples:

```
$ images delete -image-ids "ami-1ec4d766,ami-c3h207b4,ami-26f1d9r37"
```

Note that you don't need to specify a region if you define multiple ids.
`images` is automatically matching the correct region and deletes it. Plus they
all are deleted concurrently.

#### Modify

`images` allows to change the tags of AWS images for the provider "aws".

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

## License

The BSD 3-Clause License - see `LICENSE` for more details
