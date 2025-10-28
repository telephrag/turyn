#!/bin/bash

gentest() {
    rm -rf testdata || mkdir testdata
    mkdir -p testdata/input/nested/deep

    root="testdata/input"

    files=(\
        ".invisible0"\
        "extralarge0"\
        "large0"\
        "medium0"\
        "nano0"\
        "nested/deep/extralarge1"\
        "nested/deep/extralarge2"\
        "nested/deep/normal0"\
        "nested/deep/ordinary0"\
        "nested/deep/tiny1"\
        "nested/large1"\
        "nested/medium1"\
        "nested/nano1"\
        "nested/nano2"\
        "nested/nano3"\
        "nested/ordinary1"\
        "nested/temporary0"\
        "small0")

    sizes=(\
        10243780 20971520 16373913 13308852 0\
        17906448 19438984 7178709 4113648 1048576\
        14841387 11776316 0 0 0 5646183 8711244\
        2581112)


    for i in "${!files[@]}"; do
        if [ "${sizes[i]}" = 0 ]; then
            touch $root/${files[$i]} # not using dd with empty file
            continue
        fi
        dd if=/dev/random of=$root/${files[$i]} bs=${sizes[$i]} count=1
    done

    fout="testdata/expected.tur"
    touch testdata/expected.tur

    for i in "${!files[@]}"; do
        filename=${files[$i]}
        printf "\n||| testdata/input/$filename\n" >> "$fout"
        if [ ! "${sizes[i]}" = 0 ]; then
            dd if=$root/${files[$i]} bs=${sizes[$i]} count=1 >> $fout
        fi
        printf "\n" >> "$fout"
    done
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
