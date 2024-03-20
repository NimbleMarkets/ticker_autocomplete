# ticker_autocomplete

*This project is under active development and is not yet ready for production use.  There is scant documentation and the interfaces may change.*

`ticker_autocomplete` is a Golang-based library for auto-completion of financial symbols ("tickers"). It is part of the [Nimble.Markets](https://nimble.markets) platform.


## Examples

You can build the example programs with [Taskfile](https://taskfile.dev).  For a list of tasks, run the `list` task:

```bash
$ task list 
task: [list] task --list-all
task: Available tasks for this project:
* build:                       Build all
* build-cassette:              Build vhs cassette gifs
* clean:                       Clean
* default:                     Default task is "build"
* go-tidy:                     Tidy all
* go-update:                   Update Go dependencies
* list:                        Lists available tasks
* publish-cassette:            Publish vhs cassette gifs to Charm
* you-tickercomplete-me:       Build you-tickercomplete-me
```

### `you-tickercomplete-me`

`you-tickercomplete-me` is a simple BubbleTea program illustrating usage of the library.  *It is still a work in progress towards a full-fledged widget.*

<img alt="you-tickercomplete-me example"  width="600" src="https://vhs.charm.sh/vhs-3vg2ydrukT2ohePaffWfHf.gif" />

```bash
$ task
task: [go-tidy] go mod tidy
task: [you-tickercomplete-me] go build -o bin/you-tickercomplete-me cmd/you-tickercomplete-me/*.go

$ ./bin/you-tickercomplete-me
```


## Feedback

Contributions are welcome!

Please submit a PR or open an issue to discuss features.


## License

All specifications and data are owned by their respective organizations and are subject to change.  This project is not affiliated with these organization.

Released under the [MIT License](https://en.wikipedia.org/wiki/MIT_License), see [LICENSE.txt](./LICENSE.txt).

Copyright (c) 2024 [Neomantra BV](https://www.neomantra.com).  All rights reserved. 

----
Made with :heart: and :fire: by the team behind [Nimble.Markets](https://nimble.markets).  Stay Nimble!
