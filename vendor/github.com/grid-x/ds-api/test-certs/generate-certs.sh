#!/usr/bin/env bash

cfssl gencert -initca ca-csr.json | cfssljson -bare ca

# Generate the client certs
cfssl gencert \
      -ca=ca.pem \
      -ca-key=ca-key.pem \
      -config=ca-config.json \
      -profile=client \
      client-csr.json | cfssljson -bare client

# Generate the auth service certs
cfssl gencert \
      -ca=ca.pem \
      -ca-key=ca-key.pem \
      -config=ca-config.json \
      -profile=server \
      device-backend-csr.json | cfssljson -bare device-backend

# Generate the jwt signing key pair
cfssl gencert \
      -ca=ca.pem \
      -ca-key=ca-key.pem \
      -config=ca-config.json \
      -profile=signing \
      jwt-csr.json | cfssljson -bare jwt
