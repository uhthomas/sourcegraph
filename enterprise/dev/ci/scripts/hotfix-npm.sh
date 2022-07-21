#!/usr/bin/env bash

set -e

sed -i 's/https:\/\/registry.npmjs.org/http:\/\/npm-verdaccio:4873/' yarn.lock
