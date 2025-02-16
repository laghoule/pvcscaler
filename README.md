# PVCScaler

[![Go Report Card](https://goreportcard.com/badge/github.com/laghoule/pvcscaler)](https://goreportcard.com/report/github.com/laghoule/pvcscaler)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=laghoule_pvcscaler&metric=coverage)](https://sonarcloud.io/summary/new_code?id=laghoule_pvcscaler)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=laghoule_pvcscaler&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=laghoule_pvcscaler)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=laghoule_pvcscaler&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=laghoule_pvcscaler)


PVCScaler is a command-line tool for scaling pods with Persistent Volume Claims (PVC) in Kubernetes.

## Features

- **Scale Up**: Scale up pods defined in the input state file.
- **Scale Down**: Scale down pods with PVCs, in specified namespaces and storage class with optional output state file.
- **Version Display**: Show the current version, git commit, and build date.

## Prerequisites

- Go 1.16+
- Configured Kubernetes cluster
- Properly configured kubeconfig file

## Installation

Clone the repository and build the executable:

```bash
git clone https://github.com/laghoule/pvcscaler.git 
cd pvcscaler 
go build -o pvcscaler
```

## Usage

### Basic Commands

- **Scale up**

```bash
./pvcscaler up -i <inputFile>
```

- **Scale down**

```bash
./pvcscaler down -n <namespace> -s <storageClass> -o <outputFile>
```

- **Show version**

```bash
./pvcscaler version
```

### Options

- `--kubeconfig, -k` : Path to the kubeconfig file (default: `$HOME/.kube/config`).
- `--dry-run, -d` : Dry run mode, makes no actual changes.
- `--namespace, -n` : Namespace to use (default: `all`).
- `--storageclass, -s` : Storage class to target (default: `default`).
- `--inputFile, -i` : State file for the `up` command.
- `--outputFile, -o` : State file for the `down` command.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue for any suggestions or bugs.
