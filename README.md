# Waldo

<img src="assets/waldo.png" width="128px">

Waldo is a simple command line application that allows you to ask questions and search the Internet. It wraps around the DuckDuckGo search engine and the Ollama LLM framework.


# Building

## MacOs

Install `cmake` and `go`:

```
$ brew install cmake go
```


Then generate dependencies:

```
$ go generate ./...
```

Then build the binary:

```
$go build .
```

## Linux

Install `cmake` and `Go` on your operating system.

Then generate dependencies:

```
$ go generate ./...
```

Then build the binary:

```
$go build .
```

# Running

```
$ ./waldo 2> /dev/null
```

# Help

Type `help` on the `waldo` prompt to see the commands.

```
Commands:
  ask         ask waldo
  clear       clear the screen
  exit        exit waldo
  help        display help
  info        information about Waldo
  search      search the Internet
  shell       run shell commands
  switch      switch to a different model.
```