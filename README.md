# Introduction
`regrep` an implementation of command-line tool similar to `grep` in Golang using trigram and a basic state machine

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/p-id/regrep/master/LICENSE)


# Basics
The current implementation of `regrep` is based on automata theory and guarantees that searches
complete in linear time with respect to the size of the input and in a fixed amount of stack space. 
`regrep` utilizes Trigram-accelerated regex search. This essentially means that each byte is processed
once during index build-up and later during the actual regular expression build - hence the name `regrep`

A trigram is a substring of 3 consecutive characters, or runes.

For example - The trigrams of trigram are: tri, rig, igr, gra and ram.

Valid regular expressions are able to be converted to a trigram query against the index. 
`regrep` searches the trigram index for possible matches and then runs the full regular
expression against the possibles. Each regular expressions must match at least one trigram.

It’s a good idea to think carefully about regular expression queries. The more possible 
trigrams matched by a query, the more results are checked against the full regular 
expression  and the less efficient the query. The reverse, however, is also true. 
The more trigrams enforced by the regular expression, the better use `regrep` can make
of the index and the smaller the possible results set.

# Building
The current distribution requires 'golang' and 'docker' for running the build/test setup.
The common options to try out are 'make' or 'make tests'

The default settings use Alpine Linux image [Busybox + musl-libc] for smaller footprint.

# Reference

 - *Trigrams and Regexp*. [Google RE2](https://github.com/google/re2)

 - From *'The Kubernetes Authors'* [Golang Build Template](https://github.com/thockin/go-build-template)

# About

This project was created by [Piyush Dewnani](mailto:piyush@dewnani.net)
