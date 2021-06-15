# Changelog

### 0.30.0 (Unreleased)

**Image:**

- `108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl:v0.30.0-linux-amd64`

_New Features:_

_Changes:_

_Documentation:_

### 0.29.2

**Image:**

- `108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl:v0.29.2-linux-amd64`

_New Features:_

_Changes:_

- [X] [#144](https://github.com/grid-x/gxctl/pull/144) Fix lint command after rework

_Documentation:_

### 0.29.1

**Image:**

- `108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl:v0.29.1-linux-amd64`

_New Features:_

_Changes:_

- [X] [#143](https://github.com/grid-x/gxctl/pull/143) Make token check work in CI
- [X] [#143](https://github.com/grid-x/gxctl/pull/143) Change version schema

_Documentation:_

# Changelog old version schema

### 0.1.29 (equivalent with 0.29.0 in current version schema)

**Image:**

- `108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl:v0.1.29-linux-amd64`

_New Features:_

- [X] [#140](https://github.com/grid-x/gxctl/pull/140) Add `gxctl prune` command for usage in CI which will delete resources which got removed from the repo.
- [X] [#139](https://github.com/grid-x/gxctl/pull/139) Improve token handling to print out expiry error in all commands
- [X] [#135](https://github.com/grid-x/gxctl/pull/135) Make is easier to find devices in all available accounts by running `gxctl get device --profile='*'`. This command can be combined with eg. `--serial` Note: get devices will be the only command working on multiple accounts

_Changes:_

- [X] [#138](https://github.com/grid-x/gxctl/pull/138) Fix `gxctl update` command which did not properly pick up changes
- [X] [#136](https://github.com/grid-x/gxctl/pull/136) Show callback error of `gxctl login` on HTML result page
- [X] [#135](https://github.com/grid-x/gxctl/pull/135) Rework get commands. **This is a BREAKING CHANGE!**
  Removed `gxctl get deployment XYZ --show-devices`, `gxctl get devices XYZ --show-deployments` and `gxctl get devices XYZ --show-pods`
  in favor of `gxctl get deployment --device-id`, `gxctl get pod --device-id` and `gxctl get devices --deployment-id`

_Documentation:_


### 0.1.28 (equivalent with 0.28.0 in current version schema)

**Image:**

- `108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl:v0.1.28-linux-amd64`

_New Features:_

- [X] [#132](https://github.com/grid-x/gxctl/pull/132) Implement `gxctl get pods --device-id`
- [X] [#130](https://github.com/grid-x/gxctl/pull/130) Check SSH config

_Changes:_

- [X] [#133](https://github.com/grid-x/gxctl/pull/133) Fix for SSH config check in OpenSSH 8.5 and above
- [X] [#131](https://github.com/grid-x/gxctl/pull/131) Also look for gxctl config in XDG_CONFIG_HOME
- [X] [#129](https://github.com/grid-x/gxctl/pull/129) Remove cleanup config 

_Documentation:_

