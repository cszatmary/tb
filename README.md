# `tb`

tb is a CLI for running TouchBistro services on a development machine.

It is aimed at making local development easy in a complicated microservices architecture by provisioning your machine with the dependencies you need and making it easy for you to run them on your machine in an environment that is close to how they run in production.

### **Table of Contents**
- [Requirements](#requirements)
    + [Installed Software](#installed-software)
    + [AWS ECR](#aws-ecr)
- [Installation](#installation)
- [Quickstart](#quickstart)
- [Commands](#commands)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [Having trouble?](#having-trouble?)
- [Gotchas / Tips](#Gotchas-/-Tips)

## Requirements

### Installed Software

Right now, the only requirement is that you have the xcode cli tools and `nvm`.

This project will install and manage all other dependencies that you need.

### AWS ECR

You will need access to AWS ECR (Amazon's docker registry) to pull artifacts instead of having `tb` build them on the host.

Once you have been provided access to our AWS account by DevOps Support, create a personal access key and make note of your secret key.

Configure your AWS CLI credentials by running `aws configure` (use `us-east-1` for region).

## Installation

`tb` is available through TouchBistro's `homebrew` tap. If you do not have homebrew, you can install it by going to [brew.sh](https://brew.sh)

First add Touchbistro's tap to get access to all the available tools:

```sh
brew tap touchbistro/tap git@github.com:TouchBistro/homebrew-tap.git
```

In order to install `tb` you will need to create a GitHub Access Token. Follow the instructions [here](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to learn more. When creating it select `repo` for the permissions. Also make sure you enable SSO for the token.

Once you have your token add the following to your `.bash_profile` or `.zshrc`:
```sh
export HOMEBREW_GITHUB_API_TOKEN=YOUR_TOKEN
```

Now you can install `tb` with `brew`:
```sh
brew install tb
```

## Quickstart

TBD: Add instructions for getting binary from homebrew when we set that app

Run `tb up -s postgres` to setup your system and start a `postgresql` service running in a docker container. Try running `tb --help` or `tb up --help` to see what else you can do.

## Commands

`tb` comes with a lot of convenient commands. See the documentation [here](docs/tb.md) for the command documentation.

## Configuration

`tb` can be configured to either build images/containers locally, or to pull existing images from ECR. This is all set in `config.yml` with the `ecr` and `imageURI` flags.

## Contributing

See [contributing](CONTRIBUTING.md) for instructions on how to contribute to `tb`.

## Having trouble?

Check the [FAQ](docs/FAQ.md) for common problems and solutions. (Pull requests welcome!)

## Gotchas / Tips

- Do not run npm run or npm run commands from the host unless you absolutely need to.

- **Previous Setup**: If you already have `postgres.app` or are running postgres with homebrew or any other way, Datagrip-like tools will be confused about which pg to connect to. You won't need these anymore, so you can just delete them. Use `pgrep postgres` and make sure you don't have any other instances running.

- **SQL EDITORS**: To use external db tools like datagrip or `psql`, keep `CORE_DB_HOST` in the .env file as it is, but use `localhost` as the hostname in datagrip (or tool of choice). see `bin/db` for an example that uses `pgcli` on the host. Inside the docker network, containers uses the service names in `docker-compose.yml` as their hostname. Externally, their hostname is just `localhost`.

- **Slowness**: If running things in Docker on a mac is slow, allocate more CPUs, Memory and Swap space using the Docker For Mac advanced preferences. Keep in mind that some tools (like `jest`) have threading issues on linux and are not going to be faster with more cores. Use `docker stats` to see resource usage by image.
