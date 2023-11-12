# Nagios-Plugin exporter configuration

The file is written in [YAML format](http://en.wikipedia.org/wiki/YAML), defined by the scheme described below.
Brackets indicate that a parameter is optional.
For non-list parameters the value is set to the specified default.

Generic placeholders are defined as follows:

* `<boolean>`: a boolean that can take the values `true` or `false`
* `<int>`: a regular integer
* `<duration>`: a duration matching the regular expression `[0-9]+(ms|[smhdwy])`
* `<filename>`: a valid path; either absolute or in the current working directory
* `<string>`: a regular string
* `<template>`: a string processed as [Golang text template][]

The other placeholders are specified separately.

See [example.yml](example.yml) for configuration examples.

```yml

modules:
     [ <string>: <module> ... ]

```


### `<module>`
```yml

  # The plugin executable to run as part of this probe
  command: <filename>

  # How long the probe will wait before giving up.
  [ timeout: <duration> ]

  # Mapping of commandline arguments/flags to their rendering instructions
  # The map key is used as argument key default value should the instructions
  # not contain an explicit definition
  arguments:
    [ <string>: <plugin_argument> ... ]

  # Variables to expose to the argument builder
  variables:
    [ <string>: <string> ... ]

  # Environment variables to expose to the command and argument builder
  environment:
    [ <string>: <string> ... ]

```

*Variables*

Variables are a map of variable names to their value/values. They are exposed to the argument
rendering system and are enriched with values from the probe query URL parameters. The map values
act as fallback. If no value is available, the variable is dropped and will not be available later on.

```yml
variables:
  example_host: ""
  example_port: 8443
  example_payload:
    - EHLO localhost
    - QUIT
  example_expect: ""
```

Using the request `/probe?module=example&example_host=example.com&example_payload=ABRT` as example,
the resulting variables would end up as follows

```yml
Vars:
  example_host: "example.com"
  example_port: 8443
  example_payload: [ "ABRT" ]
```

*Environment*

Environment variables are a map of variable names to their value/values. They are exposed to the argument
rendering system and are enriched with values from the process environment. The map values
act as fallback. If no value is available, the variable is dropped and will not be available later on.

```yml
environment:
  HOSTNAME: ""
  EMAIL: "root@localhost"
  USER: "foo"
  HOME: ""
```

Assuming the exporter was started with two environment variables (`HOSTNAME=gaff` and `USER=bac`),
the resulting environment would end up as follows

```yml
Env:
  HOSTNAME: "gaff"
  EMAIL: "root@localhost"
  USER: "bac"
```

#### `<plugin_argument>`
```yml
  # Must resolve to *true* in order for the argument to end up in the commandline arguments
  [ set_if: <template|boolean> ]

  # Optional value to pass along with the key
  [ value: <template> ... ]

  # Order of the argument in the commandline arguments list
  [ order: <int> ]

  # Flag value for the commandline argument. Passed, as-is to the executable, so
  # and hypens or other flag indicators must be included in this value
  [ key: <string> ]

  # If the value yields multiple values, the key is passed along with each of them,
  # unless this expression resolves to *false*
  [ repeat_key: <template|boolean> ]

  # If this expression resolves to *true* the key is omitted
  [ skip_key: <template|boolean> ]

  # Optional key/value seperator for the commandline argument. If not defined,
  # the option and value are passed to the plugin individually instead of concatenated.
  [ separator: <string> ]
```

The templates in this context have access to both the `argument_variables` (*Vars*)
and `argument_environment` (*Env*), after they have been evaluated against the current probe request.

## Template rendering

Template expressions are rendered using the [Golang text template][] feature. The template scope
includes a variety of functions to help with rendering values.

[Golang text template]: https://pkg.go.dev/text/template

### `lines`

Wrapper around [`strings.Join`](https://pkg.go.dev/strings#Join)
to concatenate values using newlines.

### `net_host`

Wrapper around [`net.SplitHostPort`](https://pkg.go.dev/net#SplitHostPort),
to extract the hostname/address part of the input.
Errors are discarded and result in an empty string as result.

### `net_port`

Wrapper around [`net.SplitHostPort`](https://pkg.go.dev/net#SplitHostPort),
to extract the port number of the input.
Errors are discarded and result in an empty string as result.

### `read_file`

Returns the content of the given filename, or an empty string if the file
does not exist or cound not be read.

### `lower`

See [`strings.ToLower`](https://pkg.go.dev/strings#ToLower)

### `trim`

See [`strings.TrimSpace`](https://pkg.go.dev/strings#TrimSpace)

### `upper`

See [`strings.ToUpper`](https://pkg.go.dev/strings#ToUpper)

### `compact`

See [`sprig.mustCompact`](https://masterminds.github.io/sprig/mustCompact.html)

### `default`

See [`sprig.default`](https://masterminds.github.io/sprig/default.html)

### `first`

See [`sprig.mustFirst`](https://masterminds.github.io/sprig/mustFirst.html)

### `initial`

See [`sprig.mustInitial`](https://masterminds.github.io/sprig/mustInitial.html)

### `join`

See [`sprig.join`](https://masterminds.github.io/sprig/join.html)

### `rest`

See [`sprig.mustLast`](https://masterminds.github.io/sprig/mustLast.html)

### `strval`

See [`sprig.toString`](https://masterminds.github.io/sprig/toString.html)

### `uniq`

See [`sprig.mustUniq`](https://masterminds.github.io/sprig/mustUniq.html)

