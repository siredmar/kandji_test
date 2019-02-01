# ds-api
The GXDS apiserver

## Terminology

* **Application**: A software application that should run on gateways
* **Deployment**: An instance of an application, i.e. a certain configuration
  and version of the application. A deployment of an application is expected to
  be run on multiple gateways (zero or more).
* **Container**: An instance of the deployment running on a certain gateway. A
  container is running on exactly one gateway.

* **Device/Gateway**: used interchangeably for the same thing

## Overall Technical Concept

* The whole setup consists of a k8s cluster with a bunch of controllers and an
  API server that is the gateway for devices and management users to access the
  resources in the cluster.
* For all devices, pods, deployments and maintenance tasks we will have 1:1
  mappings to custom resources in the K8s cluster
* The K8s controller will transform different resources to each other to
  schedule workloads on the devices according to the manifests provided by the
  user
* For the MVP most of the API calls will essentially be translated and
  forwarded to the K8s cluster with a few optimizations around caching
  resources to reduce the number of calls, authenticate requests and perform
  authorization as well as domain specific validations before touching k8s (e.g.
  ensure there exists an application for a deployment - only the deployment will
  exist in k8s)
* We will create a new namespace (= K8s namespace) per account identified by
  the UUID which is generated when the account is first created
* All resources of a given account, i.e. device, pods, deployments, only exist
  in the corresponding namespace.


## API

Base URL: `https://api.gridx.ai/`

### Device

Base path: `/device/`

* **POST**  `/auth_request`: Allows to request a new token.

* *probably not needed*: **GET** `/`: Get spec and status of device
* **PUT**   `/`: Update the status of the device/gateway; matches
  to the status part of the device resource in k8s. It should set all the
  values, like device conditions, images, ...
* **PUT**   `/logs`: Upload system logs; for now just forward to LogDNA

* **GET**   `/pods/`: List all pods (only for the gateway of course)
* *probably not needed* **GET**   `/pods/{pod_id}`: Get individual Pod
* **PUT**   `/pods/{pod_id}/status`: Update status of pod with given ID

* *probably not needed*: **PUT** `/pods/{pod_id}/containers/{container_id}/status`:
  Update status of a specific container in the pod
* *in the future*: **PUT** `/pods/{pod_id}/containers/{container_id}/logs`:
  Upload logs per container

### Management

Base path: `/management/`

#### User and Auth specific stuff

All related resources need to be stored in the Postgres DB

* `/auth/`
* `/accounts/`
* `/users/`

#### Device

This is 1:1 mapping of K8s resources

* **GET**  `/devices/`: List all devices of this account
* **POST** `/devices/`: Create a new device
* **GET**  `/devices/{device_id}`: Fetch details of specific device
* **PUT**  `/devices/{device_id}`: Update device, e.g. modify labels
* **GET**  `/devices/{device_id}/pods`: List pods of specific device

* *in the future*: **GET** `/devices/{device_id}/logs`
* *in the future*: **POST** `/devices/{device_id}/`

#### Maintenance

Mostly directly mapping to K8s resources. Maintenance tasks are immutable and
cannot be modified after creation (because this will only lead to confusing
about the behavior of a maintenance task when it is changed: should it be rerun
for boxes that have already executed it? Should it be stopped for boxes that
are currently running it? ... )

* **GET** `/maintenance/`: List existing maintenance tasks
* **POST** `/maintenance/`: Create a new maintenance task for existing devices
* **GET** `/maintenance/{id}`: Fetch details of a single maintenance task
* **DELETE** `/maintenance/{id}`: Delete a maintenance task.

#### Application

Applications should only exist within the Postgres DB. They are used for
validation purposes only.

For now applications only have a name and a description. In addition,
applications are immutable and cannot be modified after creation (because
renaming applications could cause strange side effects).

* **GET** `/applications/`: List all applications
* **GET** `/applications/{application_id}`: Fetch an individual application
* **DELETE** `/application/{application_id}`: Delete an application

#### Deployment

Mostly directly mapping to K8s resources.

* **GET**  `/deployments/`: List all existing deployments. Should allow
  filtering by application
* **GET**  `/deployments/{deployment_id}`: Get details of an individual
  deployment.
* **PUT**  `/deployments/{deployment_id}`: Update an existing deployment
* **DELETE** `/deployments/{deployment_id}`: Delete an existing deployment (as
  well as all pods - which will happen automatically)
* **POST** `/deployments/`: Create a new deployment: we need to validate if the
  application of the deployment exists.

#### Pods

Mostly directly mapping to K8s resources.

Pods are read-only as they can only be created indirectly via deployments.

* **GET** `/pods/`: List pods
* **GET** `/pods/{pod_id}`: Get details of specific pod

* *in the future* **GET**  `/pods/{pod_id}/containers/{container_id}/logs`:
  Fetch logs of a container
* *in the future* **POST** `/pods/{pod_id}/containers/{container_id}/exec`:
  Exec into a container (Kubernetes does this using an connection protocol
  upgrade to HTTP2 for being able to multiplex stdin, stdout, and stderr in the
  same stream)

## Build

```shell
make bin/server
```

## Running

To start the server run:
```shell
$ # Start a k8s cluster using minikube
$ minikube start
$ # Create CRDs
$ kubectl apply -f vendor/github.com/grid-x/ds-k8s/config/crds
$ # Start a postgres DB
$ docker run --rm -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=pw --name postgres --mount type=tmpfs,destination=/var/lib/postgresql/ -p 5432:5432 postgres:9.6
$ # Start the server (NOTE: change the postgres url accordingly)
$ bin/server -auth.rsa-key ./test-certs/jwt.pem -k8s.config ~/.kube/config -postgres.migrate -postgres.url="postgres://postgres:pw@192.168.99.100:5432/postgres?sslmode=disable"
```

### Management API Access

To generate a valid token run:
```shell
$ jwtgen -a RS256 -p ./test-certs/jwt-key.pem -c "sub=auth0|5969f0804349fc3abb247fc2" -c "email=j.hermanns@gridx.de"
```

And to make an API request run:
```shell
$ curl -H 'Accept: 2018-11-02' -H 'Authorization: Bearer <token> localhost:8080/api/management/devices
```

### Device API Access

To generate a valid token run:
```shell
$ jwtgen -a RS256 -p ./test-certs/jwt-key.pem -c "sub=foo" -c "accountID=default"
```

```shell
$ # Create a device "foo" with the sample device file
$ kubectl apply -f test/k8s/core_v1beta1_device.yaml
$ # Create a devicepod for device "foo" with the sample file
$ kubectl apply -f test/k8s/core_v1beta1_devicepod.yaml
$ # You can now use the device-api to get the pods for the device
$ curl -H 'Accept: 2018-11-02' -H 'Authorization: Bearer <token> localhost:8080/api/device/pods
```
