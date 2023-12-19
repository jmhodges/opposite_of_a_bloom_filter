# The Opposite of a Bloom filter

A Bloom filter is a data structure that may report it contains an item that it
does not (a false positive), but is guaranteed to report correctly if it
contains the item ("no false negatives"). The opposite of a Bloom filter is a
data structure that may report a false negative, but can never report a false
positive. That is, it may claim that it has not seen an item when it has, but
will never claim to have seen an item it has not.

More words about this datastructure can be found in my blog post, [The Opposite
of a Bloom
Filter](https://www.somethingsimilar.com/2012/05/21/the-opposite-of-a-bloom-filter/).

This repository contains thread-safe implementations of "the opposite of a Bloom
filter" in Java and Go.

The Java implementation uses maven and may be packaged up with the usual `mvn`
commands.

The Go implementation may be built or installed with the `go` tool:

    go get github.com/jmhodges/opposite_of_a_bloom_filter/go/oppobloom
