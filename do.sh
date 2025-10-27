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
    
    go test -v -bench . -benchtime=1000x # -cpuprofile cpu.out -memprofile mem.out
    rm test_*.tur >/dev/null 2>&1
}

gentest() {
    rm -rf testdata || mkdir testdata
    mkdir -p testdata/input/nested/deep

    root="testdata/input"
    dd if=/dev/urandom of=$root/nested/deep/tiny bs=32 count=1
    dd if=/dev/urandom of=$root/nested/large1 bs=7197884 count=1
    dd if=/dev/urandom of=$root/nested/medium1 bs=171572 count=1
    dd if=/dev/urandom of=$root/large0 bs=2397171 count=1
    dd if=/dev/urandom of=$root/medium0 bs=71536 count=1
    touch $root/nano  # can't create with dd
    dd if=/dev/urandom of=$root/small0 bs=100 count=1

    fout="testdata/expected.tur"
    touch testdata/expected.tur
    
    printf '\n||| testdata/input/large0\n' >> "$fout"
    dd if=testdata/input/large0 bs=2397171 count=1 status=none >> "$fout"
    printf '\n' >> "$fout"
    
    printf '\n||| testdata/input/medium0\n' >> "$fout"
    dd if=testdata/input/medium0 bs=71536 count=1 status=none >> "$fout" 
    printf '\n' >> "$fout"
    
    printf '\n||| testdata/input/nano\n' >> "$fout"
    printf '\n' >> "$fout"
    
    printf '\n||| testdata/input/nested/deep/tiny\n' >> "$fout"
    dd if=testdata/input/nested/deep/tiny bs=32 count=1 status=none >> "$fout" 

    printf '\n' >> "$fout"
    printf '\n||| testdata/input/nested/large1\n' >> "$fout"
    dd if=testdata/input/nested/large1 bs=7197884 count=1 status=none >> "$fout"
    printf '\n' >> "$fout"
    
    printf '\n||| testdata/input/nested/medium1\n' >> "$fout"
    dd if=testdata/input/nested/medium1 bs=171572 count=1 status=none >> "$fout"
    printf '\n' >> "$fout"
    
    printf '\n||| testdata/input/small0\n' >> "$fout"
    dd if=testdata/input/small0 bs=100 count=1 status=none >> "$fout"
    printf '\n' >> "$fout"
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
