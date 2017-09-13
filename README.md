# consul-leader-election
[![GitHub version](https://badge.fury.io/gh/wywy%2Fconsul-leader-election.svg)](https://badge.fury.io/gh/wywy%2Fconsul-leader-election)
[![Build Status](https://travis-ci.org/wywy/consul-leader-election.svg?branch=master)](https://travis-ci.org/wywy/consul-leader-election)


Application to support [consul leader election](https://www.consul.io/docs/guides/leader-election.html).

## Details

Acquires a consul session for a given key.

Exits with `0` if the local node successfully acquired the session or owns the session for the given key.

Exits with `1` if the local node is not able to acquire the session and not owner of the session for the given key.

Exits with `2` on error.

With `-leader-tag` and `-not-leader-tag` you can tag a service based on the leader election outcome.

## Arguments

`-key`

  Name of the key, which will be used to do leader election. All nodes that are participating should agree on a given key to coordinate.

`-key-value`

  Value of the key (`-key`). (Default: consul node name)

`-session-name`

  Name of the session, which will be used to acquire the key (`-key`). (Default: `-key`)

`-health-check`

  Health checks, which will be used for the session. (`serfHealth` is set by default)

`-leader-exit-code`

  Overwrite exit code if leader. (Default: 0)

`-not-leader-exit-code`

  Overwrite exit code if not leader. (Default: 1)

`-error-exit-code`

  Overwrite exit code for errors. (Default: 2)

`-service-name`

  Name of the service you want to tag.

`-leader-tag`

  Tag which will be set to `-service-name` if leader.

`-not-leader-tag`

  Tag which will be set to `-service-name` if not leader.

## License

Copyright 2017 ATVAG GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
