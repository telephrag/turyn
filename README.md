# Turyn

Turyn is a CLI utility for combining multiple files into one. Writes files' contents into output in parallel using `OffsetWriter` so, I presume it should be pretty fast but no comparisons were done. You can generate testdata and feed it to your favourite program of similar purpose. Probably will not work on OS's other then Linux because of use of `fallocate` and maybe other things I can't test right now. Props if you know the language the name is from.

# Installing

`go install github.com/telephrag/turyn`

# Testing

Generate testdata:
`chmod +x do.sh`

Run tests as usual:
`go test`

Run benchmark:
`go test -bench=main`
