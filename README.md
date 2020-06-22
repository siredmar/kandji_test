# gxctl
gxctl is a command line interface for running commands against the gridX device services clusters. This overview covers gxctl syntax, describes the command operations, and provides common examples.

## Installing / Setup

You should use the provided Makefile to build gxctl:

```
$ make build
$ cp bin/gxctl ${GOPATH//://bin:}/bin
```

Then, copy the provided base config file in to ~/.gxctl/config.yaml. This file mainly lists the settings for
authenticating gxctl for usage with the different subaccounts of Device Services. For each account, only the tenant
settings are set up. Before you can actually issue commands, you have to use `gxctl login` to retrieve a token by
authenticating with your gridx / Google account:

```
$ mkdir -p ~/.gxctl
$ cp config.yaml ~/.gxctl/config
$ gxctl login # will open a browser window where you can sign in using your gridx.de email
```

The generated token is only valid for a limited amount of time (currently 4 weeks); if you see errors regarding authentication
after that time, you may have to simply refresh the token by running `gxctl login` again.

## Profiles / Accounts

The managed devices are split into different server-side accounts; to differentiate between them when using gxctl commands, you
can pass the `--profile` argument. You will have to issue a `gxctl login` command separately for each profile that you want to use:

```
$ gxctl --profile viessmann login
$ gxctl --profile viessmann get device
```

See the default config file for the list of accounts / profiles.

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

* **config**   `gxctl config [ACTION] [TYPE] [flags]`
* **copy**   `gxctl copy [SOURCE] [DESTIONATION] [flags]`
* **create**   `gxctl create [[-f | ----filename]=Filename]`
* **delete**   `gxctl delete [TYPE] [NAME] [flags]`
* **get**   `gxctl get [TYPE] [NAME] [[-o | --output]=OUTPUT_FORMAT] [flags]`
* **label**   `gxctl label [TYPE] [NAME] [flags]`
* **update**   `gxctl update [TYPE] [NAME] [[-f | ----filename]=Filename] [flags]`
* **port-forward**   `gxctl port-forward [NAME] [LOCALPORT] [TARGET] [flags]`
* **ssh**   `gxctl ssh [NAME] [flags]`
* **syslog**   `gxctl syslog [NAME] [flags]`


## General resource types
* **applications**   Abbreviated alias `application`,`app`
* **devices**   Abbreviated alias `device`
* **deployments**   Abbreviated alias `deployment`,`deploy`
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

`gxctl config` - All device config related commands
```shell
# Get a List of docker configurations
$ gxctl config get docker
# Get a List of cleanup configurations
$ gxctl config get cleanup
# Get a List of docker configurations and include additional information (such as selectors).
$ gxctl config get docker -o wide
# Get a List of cleanup configurations and include additional information (such as selectors).
$ gxctl config get cleanup -o wide
# Create a docker config for AWS targeting all devices with label demo=demo
$ gxctl config create docker-aws https://123456789.dkr.ecr.eu-central-1.amazonaws.com --access-key-id=AAABBBCCCDDDEEE --secret-access-key=dpohx+EgPWQK+Fadsads123adeqwuIwnM4atH --region=eu-central-1 --selector demo=demo
# Delete a docker configuration
$ gxctl config delete docker 35e3dede-2b45-4212-82fb-b92f7d391e05 
```

`gxctl copy` - Copies files to devices
```shell
# Copy a local file to a device
$ gxctl copy ./testfile.tar.gz 57e82f8e-08f4-48f9-8e75-28552d09701f:/opt/incoming/testfile.tar.gz
# Copy a file from the device to the local machine
$ gxctl copy 57e82f8e-08f4-48f9-8e75-28552d09701f:/opt/incoming/testfile.tar.gz ./testfile.tar.gz
# Copy a local file using uuid abbreviation
$ gxctl copy ./testfile.tar.gz 57e:/opt/incoming/testfile.tar.gz
```

`gxctl create` - Create a new resource

```shell
# Create a new deployment
$ gxctl create -f new_deployment.json
# Create a new app
$ gxctl create app testapp
# Create a new nginx deployment for app testapp
$ gxctl create deployment nginx:1.15.8 -a testapp -s gridx.de/channel=stable
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

`gxctl update` - Updates a existing resource

```shell
# Update a device using a updatefile
$ gxctl update -f update_device.json
# Update labels of a device
$ gxctl update device 57e82f8e-08f4-48f9-8e75-28552d09701f -s "gridx.de/channel=stable gridx.de=area:west"
# Update mac address of a device
$ gxctl update device 57e82f8e-08f4-48f9-8e75-28552d09701f -a "11-22-33-44-55-66-77-88-99"
# Update maintenance window of a device
$ gxctl update device 57e82f8e-08f4-48f9-8e75-28552d09701f -m "Sun:11:00-Sun:13:00"
```

`gxctl apply` - Creates or Updates resources

```shell
# Creates a resource if not existing, otherwise updates.
$ gxctl apply -f deployment.json
```

`gxctl port-forward` - Forward an port from a device to a local port

```shell
# Forward port 8080 from the device on local port 4444
$ gxctl port-forward 57e82f8e-08f4-48f9-8e75-28552d09701f --localport 4444 --target 127.0.0.1:8080
# Forward port 8080 from a machine in the same network as the device (eg. router)  on local port 4444
$ gxctl port-forward 57e82f8e-08f4-48f9-8e75-28552d09701f --localport 4444 --target 192.168.0.1:8080
```

`gxctl ssh` - SSH to a devie

```shell
# SSH to device 
$ gxctl ssh 57e82f8e-08f4-48f9-8e75-28552d09701f
# SSH to device and execute an inital command
$ gxctl ssh -c "tail -f /var/log/syslog" 57e82f8e-08f4-48f9-8e75-28552d09701f
```

`gxctl syslog` - Stream device syslog

```shell
# Stream the current device syslog
$ gxctl syslog 57e82f8e-08f4-48f9-8e75-28552d09701f
```

`gxctl restart` - Restarts a device

```shell
# Restart the device
$ gxctl restart 57e82f8e-08f4-48f9-8e75-28552d09701f
```

`gxctl validate` - Validates a resource file

```shell
# Validates if required fields are set and that there are no unsupported fields
$ gxctl validate -f deployment.json
```

`gxctl diff` - Diff a resource file

```shell
# Shows the diff between the live system and the provided resource file
$ gxctl diff -f deployment.json
```

## Examples: Special operations

```shell
# Get the public key of a device
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f --show-publickey
# Get the docker config of a device
$ gxctl get device 57e82f8e-08f4-48f9-8e75-28552d09701f --show-dockerconfig
```

## Examples: Sorted output

You can sort the output of the get command using JSONPath expressions.

```shell
# Get a list of all pods sorted by their starttime
$ gxctl get pods -s .status.startTime
# Get a list of all devices sorted by their serialnumber in wide output
$ gxctl get devices -s .spec.serialnumber -o wide
# Get a list of all deployments sorted by their app
$ gxctl get deploy -s .spec.app
```
