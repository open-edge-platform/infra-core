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

The former is used to pre-check data and the latter is used, once the data, have been validated to import the data in
Edge Infrastructure Manager using the Northboud REST APIs.

## Features

- Automate and reduce the number of individual steps required to add multiple new hosts
- Reduce the likelihood of data entry problems and human error in the process
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
- [gocover-cobertura](github.com/boumenot/gocover-cobertura) - check [$GOCOBERTURAVERSION_REQ](../version.mk)

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

Tools run as standalone binary and there are no difference when deploying in production or during developement and
testing phases.

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

Now, you can populate the csv file by appending details of systems like below -

```bash
  Serial,UUID,Error - do not fill
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
        file                   Required source CSV file to read data from
        url                    Required Edge Orchestrator URL
version                        Display version information
help                           Show this help message

OPTIONS:

--onboard                      If set, hosts will be automatically onboarded when connected
--project <name>               Optional project name in Edge Orchestrator. Alternatively, set env variable EDGEORCH_PROJECT
--os-profile <name/id>         Optional operating system profile name/id to configure for hosts. Alternatively, set env variable EDGEORCH_OSPROFILE
--site <name/id>               Optional site name/id to configure for hosts. Alternatively, set env variable EDGEORCH_SITE
--secure <value>               Optional security feature to configure for hosts. Alternatively, set env variable EDGEORCH_SECURE. Valid values: true, false
--remote-user <name/id>        Optional remote user name/id to configure for hosts. Alternatively, set env variable EDGEORCH_REMOTEUSER
--metadata <data>              Optional metadata to configure for hosts. Alternatively, set env variable EDGEORCH_METADATA. Metadata format: key=value&key=value

```

Before running the bulk import tool, project name can be optionally set in envioronment variable or can be passed
later as the last argument to import command. Examples below -

```bash
  export EDGEORCH_PROJECT=myproject
```

```bash
  ./orch-host-bulk-import import test.csv https://api.kind.internal myproject
```

Note that if the argument is passed along with the environment variable set, the argument shall take precedence.

The tool also requires authentication with the orchestrator before it can import hosts. There are two way to make
credentials available to the tool.

1. **Environment variables** - Set the username and password in environment variables `EDGEORCH_USER` and
`EDGEORCH_PASSWORD` respectively. You can use commands like below -

   ```bash
     export EDGEORCH_USER=myusername
     export EDGEORCH_PASSWORD=mypassword
   ```

   Run the bulk host import tool now.

   ```bash
     chmod +x orch-host-bulk-import
     ./orch-host-bulk-import import test.csv https://api.kind.internal
   ```

2. **Interactive shell** - If credentials are not provided via environment variables, the tool shall prompt for the
same during invocation like below -

   ```bash
     $ chmod +x orch-host-bulk-import
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
     Host Serial number : 2500JF3  UUID : 4c4c4544-2046-5310-8052-cac04f515233 registered. Name : host-a835ac40
     Host Serial number : ICW814D  UUID : 4c4c4544-4046-5310-8052-cac04f515233 registered. Name : host-17f57696
     Host Serial number : FW908CX  UUID : 4c4c4544-0946-5310-8052-cac04f515233 registered. Name : host-7bd98ae8
     CSV import successful
   ```

   However, if there are errors during registration, expect a new csv (with name `import_error_timestamp_filename`) to be
   generated with each failed line having corresponding error message. See below a sample invocation and failure.

   ```bash
     $ ./orch-host-bulk-import import test.csv https://api.kind.internal
     Importing hosts from file: test.csv to server: https://api.kind.internal
     Checking CSV file: test.csv
     Generating error file: import_error_2024-10-03T10:36:07Z_test.csv
     error: Failed to import all hosts
 
     $ cat import_error_2024-10-03T10\:36\:07Z_test.csv
     Serial,UUID,Error - do not fill
     JFSRQR3,4c4c4544-0046-5310-8052-cac04f515233,Host UUID already registered
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

[user-guide-url]: https://literate-adventure-7vjeyem.pages.github.io/edge_orchestrator/user_guide_main/content/user_guide/get_started_guide/gsg_content.html
[inframanager-dev-guide-url]: (https://literate-adventure-7vjeyem.pages.github.io/edge_orchestrator/user_guide_main/content/user_guide/get_started_guide/gsg_content.html)
[contributors-guide-url]: https://literate-adventure-7vjeyem.pages.github.io/edge_orchestrator/user_guide_main/content/user_guide/index.html
