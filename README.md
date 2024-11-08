# Homelab CLI (homelab)

[![Build](https://github.com/tuxgal/homelab/actions/workflows/build.yml/badge.svg)](https://github.com/tuxgal/homelab/actions/workflows/build.yml) [![Tests](https://github.com/tuxgal/homelab/actions/workflows/tests.yml/badge.svg)](https://github.com/tuxgal/homelab/actions/workflows/tests.yml) [![Lint](https://github.com/tuxgal/homelab/actions/workflows/lint.yml/badge.svg)](https://github.com/tuxgal/homelab/actions/workflows/lint.yml) [![CodeQL](https://github.com/tuxgal/homelab/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/tuxgal/homelab/actions/workflows/codeql-analysis.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/tuxgal/homelab)](https://goreportcard.com/report/github.com/tuxgal/homelab)

A CLI for managing both the configuration and deployment of groups of
docker containers on a given host.

The configuration is managed using a yaml file. The configuration
specifies the container groups and individual containers, their
properties and how to deploy them.

Even though there are tools like `docker compose` to manage deployment
of a group of containers, there are still lots of gaps especially
when you would like to configure certain flags during the deployment
dynamically.

`homelab` is suited for a mid to large scale setup running at east 20
containers, where there are challenges like managing the IP address
space, networks, etc.

This CLI is in the early stages of development.
