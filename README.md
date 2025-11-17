<!-- Caddy Speedtest
     https://github.com/gucci-on-fleek/caddy-speedtest
     SPDX-License-Identifier: Apache-2.0+
     SPDX-FileCopyrightText: 2025 Max Chernoff
-->

Caddy Speedtest
===============

This is a Caddy module that provides a basic HTTP speed test service,
intended to be used via command-line programs like `curl` or `wget`.


Installation
------------

- Using [`xcaddy`](https://github.com/caddyserver/xcaddy):

  ```console
  $ xcaddy build --with github.com/gucci-on-fleek/caddy-speedtest
  ```

- Using `caddy add-package` ([not
  recommended](https://github.com/caddyserver/caddy/issues/7010)):

  ```console
  $ caddy add-package github.com/gucci-on-fleek/caddy-speedtest
  ```

- Manually, by visiting
  [`caddyserver.com/download`](https://caddyserver.com/download) and
  selecting `github.com/gucci-on-fleek/caddy-speedtest` from the list of
  plugins.


Usage
-----

### Server

Add the following to your `Caddyfile`:

```caddyfile
example.com {
	speedtest /speedtest
}
```

### Client

To test download speed:

```console
$ curl --output /dev/null --progress-meter https://example.com/speedtest?bytes=100MB
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 95.3M  100 95.3M    0     0   410M      0 --:--:-- --:--:-- --:--:--  411M
```

To test upload speed:

```console
$ head --bytes=100M /dev/urandom | curl --output /dev/null --progress-meter --form file=@- https://example.com/speedtest
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  100M  100    23  100  100M    149   651M --:--:-- --:--:-- --:--:--  649M
```

Why does this exist?
--------------------

There are a plethora of speed test services available online, but none
of them meet _all_ the following criteria:

- Must be directly usable from `curl`.

- Must support uploading files of arbitrary size.

- Must have a well-defined geographical location (no anycast).

Since I'm already using Caddy for other purposes, the easiest solution
was to write a Caddy module that provides this functionality.

Licence
-------

[Apache License, Version 2.0 or
later](https://www.apache.org/licenses/LICENSE-2.0).

