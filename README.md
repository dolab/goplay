# Goplay

Goplay is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads `Playfile`, a YAML configuration file, which defines networks (groups of hosts), commands and targets. It's inspired from [Sup](https://github.com/pressly/sup).

[![Build Status](https://travis-ci.org/dolab/goplay.svg?branch=master&style=flat)](https://travis-ci.org/dolab/goplay) [![Coverage](http://gocover.io/_badge/github.com/dolab/goplay?0)](http://gocover.io/github.com/dolab/goplay) [![GoDoc](https://godoc.org/github.com/dolab/goplay?status.svg)](http://godoc.org/github.com/dolab/goplay)

## Install

    $ go get -u github.com/dolab/goplay

## Usage

    $ goplay [global options] command [command options] [arguments...]

### Golbal Options

| Option            | Description                      |
|-------------------|----------------------------------|
| `--playfile FILE` | Custom path to Playfile          |
| `--keyfile FILE`  | Custom path to ssh PUB key file  |
| `--prompt`        | Enable outputs mode              |
| `--debug`         | Enable debug/verbose mode        |
| `--help`, `-h`    | Show help/usage                  |
| `--version`, `-v` | Print version                    |


## Playfile

### Network

A group of hosts.

```yaml
# Playfile

networks:
    production:
        all:
            - app1.example.com
            - app2.example.com
            - app3.example.com
            - db1.example.com
            - db2.example.com
        app:
            - app1.example.com
            - app2.example.com
            - app3.example.com
        mysql:
            - db1.example.com
            - db2.example.com
    staging:
        # fetch dynamic list of hosts
        inventory: curl http://example.com/latest/meta-data/hostname
```

`$ goplay production.all COMMAND` will run COMMAND on `app1`, `app2`, `app3`, `db1` and `db2` hosts in parallel.

### Command

A shell command(s) to be run remotely.

```yaml
# Playfile

commands:
    restart:
        desc: Restart APP Container
        run: sudo docker restart app

    restart-mysql:
        desc: Restart MySQL Container
        run: sudo docker restart mysql

    tail-logs:
        desc: Watch Docker Logs
        run: sudo docker logs --tail=20 -f mysql
```

`$ goplay production.mysql restart-mysql` will restart all MySQL docker containers on `db1` and `db2` hosts in parallel.

`$ goplay production.app tail-logs` will tail logs from all APP docker containers on `app1`, `app2` and `app3` hosts in parallel.

### Serial command (a.k.a. Rolling Update)

`serial: N` constraints a command to be run on `N` hosts at a time at maximum. Rolling Update for free!

```yaml
# Playfile

commands:
    release:
        desc: Release APP
        run: sudo docker pull app:latest
        serial: 2
```

`$ goplay production.app release` will pull `app:latest` image from all APP docker containers on `app1`, `app2` and `app3` hosts in parallel. Two at a time at maximum.

### Once command (one host only)

`once: true` constraints a command to be run only on one host. Useful for one-time books.

```yaml
# Playfile

commands:
    build:
        desc: Build Docker Image
        run: sudo docker build -t app:latest .
        once: true # one host only
```

`$ goplay production.app build` will build docker image on one production APP host only.

### Local command

`locally: true` constraints a command to be run locally. Useful for development books.

```yaml
# Playfile

commands:
    frontend-build:
        desc: Package Frontend
        run: npm run build
        locally: true
```

`$ goplay frontend-build` will run `npm` command on localhost only.

### Upload command

Uploads files/directories to all remote hosts. Uses `tar` under the hood.

```yaml
# Playfile

commands:
    upload:
        frontend:
            desc: Upload Frontend
            src: ./dist
            dst: /home/deploy/website
```

`$ goplay production.app upload` will upload frontend build files to `app1`, `app2` and `app3` hosts in parallel.

### Interactive Bash

Do you want to interact over all of hosts at once? Sure!

```yaml
# Playfile

commands:
    bash:
        desc: Interactive Bash
        stdin: true
        run: bash
```

```bash
$ goplay production.app bash
#
# type in commands and see output from all hosts!
# >$ sudo nginx -t && sudo nginx -s reload
# >$ ...
# ^C
```

Passing commands to all hosts:

```bash
$ echo 'sudo nginx -s reload' | goplay production.app bash

# or
$ goplay production.app bash <<< 'sudo nginx -s reload'

# or
$ cat <<EOF | goplay production.app bash
sudo nginx -s reload
uname -a
EOF
```

## Book

Book is an alias for multiple commands. Each command will be run on all hosts in parallel,
`goplay` will check returned status from all hosts, and run subsequent commands on success only
(thus any error on any host will interrupt the process).

```yaml
# Playfile

books:
    deploy:
        - build
        - release
        - restart
```

`$ goplay production.app deploy` is equivalent to `$ goplay production.app build release restart`

# Playfile

## Basic

```yaml
# Playfile

---
version: 1.0.0

# Global environment variables
env:
  NAME: goplay
  IMAGE: example/ci

networks:
  local:
    all:
      - localhost
  staging:
    all:
      - dev1.example.com
  production:
    all:
      - app1.example.com
      - app2.example.com
      - db1.example.com
      - db2.example.com
    app:
      - app1.example.com
      - app2.example.com
    db:
      - db1.example.com
      - db2.example.com

commands:
  echo:
    desc: Print some env vars
    run: echo $NAME $IMAGE $PLAY_NETWORK
  date:
    desc: Print OS name and current date/time
    run: uname -a; date

books:
  all:
    - echo
    - date
```

### Default environment variables available

- `$PLAY_NETWORK` - Current network.
- `$PLAY_HOST` - Current host.
- `$PLAY_USER` - User who invoked sup command.
- `$PLAY_TIME` - Date/time of sup command invocation.
- `$PLAY_ENV` - Environment variables provided on goplay command invocation. You can pass `$PLAY_ENV` to another `goplay` or `docker` commands in your Playfile.

# Common SSH Problem

if for some reason sup doesn't connect and you get the following error,

```bash
connecting to clients failed: connecting to remote host failed: Connect("user@xxx.xxx.xxx.xxx"): ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain
```

it means that your `ssh-agent` dosen't have access to your public and private keys. in order to fix this issue, follow the below instructions:

- run the following command and make sure you have a key register with `ssh-agent`

```bash
ssh-add -l
```

if you see something like `The agent has no identities.` it means that you need to manually add your key to `ssh-agent`.
in order to do that, run the following command

```bash
ssh-add ~/.ssh/id_rsa
```

you should now be able to use sup with your ssh key.


# Development

    fork it, hack it..

    $ make build

    create new Pull Request

We'll be happy to review & accept all Pull Requests!
