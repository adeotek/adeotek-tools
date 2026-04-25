#!/bin/bash

if [ "$1" == "-i" ]; then
  npm ci
fi

npm run dev
