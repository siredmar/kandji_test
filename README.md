# gxctl
gxctl is a command line interface for running commands against the gridX device services clusters. This overview covers gxctl syntax, describes the command operations, and provides common examples.

## Syntax

Use the following syntax to run gxctl commands from your terminal window:

`gxctl [command] [TYPE] [NAME] [flags]`

where `command`, `TYPE`, `NAME`, and `flags` are:

* **command**   Specifies the operation that you want to perform on one or more resources, for example `create`, `get`, `patch`
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

It's possible to abbreviate every uuid independent of it's type eg. Please not that identifiers not of the format of an uuid needs to be specified fully.

```shell
$ gxctl get pod 57e 21d
```

* **flags**   Specifies optional flags. For example, you can use the -o or --output flags to specify the output format of your command

## Operations

* **create**   `gxctl create [[-f | ----filename]=Filename]`
* **get**   `gxctl get [TYPE] [NAME] [[-o | --output]=OUTPUT_FORMAT] [flags]`
* **patch**   `gxctl patch [TYPE] [NAME] [[-f | ----filename]=Filename] [flags]`
* **delete**   `gxctl delete [TYPE] [NAME] [flags]`


## Resource types
* **applications**   Abbreviated alias `application`,`app`
* **devices**   Abbreviated alias `device`
* **deployments**   Abbreviated alias `deployment`,`deploy`
* **maintenance**
* **pods**   Abbreviated alias `pod`,`po`


## Output options

The default output format for all gxctl commands is the human readable plain-text format. To output details to your terminal window in a specific format, you can add either the -o or --output flags to a supported gxctl command.

```shell
$ gxctl [command] [TYPE] [NAME] -o=<output_format>
```

* **-o=json**   Output a JSON formatted API object.
* **-o=wide**   Output in the plain-text format with any additional information.
* **-o=yaml**   Output a YAML formatted API object.

## Examples: Common operations

`gxctl get` - List one or more resources.

```shell
# Get a List of all devices 
$ gxctl get devices
# Get a List of all devices and include additional information (such as labels).
$ gxctl get devices -o wide
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
# Get a information of a deployment include additional information using uuid abbreviation
$ gxctl get deploy c78 -o wide
```

`gxctl create` - Create a new resource.

```shell
# Create a new deployment
$ gxctl create -f new_deployment.json
# Create a new app
$ gxctl create app testapp
# Create a new nginx deployment for app testapp
$ gxctl create deployment nginx:1.15.8 -a testapp -l gridx.de/channel:stable
```

`gxctl patch` - Patch a existing resource.

```shell
# Patch a device using a patchfile
$ gxctl patch -f patch_device.json
# Patch labels of a device
$ gxctl patch device 57e82f8e-08f4-48f9-8e75-28552d09701f -l "gridx.de/channel:stable,gridx.de/area:west"
# Patch mac address of a device
$ gxctl patch device 57e82f8e-08f4-48f9-8e75-28552d09701f -a "11-22-33-44-55-66-77-88-99"
# Patch maintanence window of a device
$ gxctl patch device 57e82f8e-08f4-48f9-8e75-28552d09701f -m "Sun:11:00-Sun:13:00"
```

`gxctl delete` - Delete a existing resource.

```shell
# Delete an app
$ gxctl delete app testapp
# Delete an Deployment using the uuid abbreviation
$ gxctl delete deploy c78
# Delete two Deployments using both uuid abbreviation and full qualified name
$ gxctl delete deploy c78 35e3dede-2b45-4212-82fb-b92f7d391e05 
```

