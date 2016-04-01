# Citar - Trigram HMM part-of-speech tagger

## Introduction

Citar is a  part-of-speech tagger, based on a trigram Hidden Markov Model
(HMM). It (partly) implements the ideas set forth in [1]. It can be used
as a set of stand-alone programs and or from Go.

## C++ Citar

The C++ version of Citar can be found is still
[available](https://github.com/danieldk/citar-cxx). Active maintenance
will proceed in Go. The choice to port Citar from C++ to Go was not made
lightly. However, I believe that switching to Go will ease maintenance,
lower the barrier for contributions, and improve cross-platform
support. Moreover, recent version of Go have made it easier to call Go
code from C.

[1] TnT - a statistical part-of-speech tagger, Thorsten Brants, 2000
