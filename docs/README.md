# gxctl - Device Services CLI

gxctl is a command line interface for running commands against gridX Device
Services API. You can use gxctl to deploy applications, inspect and manage
devices, and run remote maintenance. This overview covers gxctl syntax,
describes the command operations, and provides common examples.

## Getting started

gxctl is installable on a variety of Linux platforms, macOS and Windows.
If you read these instructions, you most likely have received an archive
containing the binary along with a default configuration file.

### Installation

#### Linux

- Extract the binary and run `sudo install -o root -g root -m 0755 gxctl /usr/local/bin/gxctl`

#### MacOS

- Move the archive outside of your Downloads folder, e.g. to your home folder.
- Extract the binary and make it executable by running `chmod +x ./gxctl`
- Create `/usr/local/bin` if it doesn't exist yet: `sudo mkdir -p /usr/local/bin`
- Move the gxctl binary to a file location on your system PATH:
`sudo mv ./gxctl /usr/local/bin/gxctl && sudo chown root: /usr/local/bin/gxctl`
- After trying to run `gxctl` for the first time, you might have to
make an exception in your system settings to allow it to run even though it is
not signed.

#### Windows

- Extract the binary and append or prepend the folder containing the gxctl
binary to your PATH environment variable.

### Configuration

- Navigate to your home directory and create a folder called `.gxctl`.
For Windows, this is the root of your `%USERPROFILE%` directory of
the user that will be running the gxctl command.
- Copy the provided base `config.yaml` file to `~/.gxctl/config.yaml`.
If using the git repo instead of the archive, you can find the file in
`config/gridx`. This file mainly lists the settings for
authenticating gxctl for usage with the different subaccounts. Before you can
start working, you have to use `gxctl login` to retrieve a token by
authenticating with your user account.

```sh
mkdir -p ~/.gxctl
cp config.yaml ~/.gxctl/config.yaml
gxctl login # will open a browser window where you can sign in using your
# gridx.de email
```

The generated token is only valid for a limited amount of time (currently 4 weeks);
if you see errors regarding authentication
after that time, you may have to simply refresh the token by running
`gxctl login` again.

### Remote Maintenance

In order to use the SSH functionality, you'll want to alter your SSH config,
typically located in `~/.ssh/config`.
If running for the first time, you can simply execute
`mkdir -p ~/.ssh && gxctl ssh setup >> ~/.ssh/config`. Otherwise,
please compare the output of `gxctl ssh setup`
with the contents of your ssh config and update the latter accordingly.
You can run `gxctl ssh check` to find out if your local configuration has been
setup correctly. It will error and print a diff if there are issues and succeed otherwise.

*Note*: As the SSH functionality uses agent forwarding, you must have
__ssh-agent__ available.
The most simple solution is running `eval $(ssh-agent)` once per terminal
session - this will work fine for basic SSH usage, but will not allow you to
open more than one connection at a time.

## Profiles / Accounts

The managed devices are split into different server-side accounts. Your available
accounts are defined in `~/.gxctl/config.yaml`. To differentiate between them
when using gxctl commands, you can pass the `--profile` argument. You will have
to issue a `gxctl login` command separately for each profile that you want to use:

```sh
 gxctl --profile profile1 login
 gxctl --profile profile1 get device
```

*Note*: You can set a default profile using `gxctl config use-profile profile1`.

## Syntax

Use the following syntax to run gxctl commands from your terminal window:

`gxctl [command] [TYPE] [NAME] [flags]`

where `command`, `TYPE`, `NAME`, and `flags` are:

- __command__   Specifies the operation that you want to perform on one or more resources,
for example `create`, `get`, `update`
- __TYPE__   Specifies the resource type. Resource types are case-insensitive and
you can specify the singular, plural, or abbreviated forms. For example, the
following commands produce the same output:

```shell
 gxctl get deployment deployment1
 gxctl get deployments deployment1
 gxctl get deploy deployment1
```

- __NAME__   Specifies the name of the resource. Names are case-sensitive.
If the name is omitted, details for all resources are displayed, for example
`gxctl get pods`

When performing an operation on multiple resources, you can specify each resource
by name:

```shell
 gxctl get pod 57e82f8e-08f4-48f9-8e75-28552d09701f 21d7d72a-ceac-437d-bf57-816a43efbaba
```

It is possible to abbreviate uuids which are used as an identifier eg. for pods
or deployments.  Please note that identifiers not of the format of an uuid eg.
in the case of applications need to be specified with it's full name.

```shell
 gxctl get pods 57e 21d
 gxctl get apps testapp testapp2
```

- __flags__   Specifies optional flags. For example, you can use the `-o` or `--output`
flags to specify the output format of your command

## Operations

- __apply__   `gxctl apply [[-f | ----filename]=Filename] [flags]`
- __create__   `gxctl create [[-f | ----filename]=Filename] [flags]`
- __config__   `gxctl config [flags]`
- __delete__   `gxctl delete [TYPE] [NAME] [flags]`
- __diff__   `gxctl apply [[-f | ----filename]=Filename]`
- __get__   `gxctl get [TYPE] [NAME] [[-o | --output]=OUTPUT_FORMAT] [flags]`
- __label__   `gxctl label [TYPE] [NAME] [flags]`
- __lint__   `gxctl lint [[-f | ----filename]=Filename] [flags]`
- __login__   `gxctl login [flags]`
- __logs__ `gxctl logs [SUBCOMMAND] [NAME] [flags]`
- __ssh__   `gxctl ssh [NAME] [flags]`
- __update__   `gxctl update [TYPE] [NAME] [[-f | ----filename]=Filename] [flags]`
- __validate__   `gxctl validate [[-f | ----filename]=Filename] [flags]`
- __version__   `gxctl version [flags]`

## General resource types

- __applications__   Abbreviated alias `application`,`app`.
Applications are used to allow a logical grouping of deployments. Read more about
[Device Selectors](#device-selectors).
- __configmaps__
Abbreviated alias `configmap`,`deviceconfigmaps`,`deviceconfigmap`,`cm`,`dcm`.
Configmaps can be used to inject arbitrary data into a container.
It can e.g. be used to provide the a JSON or YAML config file or add some
custom content. In our [Example Deployment](#howto-example-deployment),
we're using a Configmap to provide the NGINX container with some custom HTML
content to display.
- __deployments__
Abbreviated alias `deployment`,`deploy`. Deployments define a container blueprint.
They're are getting translated into Pods as a concrete instance. Those Pods are
getting started as a container on the corresponding device. Deployments can
either match a single device or a group of devices. Read more about [Device Selectors](#device-selectors).
- __devices__
Abbreviated alias `device`. Devices are the API representation of the physical
gateway. The resource stores general information like the serialnumber or MAC
address and a current state of the device providing real time information about
the device itself as well as information about the network the device is operating
in.
- __pods__
Abbreviated alias `pod`,`po`. Pods are the API representation of a container
running on a specific device. They're getting created by a controller for the
best matching deployment of an application.

## Output options

The default output format for all gxctl commands is the human readable plain-text
format. To output details to your terminal window in a specific format, you can
add either the `-o` or `--output` flags to a supported gxctl command.

```shell
 gxctl [command] [TYPE] [NAME] -o=<output_format>
```

- __-o=json__   Output a JSON formatted API object.
- __-o=wide__   Output in the plain-text format with any additional information.
- __-o=yaml__   Output a YAML formatted API object.

## Howto: Example Deployment

gxctl is designed to be used in a declarative way, thus most commands are expecting
a file to get provided. We generally recommend using the combination of
`diff` and `apply` instead of using `create` and `update`.

The following files are given an easy example of a combination of deployment
and configmap to deploy a NGINX including a specific content on a specific
device called `7b6419fa-7ac5-4320-a87b-8fe8513130dc`.

### configmap.yaml

```yaml
metadata:
  id: 2cde6802-a7f9-4a35-872c-c10c364babec
spec:
  data:
    index.html: |-
      <html>
        <head>
        </head>
        <body>
          <marquee width="50%" direction="left" scrollamount="10">
            <h1 style="color:#0fb9b6">gridX is a very nice company</h1>
          </marquee>
        </body>
      </html>
```

### deployment.yaml

```yaml
metadata:
  id: 5c4dc7b2-c684-41c5-9e83-b2cc87b5c3cb
spec:
  app: nginx
  selector:
    matchByDeviceID: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
  template:
    spec:
      containers:
      - image: nginx:1.21.5-alpine
        name: nginx-demo
        ports:
        - containerPort: 80
          hostPort: 80
          name: http
        volumeMounts:
        - mountPath: /usr/share/nginx/html
          name: content
      volumes:
      - configMap:
          name: 2cde6802-a7f9-4a35-872c-c10c364babec
        name: content
```

## Device Selectors

We're using the concept of a device selector to make sure the right pods are
running on the right device. There are two different types of selector's which
can be added to a deployment, namely `matchByDeviceID` and `matchByLabels`.
While `matchByDeviceID` is used as a 1:1 relation to allocate a deployment to
a specific device using it's UUID, `matchByLabels` can be used to target a set
of devices based on their labels. Those selector's are always working in the
scope of an `application`.

For `matchByLabels`, the following applies to allocate a deployment to a set of devices:

- The following rules are always evaluated per `application`.
- `matchByDeviceID` has precedence over `matchByLabels`
- All labels defined in a `matchByLabels` selector must be attached to the
targeted devices in order to match.
- If there are multiple deployments matching a device for the same `application`,
the one will be chosen, which matches most specifically, meaning having the
highest number of matching labels.
- If there are multiple deployments matching a device for the same `application`,
and also those deployments having an equal number of matching labels, we use
the most recent deployment based on its creation timestamp.

__A concrete example:__

Following we got a set of devices and deployments.

### Device01

```yaml
id: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
labels:
  core.gridx.ai/status: running
  gridx.de/channel: stable
  gridx.de/region: eu-central
```

### Device02

```yaml
id: de6dfaf0-bebd-4338-b4ce-bd451406b39a
labels:
  core.gridx.ai/status: running
  gridx.de/channel: stable
  gridx.de/region: eu-central
```

### Device03

```yaml
id: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
labels:
  core.gridx.ai/status: running
  gridx.de/channel: stable
  gridx.de/region: eu-west
```

### Deployment1

```yaml
app: nginx
selector:
  matchByDeviceID: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
```

### Deployment2

```yaml
app: nginx
selector:
  matchByLabels:
    gridx.de/channel: stable
```

### Deployment3

```yaml
app: nginx
selector:
  matchByLabels:
    gridx.de/channel: stable
    gridx.de/region: eu-west
```

We end up with the following allocation:

- `Device01` runs `Deployment1`. Even though `Deployment2` would also match,
based on it's values, it runs on `Deployment1` as `matchByDeviceID` always has
precedence over `matchByLabels`. Remember: Since both `Deployment1` and
`Deployment2` are of the same application only one will be running.
- `Device02` runs `Deployment2` as the labels are matching.
- `Device03` runs `Deployment3`. Even though, both `Deployment2` and
`Deployment3` are matching based on the labels, `Deployment3` is more specific
(2 vs. 1 matching labels). Remember: Since both `Deployment1` and `Deployment2`
are of the same application only one will be running.

## Howto: Common operations

`gxctl create` - Create a new resource

```shell
# Create a new deployment
$ gxctl create -f new_deployment.json
# Create a new app
$ gxctl create app testapp
```

`gxctl config` - Deal with profile configuration

```shell
# Set default profile
$ gxctl config use-profile some_customer_profile
# Get default profile
$ gxctl config current-profile
# See token status per profile
$ gxctl config status
# List all profiles
$ gxctl config list-profiles
```

`gxctl update` - Updates a existing resource

```shell
# Update a device using a updatefile
$ gxctl update -f update_device.json
# Update a deployment using a updatefile
$ gxctl update -f update_deployment.json
```

`gxctl delete` - Delete a existing resource

```shell
# Delete an app
$ gxctl delete app testapp
# Delete an Deployment using the uuid abbreviation
$ gxctl delete deploy c78
# Delete two Deployments using both uuid abbreviation and full qualified name
$ gxctl delete deploy c78 35e3dede-2b45-4212-82fb-b92f7d391e05
```

`gxctl get` - List one or more resources

```shell
# Get a List of all devices
$ gxctl get devices
# Get a List of all devices and include additional information (such as labels).
$ gxctl get devices -o wide
# get a List of all devices that match a serialnumber
$ gxctl get devices --serial D294-200-000-000-581-P-X
# get a List of all devices that match a serialnumber (wildcard)
$ gxctl get devices --serial 581-P-X
# get a List of all devices that have a label with key gridx.de/channel
$ gxctl get devices --label gridx.de/channel
# get a List of all devices that have a label with key gridx.de/channel and value
# alpha
$ gxctl get devices --label gridx.de/channel=alpha
# get a List of all devices that match all key/value label pairs
$ gxctl get devices --label gridx.de/channel=alpha,gridx.de/datadog=true
# Get information of a single device
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f
# Get information of a single device in json format
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f -o json
# Get information of two devices
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f 8st62f8e-22gd-ab45-ll23-115980970ab
# Get a List of all pods
$ gxctl get pods
# Get a List of all pods on a certain device
$ gxctl get pods -d 57e82f8e-08f4-48f9-8e75-28552d09701f
# Get a List of all deployments
$ gxctl get deploy
# Get information of a single device, showing its deployments
$ gxctl get deploy --device-id 57e82f8e-08f4-48f9-8e75-28552d09701f
# Get information of a single device, showing its deployments by using the serialnumber
$ gxctl get deploy -S D294-200-000-000-581-P-X
# Get information of monitoring deployment by using the serialnumber
$ gxctl get deploy -S D294-200-000-000-581-P-X -a monitoring -o yaml
# Get a information of a deployment include additional information using uuid abbreviation
$ gxctl get deploy c78 -o wide
```

`gxctl logs` - manage logs settings on the devices. Allows to control
how long and on which level the device is going to save logs.

```shell
# Get help for the command. The -h flag is also applicable to any of subcommands
# of logs.
$ gxctl logs -h
# Get logs settings for all the devices
$ gxctl logs list
# Get logs settings for a specific device
$ gxctl logs get <device_id>
# Get logs settings for a specific device in a yaml format
$ gxctl logs get <device_id> -o yaml
# Get logs for a specific device while referring to a device by its serial number
$ gxctl logs get <device_serial_number> -S
# Enable logs for a specific device for 1 hour on the INFO level
# (also works with -S, as well as all following commands if you want to give a
# serial number instead of a device id)
$ gxctl logs enable <device_id> -e 1h -l info
# Change logs settings for a specific device: make them expire in 2 hours
# instead of 1 hour and change the logs level to WARNING. 
# If there are logs settings creted by someone else before you,
# provide a -c flag to become a new owner of these settings. Otherwise,
# you will get an error.
$ gxctl logs update <device_serial_number> -S -e 2h -l warning 
# Disable logs for a specific device
$ gxctl logs disable <device_id>
```

`gxctl label` - Labels different resources

```shell
# Add a new label to a device
$ gxctl label device 57e82f8e-08f4-48f9-8e75-28552d09701f test=test
# Add multiple new labels to a device using uuid abbreviation
$ gxctl label device 57e test=test demo=demo
# Remove a label "test" from a device
$ gxctl label device 57e82f8e-08f4-48f9-8e75-28552d09701f test-
# Remove a label "test" from a device and a new one
$ gxctl label device 57e82f8e-08f4-48f9-8e75-28552d09701f test- demo=demo
```

`gxctl apply` - Creates or Updates resources

```shell
# Creates a resource if not existing, otherwise updates.
$ gxctl apply -f deployment.json
```

`gxctl ssh` - Show SSH config and setup tunnel

```shell
# Pint SSH config which needs to be setup
$ gxctl ssh setup
# Check if local SSH config has been setup correctly
$ gxctl ssh check
# all commands below need up-to-date SSH client config according to $(gxctl ssh setup)!
# SSH remote terminal session to device
$ ssh D244-200-000-000-445-P-X.gridbox
# SSH remote terminal session to device (wildcard match)
$ ssh 445-P-X.gridbox
# copy local file to device
$ scp /tmp/foo.txt D244-200-000-000-445-P-X.gridbox:/tmp
# copy remote file from device
$ scp D244-200-000-000-445-P-X.gridbox:/tmp/foo.txt /tmp
# forward port 8080 of devices 192.168.169.198 and 192.168.169.199 in the remote
# network to local ports 8080 and 8081
$ ssh -L 8080:192.168.169.198:8080 -L 8081:192.168.169.199:8080 D294-200-000-000-581-P-X.gridbox
# open up port 2210 on the gridbox in the remote network and forward incoming
# traffic to local port 2210
$ ssh -R 2210:localhost:2210 D294-200-000-000-581-P-X.gridbox

`gxctl diff` - Diff a resource file

```shell
# Shows the diff between the live system and the provided resource file
$ gxctl diff -f deployment.json
```

`gxctl validate` - Validates a resource file

```shell
# Validates if required fields are set and that there are no unsupported fields
$ gxctl validate -f deployment.json
```

`gxctl lint` - Lint a resource file

```shell
# Lint for best practices and common errors
$ gxctl lint -f deployment.json
```

## Howto: Special operations

```shell
# Get the public key of a device
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f --show-publickey
# Get the public IP of a device
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f --show-public-ip
```

## Howto: Sorted output

You can sort the output of the get command using JSONPath expressions.

```shell
# Get a list of all pods sorted by their starttime
$ gxctl get pods -s .status.startTime
# Get a list of all devices sorted by their serialnumber in wide output
$ gxctl get devices -s .spec.serialnumber -o wide
# Get a list of all deployments sorted by their app
$ gxctl get deploy -s .spec.app
```
