#!/usr/bin/env bash

# try to adjust dependences
export APPROOT=$(pwd)

if [[ -n "$ZSH_VERSION" ]];
then
    readopts="rA"
else
    readopts="ra"
fi

# adjust GOPATH
while IFS=':' read -$readopts ARR; do
    for i in "${ARR[@]}"; do
        case ":$GOPATH:" in
            *":$i:"*) :;;
            *) GOPATH=$GOPATH:$i;;
        esac
    done
done <<< "$APPROOT"
export GOPATH=$GOPATH

# adjust PATH
while IFS=':' read -$readopts ARR; do
    for i in "${ARR[@]}"; do
        case ":$PATH:" in
            *":$i/bin:"*) :;;
            *) PATH=$i/bin:$PATH;;
        esac
    done
done <<< "$GOPATH"
export PATH


# mock development && test envs
if [[ ! -d "$(pwd)/src/github.com/dolab/goplay" ]];
then
    mkdir -p "$(pwd)/src/github.com/dolab"
    ln -s "$(pwd)/goplay" "$(pwd)/src/github.com/dolab/goplay"
fi
