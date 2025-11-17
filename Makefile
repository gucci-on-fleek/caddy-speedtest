# Caddy Speedtest
# https://maxchernoff.ca/tools/speedtest
# SPDX-License-Identifier: Apache-2.0+
# SPDX-FileCopyrightText: 2025 Max Chernoff

################
### Settings ###
################

# Remove the builtin rules
.SUFFIXES:
MAKEFLAGS += --no-builtin-rules

# Silence the commands
.SILENT:

# Shell settings
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
SHELL := /usr/bin/bash

# Default target
.DEFAULT_GOAL := default
.PHONY: default
default:
	$(error Please specify a target.)

# Build Caddy with the speedtest module
.PHONY: build
build: caddy

caddy: go.mod speedtest.go
	xcaddy build --with maxchernoff.ca/tools/speedtest=.

# Run the Caddy server with the speedtest module
.PHONY: run
run: caddy Caddyfile
	./caddy run --config Caddyfile
