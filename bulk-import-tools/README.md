# Edge Infrastructure Manager Bulk Import Tools

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Get Started](#get-started)
- [Contribute](#contribute)

## Overview

This sub-repository contains the Bulk Import Tools for non-interactive onboarding of edge devices in Edge
Infrastructure Manager. Two tools have been created to automate the registration of multiple edge nodes in
Edge Infrastructure Manager.

1. orch-host-preflight
2. orch-host-bulk-import

The former is used to pre-check the data and the latter is used, once the data have been validated, to import the data into
Edge Infrastructure Manager using the Northboud REST APIs.

## Features

- Automated and reduced the number of individual steps required to add multiple new hosts
- Reduced the likelihood of data entry problems and human error in the process
- Support for well-known CSV files that are used to inject data into the tools
- Built with the support for Multitenancy

## Get Started

Instructions on how to install and set up the bulk import tools on your development machine.

### Dependencies

Firstly, please verify that all dependencies have been installed. This code requires the following tools to be
installed on your development machine:

- [Go\* programming language](https://go.dev) - check [$GOVERSION_REQ](../version.mk)
- [golangci-lint](https://github.com/golangci/golangci-lint) - check [$GOLINTVERSION_REQ](../version.mk)
- [go-junit-report](https://github.com/jstemmer/go-junit-report) - check [$GOJUNITREPORTVERSION_REQ](../version.mk)
- Python\* programming language version 3.10 or later
- [gocover-cobertura](https://github.com/boumenot/gocover-cobertura) - check [$GOCOBERTURAVERSION_REQ](../version.mk)

### Build the Binary

Build the project as follows:

```bash
# Build go binary
make build
```

The binaries are installed in the [$OUT_DIR](../common.mk) folder:

- orch-host-preflight
- orch-host-bulk-import

### Usage

Tools run as standalone binaries and their deployment process is consistent across production, development, and testing phases.

#### Pre-flight tool

```bash
  Create an empty template and scrutinize input CSV file for orch-host-bulk-import tool.

  Usage: orch-host-preflight COMMAND

  Commands:
    generate <output.csv>  Generate a template CSV file with the given filename
    check <input.csv>      Check the contents of the given CSV file
    version                Display version information
    help                   Display this help information
```

Run the pre-flight tool after you step into the `out/` directory

```bash
  cd out
  chmod +x orch-host-preflight
  ./orch-host-preflight generate test.csv
```

Now, you can populate the csv file by appending details of systems like below. Do not change the first line
`Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill` because that is the expected format. You
only need to fill in the first two columns, `Serial` and `UUID`, with the serial number and UUID of the edge node(s)
you want to register. The other columns are not meant for this stage.

```bash
  Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill
  2500JF3,4c4c4544-2046-5310-8052-cac04f515233
  ICW814D,4c4c4544-4046-5310-8052-cac04f515233
  FW908CX,4c4c4544-0946-5310-8052-cac04f515233
```

With the manual entries in place, you can go ahead and validate the csv. Note that you need to provide the same
filename you provided in the previous command or skip to default. If there are errors in the input file, expect a new
csv(`preflight_error_timestamp_filename`) to be generated with error messages corresponding to each record in the csv.

```bash
  ./orch-host-preflight check test.csv
```

#### Bulk import tool

```bash
  Import host data from input file into the Edge Orchestrator.

  Usage: orch-host-bulk-import COMMAND

  COMMANDS:

  import [OPTIONS] <file> <url>  Import data from given CSV file to orchestrator URL
        file                     Required source CSV file to read data from
        url                      Required Edge Orchestrator URL
  version                        Display version information
  help                           Show this help message

  OPTIONS:

  --onboard                      Optional onboard flag. If set, hosts will be automatically onboarded when connected
  --project <name>               Required project name in Edge Orchestrator. Alternatively, set env variable EDGEORCH_PROJECT
  --os-profile <name/id>         Optional operating system profile name/id to configure for hosts. Alternatively, set env variable EDGEORCH_OSPROFILE
  --site <name/id>               Optional site name/id to configure for hosts. Alternatively, set env variable EDGEORCH_SITE
  --secure <value>               Optional security feature to configure for hosts. Alternatively, set env variable EDGEORCH_SECURE. Valid values: true, false
  --remote-user <name/id>        Optional remote user name/id to configure for hosts. Alternatively, set env variable EDGEORCH_REMOTEUSER
  --metadata <data>              Optional metadata to configure for hosts. Alternatively, set env variable EDGEORCH_METADATA. Metadata format: key=value&key=value
```

The fields `OSProfile`, `Site`, `Secure`, `RemoteUser`, and `Metadata` are used for configuration of provisioning for the Edge Node. `OSProfile`, `Site`, and `RemoteUser` are fields that allow both name and ID to be used. The `Secure` field is a boolean value that can be set to `true` or `false`. The `Metadata` field is a key-value pair separated by an `=` sign, and multiple key-value pairs are separated by an `&` sign.

Complete the CSV file with the provisioning details for the edge nodes you want to register. `OSProfile` is a mandatory field
here without which provisioning configuration cannot be completed. Also, be aware that the `OSProfile` and `Secure` fields are
related. If `Secure` is set to `true`, the `OSProfile` must support it. If left blank, `Secure` defaults to `false`. The value
in other fields are validated before consumption though an empty string is allowed for all of them.
The following is an example:

```bash
  Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill
  2500JF3,4c4c4544-2046-5310-8052-cac04f515233,os-7d650dd1,site-08c1e377,true,localaccount-9dfb57cb,key1=value1&key2=value2,
  ICW814D,4c4c4544-4046-5310-8052-cac04f515233,ubuntu-22.04-lts-generic,Folsom,true,myuser-key,key1=value1&key2=value2,
  FW908CX,4c4c4544-0946-5310-8052-cac04f515233,os-7d650dd1,Folsom,true,myuser-key,key1=value1&key2=value2,
```

Before running the bulk import tool, project name can either be set in environment variable or passed later as an optional
parameter to import command. Examples below -

```bash
  export EDGEORCH_PROJECT=myproject
```

```bash
  ./orch-host-bulk-import import --project myproject test.csv https://api.kind.internal
```

There are several other optional parameters that can be set in the environment or passed as optional parameters to the import
command. The following are examples:

```bash
  export EDGEORCH_OSPROFILE=myosprofile
  export EDGEORCH_SITE=mysite
  export EDGEORCH_SECURE=true
  export EDGEORCH_REMOTEUSER=myremoteuser
  export EDGEORCH_METADATA=key1=value1&key2=value2
```

```bash
  ./orch-host-bulk-import import --onboard --os-profile myosprofile --site mysite --secure true --remote-user myremoteuser --metadata key1=value1&key2=value2 test.csv https://api.kind.internal
```

Note that for all the options (except onboard), if optional parameter is passed along with the environment variable set, the
optional parameter will take precedence. If either the environment variable or the optional parameter is set, they act as global
values for the corresponding field in the input file and override the local value for all rows.

The tool also requires authentication with the orchestrator before it can import hosts. There are two way to make
credentials available to the tool.

1. **Environment variables** - Set the username and password in environment variables `EDGEORCH_USER` and
`EDGEORCH_PASSWORD` respectively. You can use commands like below -

   ```bash
    export EDGEORCH_USER=myusername
    export EDGEORCH_PASSWORD=mypassword
   ```

2. **Interactive shell** - If credentials are not provided via environment variables, the tool shall prompt for the
same during invocation like below -

   ```bash
     $ ./orch-host-bulk-import import test.csv https://api.kind.internal
     Importing hosts from file: test.csv to server: https://api.kind.internal
     Checking CSV file: test.csv
     Enter Username: myusername
     Enter Password: mypassword
   ```

   Service URL is a mandatory argument to the import command. You can optionally provide a name of the csv file you want
   to use as source of hosts else it defaults to `edge_nodes.csv`. Also provide the option `--onboard` if it is desireable
   to auto onboard the hosts in which case the command should appear like below.

   ```bash
     ./orch-host-bulk-import import --onboard orch.csv https://api.kind.internal
   ```

   The bulk import tool validates the input file again similar to the pre-flight tool and generates an error report if
   validation fails. If validation passes, the bulk import tool proceeds to registration phase. For each host
   registration that succeeds, expect output similar to below on console.

   ```bash
      ✔ Host Serial number : 2500JF3  UUID : 4c4c4544-2046-5310-8052-cac04f515233 registered. Name : host-a835ac40
      ✔ Host Serial number : ICW814D  UUID : 4c4c4544-4046-5310-8052-cac04f515233 registered. Name : host-17f57696
      ✔ Host Serial number : FW908CX  UUID : 4c4c4544-0946-5310-8052-cac04f515233 registered. Name : host-7bd98ae8
      CSV import successful
   ```

   However, if there are errors during registration, expect a new csv (with name `import_error_timestamp_filename`) to be
   generated with each failed line having corresponding error message. See below a sample invocation and failure.

   ```bash
	   $ ./orch-host-bulk-import import --onboard --project testProject test.csv https://api.CLUSTER_FQDN
	   Importing hosts from file: test.csv to server: https://api.CLUSTER_FQDN
	   Onboarding is enabled
	   Checking CSV file: hosts.csv
	   Generating error file: import_error_2025-04-15T18:28:44+05:30_test.csv
	   error: Failed to import all hosts

	   $ cat import_error_2025-04-15T18\:28\:44+05\:30_test.csv
	   Serial,UUID,OSProfile,Site,Secure,RemoteUser,Metadata,Error - do not fill
	   FW908CX,4c4c4544-0946-5310-8052-cac04f515233,os-7d650dd1,Folsom,true,myuser-key,key1=value1&key2=value2,Host already registered
   ```

See the [documentation][user-guide-url] if you want to learn more about using Edge Orchestrator.

## Contribute

To learn how to contribute to the project, see the [contributor's guide][contributors-guide-url]. The project will
accept contributions through Pull-Requests (PRs). PRs must be built successfully by the CI pipeline, pass linters
verifications and the unit tests.

There are several convenience make targets to support developer activities, you can use `help` to see a list of makefile
targets. The following is a list of makefile targets that support developer activities:

- `lint` to run a list of linting targets
- `test` to run the tools unit test
- `go-tidy` to update the Go dependencies and regenerate the `go.sum` file
- `build` to build the project and generate executable files

See the [docs](docs) for advanced development topics:

- [Downloading Released Tools](docs/download.md)

To learn more about internals and software architecture, see
[Edge Infrastructure Manager developer documentation][inframanager-dev-guide-url].

[user-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html
[inframanager-dev-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/infra_manager/index.html
[contributors-guide-url]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html

Last Updated Date: April 10, 2025
