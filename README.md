# Waldo

<img src="assets/waldo.png" width="128px">

Waldo is a simple command line application that allows you to ask questions and search the Internet. It wraps around the [DuckDuckGo](https://duckduckgo.com) search engine, the [Ollama](https://github.com/jmorganca/ollama) LLM framework, OpenAI's GPT models as well as Google's Gemini model.


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
$ go build .
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
$ ./run
```

# Help

Type `help` on the `waldo` prompt to see the commands.

```
Commands:
  add         add a new model to Waldo
  ask         ask waldo
  clear       clear the screen
  exit        exit waldo
  help        display help
  info        information about Waldo
  search      search the Internet
  shell       run shell commands
  switch      switch to a different model
```

# Features

## Add a model

You can add a local model into Waldo and use it. Waldo's local models are based off Ollama's so you can add any of the Ollama models available here https://ollama.ai/library.

```
waldo> add
model name? phi:chat
verifying sha256 digest
writing manifest
removing any unused layers
success
(1 second 979 milliseconds)
waldo>
```

(The output below shows ~2 seconds to add phi:chat because it's already added)

## Switch to a different model

You can switch to any of the local models that you have downloaded and added to Waldo. If you want to use OpenAI or Google Gemini, please get an API key first, then add it into the `.env` file.

## Info

Provides information about Waldo.

## Shell

Allows you to step into a shell-like environment within Waldo.

```
waldo> shell
shell> ls -al
executing> ls -al
total 53144
drwxr-xr-x   18 sausheong  staff       576 Dec 30 15:08 .
drwxr-xr-x  129 sausheong  staff      4128 Dec 28 23:01 ..
-rw-r--r--    1 sausheong  staff       166 Dec 30 15:02 .env
-rw-r--r--    1 sausheong  staff       126 Dec 30 15:02 .env.example
drwxr-xr-x   13 sausheong  staff       416 Dec 30 15:09 .git
-rw-r--r--    1 sausheong  staff        30 Dec 30 15:02 .gitignore
-rw-r--r--    1 sausheong  staff        79 Dec 29 00:28 .gitmodules
-rw-r--r--    1 sausheong  staff      1116 Dec 30 15:04 README.md
drwxr-xr-x    3 sausheong  staff        96 Dec 29 22:27 assets
-rw-r--r--    1 sausheong  staff      3926 Dec 30 14:48 go.mod
-rw-r--r--    1 sausheong  staff     31271 Dec 30 14:48 go.sum
-rw-r--r--    1 sausheong  staff        43 Dec 30 14:23 llama.log
-rw-r--r--    1 sausheong  staff      6653 Dec 30 14:54 main.go
drwxr-xr-x   28 sausheong  staff       896 Dec 29 11:11 ollama
-rwxr-xr-x    1 sausheong  staff        20 Dec 30 15:08 run
-rw-r--r--    1 sausheong  staff      2528 Dec 30 09:54 types.go
-rwxr-xr-x    1 sausheong  staff  27120498 Dec 30 14:56 waldo
-rw-r--r--    1 sausheong  staff      5711 Dec 30 14:56 waldo.go

```

## Ask

Allows you to ask the current model any questions.

```
waldo> ask
ask> Why is the sky blue?
The sky appears blue because of a phenomenon called Rayleigh scattering. When sunlight enters Earth's atmosphere, it encounters tiny molecules of gases such as nitrogen and oxygen. These molecules absorb and scatter the light in all directions, but they scatter shorter (blue) wavelengths more than longer (red) wavelengths. This is known as Rayleigh scattering.

As a result of this scattering, the blue light is dispersed throughout the atmosphere, giving the sky its blue appearance. The blue color we see is actually the combination of light from the sun that has been scattered in all directions by the tiny molecules of gases in the atmosphere.

So, to summarize, the sky appears blue because of the way light interacts with the tiny molecules of gases in Earth's atmosphere through Rayleigh scattering.

(5 seconds 279 milliseconds)
```

## Search

Allows you to ask for answers through the Internet (using DuckDuckGo).

```
waldo> search
search> COVID-19 cases in Singapore
The query "COVID-19 cases in Singapore" has resulted in several relevant search results that provide up-to-date information on the current situation of COVID-19 in Singapore. Here are some key points from each result:

* According to the Ministry of Health (MOH) website, the estimated number of COVID-19 cases in Singapore rose to 56,043 in the week of December 3 to 9, compared to 32,035 cases in the previous week. The average daily COVID-19 hospitalizations also increased to 350 from 225 in the previous week, while the average daily Intensive Care Unit (ICU) cases rose to nine cases from four cases in the previous week. (Source: MOH | COVID-19 Statistics - Ministry of Health)
* Singapore has seen a significant increase in COVID-19 cases in recent weeks, with the number of new cases admitted to hospitals jumping to 965 in the past week, up from 763 the previous week. (Source: New weekly COVID-19 cases admitted to hospitals and ICUs highest for ... - CNA)
* The Ministry of Health (MOH) has also confirmed that there have been rumors about a large increase in severe COVID-19 cases and deaths due to the XBB strain, but these are not true. (Source: Latest COVID-19 News and Data - CNA)
* To prepare for another potential COVID-19 wave, Singapore has taken various measures such as increasing the capacity of isolation facilities and enhancing contact tracing efforts. (Source: Is Singapore ready for another COVID-19 wave? - CNA)

Overall, these search results suggest that COVID-19 cases in Singapore have been on the rise in recent weeks, with a significant increase in hospitalizations and ICU cases. The Ministry of Health has also taken measures to prepare for another potential COVID-19 wave.

List of URLs returned:

* [MOH | COVID-19 Statistics - Ministry of Health](https://www.moh.gov.sg/covid-19/statistics)
* [Singapore COVID - Coronavirus Statistics - Worldometer](https://www.worldometers.info/coronavirus/country/singapore/)
* [MOH | News Highlights - Ministry of Health](https://www.moh.gov.sg/news-highlights/details/update-on-local-covid-19-situation-and-measures-to-protect-healthcare-capacity)
* [Latest COVID-19 News and Data - CNA](https://www.channelnewsasia.com/coronavirus-covid-19)
* [Is Singapore ready for another COVID-19 wave? - CNA](https://www.channelnewsasia.com/singapore/is-singapore-ready-for-another-covid-19-wave-1435176)
* [New weekly COVID-19 cases admitted to hospitals and ICUs highest for ... - CNA](https://www.channelnewsasia.com/singapore/new-weekly-covid-19-cases-admitted-to-hospitals-and-icu-highest-for-2023-4005086)

(28 seconds 222 milliseconds)
```