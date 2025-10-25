#!/bin/sh

bench() {
    mkdir testdata/ >/dev/null 2>&1 || echo "testdata/ already exists, skipping..." 
    
    if [ -z "$(ls -A testdata/)" ]; then
        echo "testdata/ is empty, fetching test data..."
        git clone -q https://github.com/telephrag/bscrap testdata/
        if  [ $? -ne 0 ]; then 
            echo "failed to fetch test data from git, aborting..." && exit 1
        else 
            echo "successfully fetched data"
        fi
    else
        cd testdata/
        hash="$(git rev-list --parents HEAD | tail -1)"
        expected="19ccbd15ebd1c67f1ebe105477503e8c99afc6d0" 
        if [[ "$hash" != "$expected" ]]; then
            echo "testdata/ has invalid contents, aborting..."
            exit 1
        fi
        cd ..
    fi
    
    go test -bench=. -count=1000
    rm test.tur >/dev/null 2>&1
}

#install() {
#    go build -o /usr/bin/turyn main.go
#}
#
#clean() {
#    if [ -f "/usr/bin/turyn" ]; then
#        rm /usr/bin/turyn
#    fi
#    rm -rf testdata/ >/dev/null 2>&1 || echo ""
#}

$@
