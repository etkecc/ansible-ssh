# ansible-ssh

<!-- vim-markdown-toc GitLab -->

* [What?](#what)
* [Why?](#why)
* [How?](#how)
* [Where to get?](#where-to-get)
    * [Binaries and distro-specific packages](#binaries-and-distro-specific-packages)
    * [Build yourself](#build-yourself)

<!-- vim-markdown-toc -->

## What?

ansible-ssh is a wrapper around the standard `ssh` client (usually, openssh client) that will try to read ansible inventory within the current directory first and connect to the matched server. If none matched - it will fallback to the standard ssh

## Why?

Because `ansible-console` is not interactive and there are plenty of occasions when you need an interactive shell over the internet and you need it **now**.

## How?

1. Copy the `config.yml.sample` into your `$XDG_CONFIG_HOME` (usually, `~/.config`) and rename it to `ansible-ssh.yml`
2. Run `ansible-ssh` in a dir with an ansible inventory

You can even add an alias to use `ansible-ssh` as wrapper around the standard ssh command:

```bash
# $HOME/.bashrc
alias ssh="ansible-ssh"
```

## Where to get?

### Binaries and distro-specific packages

[Releases page](https://gitlab.com/etke.cc/int/ansible-ssh/-/releases)

### Build yourself

`just build` or `go build .`
