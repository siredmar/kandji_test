# gxctl - Device Services CLI
gxctl is a command line interface for running commands against gridX Device Services API. You can use gxctl to deploy applications, inspect and manage devices, and run remote maintenance. This overview covers gxctl syntax, describes the command operations, and provides common examples.

## Getting started
gxctl is installable on a variety of Linux platforms, macOS and Windows.

### Installation

#### Linux
- Download the binary and run `sudo install -o root -g root -m 0755 gxctl /usr/local/bin/gxctl`

#### MacOS
- Download the binary and make it executable by running `chmod +x ./gxctl`
- Move the gxctl binary to a file location on your system PATH. `sudo mv ./gxctl /usr/local/bin/gxctl && sudo chown root: /usr/local/bin/gxctl`

#### Windows
- Append or prepend the folder containig the gxctl binary to your PATH environment variable.

### Configuration

- Navigate to your home directory and create a folder called `.gxctl`.
- Copy the provided base config file in to `~/.gxctl/config.yaml`. This file mainly lists the settings for
authenticating gxctl for usage with the different subaccounts. Before you can start working, you have to use `gxctl login` to retrieve a token by
authenticating with your user account.

```
$ mkdir -p ~/.gxctl
$ cp config.yaml ~/.gxctl/config.yaml
$ gxctl login # will open a browser window where you can sign in using your gridx.de email
```

The generated token is only valid for a limited amount of time (currently 4 weeks); if you see errors regarding authentication
after that time, you may have to simply refresh the token by running `gxctl login` again.

### Remote Maintenance

In order to use the SSH functionality, you'll want to alter your SSH config, typically located in `~/.ssh/config`.  
If running for the first time, you can simply execute `gxctl ssh setup >> ~/.ssh/config`. Otherwise, please compare the output of `gxctl ssh setup`
with the contents of your ssh config and update the latter accordingly.  
*Note*: As the SSH functionality uses agent forwarding, you must have __ssh-agent__ available.  

## Profiles / Accounts

The managed devices are split into different server-side accounts. Your available accounts are defined in `~/.gxctl/config.yaml`. To differentiate between them when using gxctl commands, you can pass the `--profile` argument. You will have to issue a `gxctl login` command separately for each profile that you want to use:

```
$ gxctl --profile profile1 login
$ gxctl --profile profile1 get device
```

*Note*: You can set a default profile using `Default: true` in your config file atx `~/.gxctl/config.yaml`.

## Syntax

Use the following syntax to run gxctl commands from your terminal window:

`gxctl [command] [TYPE] [NAME] [flags]`

where `command`, `TYPE`, `NAME`, and `flags` are:

* **command**   Specifies the operation that you want to perform on one or more resources, for example `create`, `get`, `update`
* **TYPE**   Specifies the resource type. Resource types are case-insensitive and you can specify the singular, plural, or abbreviated forms. For example, the following commands produce the same output:

```shell
$ gxctl get deployment deployment1
$ gxctl get deployments deployment1
$ gxctl get deploy deployment1
```
* **NAME**   Specifies the name of the resource. Names are case-sensitive. If the name is omitted, details for all resources are displayed, for example `gxctl get pods`

When performing an operation on multiple resources, you can specify each resource by name:

```shell
$ gxctl get pod 57e82f8e-08f4-48f9-8e75-28552d09701f 21d7d72a-ceac-437d-bf57-816a43efbaba
```

It is possible to abbreviate uuids which are used as an identifier eg. for pods or deployments. Please note that identifiers not of the format of an uuid eg. in the case of applications need to be specified with it's full name.

```shell
$ gxctl get pods 57e 21d
$ gxctl get apps testapp testapp2
```

* **flags**   Specifies optional flags. For example, you can use the -o or --output flags to specify the output format of your command

## Operations

* **apply**   `gxctl apply [[-f | ----filename]=Filename] [flags]`
* **create**   `gxctl create [[-f | ----filename]=Filename] [flags]`
* **delete**   `gxctl delete [TYPE] [NAME] [flags]`
* **diff**   `gxctl apply [[-f | ----filename]=Filename]`
* **get**   `gxctl get [TYPE] [NAME] [[-o | --output]=OUTPUT_FORMAT] [flags]`
* **label**   `gxctl label [TYPE] [NAME] [flags]`
* **lint**   `gxctl lint [[-f | ----filename]=Filename] [flags]`
* **login**   `gxctl login [flags]`
* **ssh**   `gxctl ssh [NAME] [flags]`
* **update**   `gxctl update [TYPE] [NAME] [[-f | ----filename]=Filename] [flags]`
* **validate**   `gxctl validate [[-f | ----filename]=Filename] [flags]`
* **version**   `gxctl version [flags]`


## General resource types
* **applications**   Abbreviated alias `application`,`app`
* **deployments**   Abbreviated alias `deployment`,`deploy`
* **devices**   Abbreviated alias `device`
* **configmaps**   Abbreviated alias `configmap`,`deviceconfigmaps`,`deviceconfigmap`,`cm`,`dcm`
* **pods**   Abbreviated alias `pod`,`po`


## Output options

The default output format for all gxctl commands is the human readable plain-text format. To output details to your terminal window in a specific format, you can add either the -o or --output flags to a supported gxctl command.

```shell
$ gxctl [command] [TYPE] [NAME] -o=<output_format>
```

* **-o=json**   Output a JSON formatted API object.
* **-o=wide**   Output in the plain-text format with any additional information.
* **-o=yaml**   Output a YAML formatted API object.

## Howto: Example Deployment

gxctl is designed to be used in a declarative way, thus most commands are expecting a file to get provided. We generally recommend using the combination of `diff` and `apply` instead of using `create` and `update`.

The following files are given an easy example of a combination of deployment and configmap to deploy a NGINX including a specific content on a specific device called `7b6419fa-7ac5-4320-a87b-8fe8513130dc`.

***configmap.yaml***
```shell
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

***deployment.yaml***
```shell
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

## Howto: Device Selectors

We're using the concept of a device selector to make sure the right pods are running on the right device. There are two different types of selector's which can be added to a deployment, namely `matchByDeviceID` and `matchByLabels`. While `matchByDeviceID` is used as a 1:1 relation to allocate a deployment to a specific device using it's UUID, `matchByLabels` can be used to target a set of devices based on their labels. Those selector's are always working in the scope of an `application`.

For `matchByLabels`, the following apply to allocate a deployment to a set of devices:

* Always scoped on `application`.
* `matchByDeviceID` has precedence over `matchByLabels`
* All labels defined in a `matchByLabels` selector must be attached to the targeted devices in order to match.
* If there are multiple deployments matching a device for the same `application`, the one will be choosen, which matches most specificly, meaning having the biggest nummer of matching labels.
* If there are multiple deployments matching a device for the same `application`, and also those deployments having an equal number of matching labels, we use the most recent deployment based on it's creation timestamp.

**A concrete example:** 

Following we got a set of devices an deployments.

***Device01***
```shell
id: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
labels:
  core.gridx.ai/status: running
  gridx.de/channel: stable
  gridx.de/region: eu-central
```

***Device02***
```shell
id: de6dfaf0-bebd-4338-b4ce-bd451406b39a
labels:
  core.gridx.ai/status: running
  gridx.de/channel: stable
  gridx.de/region: eu-central
```

***Device03***
```shell
id: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
labels:
  core.gridx.ai/status: running
  gridx.de/channel: stable
  gridx.de/region: eu-west
```

***Deployment1***
```shell
matchByDeviceID: 7b6419fa-7ac5-4320-a87b-8fe8513130dc
```

***Deployment2***
```shell
matchByLabels:
  gridx.de/channel: stable
```

***Deployment3***
```shell
matchByLabels:
  gridx.de/channel: stable
  gridx.de/region: eu-west
```

We end up with the following allocation:

* `Device01` runs `Deployment1`. Even though `Deployment2` would also match, based on it's values, it runs on `Deployment1` as `matchByDeviceID` always has precedence over `matchByLabels`
* `Device02` runs `Deployment2` as the labels are matching.
* `Device03` runs `Deployment3`. Even though, both `Deployment2` and `Deployment3` are matching based on the labels, `Deployment3` is more specific (2 vs. 1 matching labels).

## Howto: Common operations

`gxctl create` - Create a new resource

```shell
# Create a new deployment
$ gxctl create -f new_deployment.json
# Create a new app
$ gxctl create app testapp
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
# get a List of all devices including the ones which were not yet online
$ gxctl get devices --all
# get a List of all devicess that match a serialnumber
$ gxctl get devices --serial D294-200-000-000-581-P-X
# get a List of all devices that match a serialnumber (wildcard)
$ gxctl get devices --serial 581-P-X 
# get a List of all devices that have a label with key gridx.de/channel
$ gxctl get devices --label gridx.de/channel
# get a List of all devices that have a label with key gridx.de/channel and value alpha
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
# get a List of all pods including the ones which were not yet started 
$ gxctl get pods --all
# Get a List of all pods on a certain device
$ gxctl get pods -d 57e82f8e-08f4-48f9-8e75-28552d09701f
# Get a List of all deployments 
$ gxctl get deploy
# Get information of a single device, showing its deployments
$ gxctl get deploy --device-id 57e82f8e-08f4-48f9-8e75-28552d09701f
# Get a information of a deployment include additional information using uuid abbreviation
$ gxctl get deploy c78 -o wide
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
