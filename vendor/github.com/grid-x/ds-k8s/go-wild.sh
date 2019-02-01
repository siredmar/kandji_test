#!/usr/bin/env bash

kubebuilder create api --group core --version v1beta1 --kind Device --controller=false
kubebuilder create api --group core --version v1beta1 --kind DeviceContainer --controller=false

kubebuilder create api --group apps --version v1beta1 --kind DeviceDeployment

#kubebuilder create api --group batch --version v1beta1 --kind DeviceJob

#kubebuilder create api --group maintenance --version v1beta1 --kind DeviceFleetMaintenanceTask
#kubebuilder create api --group maintenance --version v1beta1 --kind DeviceMaintenanceTask
