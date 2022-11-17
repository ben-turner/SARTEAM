# SARTEAM

SARTEAM is a tool for managing SAR teams during an incident, tracking their
location and task status, and syncing that information with external tools such
as SARTopo. It was originally created as for Juan de Fuca Search and Rescue as a
successor to the excellent
[RadioTracker](https://github.com/jkatton/RadioTracker) tool created by Jesse
Katton, but providing a web interface so that it can be used by multiple team
members on different devices. It is designed to work in an offline/LAN
environment, and therefore does not require an internet connection.

_A note for less-technical users:_ This "README" file is contains more technical
information about the tool. The [Getting Started](#getting-started-with-sarteam)
section below should be all you need to get started using the tool. If you want
to know more about how the tool works, or how to customize it, you can read the
rest of this file.

## Getting started with SARTEAM

### Installation

The latest version of SARTEAM can be downloaded from the [releases
page](https://github.com/ben-turner/SARTEAM/releases). Download the file
appropriate for your operating system and run the downloaded file to install.

### Running SARTEAM

The SARTEAM application you installed is a server that runs on your computer.
Depending on how you installed it, you may be able to run it by double-clicking
the icon, or you may need to run it from the command line. If you need to run it
from the command line, you can do so by opening a terminal window and typing the
following:

    $ sarteam

Once the server is running, you can access the SARTEAM web interface by opening
a web browser and navigating to `http://localhost:8780`. If you are running
SARTEAM on a different computer, you can access it by navigating to
`http://<computer name or IP address>:8780`.

It is recommended that you configure SARTEAM to run automatically when your
computer starts. This will ensure that SARTEAM is always running when you need
it. The exact method for doing this will depend on your operating system.

## Building and local development

This application has two pieces: a server, written in Go, and a web interface
that is served by the server. The main entrypoint for the server is
`cmd/sarteam/main.go`. The web interface is located in the `web` directory.

### Building

Most build steps are handled by the `Makefile`. To build the server, run the
following command:

    $ make build

This will create a binary file in the `bin` directory. You can run this file
directly to start the server.

Packaged versions of SARTEAM can be built by running the following command:

    $ make package

This will create a packaged version of SARTEAM in the `dist` directory.
Configurations for building packages for different operating systems can be
found in the `build/package` directory.

## Contributing

Contributions are very much welcome! The primary goal of this fork is to meet
the needs of Juan de Fuca Search and Rescue, but we are happy to accept
contributions that are useful to other teams as well. Please open an issue or
pull request if you have any questions or suggestions.

## Design

### Assumptions

SARTEAM makes a few assumptions about the environment in which it is used. In no
particular order:

- SARTEAM will run on a secure network, and will not be exposed to the internet.
- An outbound internet connection is not always guaranteed to be available.
- Only one incident will be active at a time.
- Data must be resilient to sudden loss of power or other failures.
- Multiple users will be accessing the tool at the same time.

### Frontend

The frontend is a single-page application written in Vue.js. It is served by the
server, and is located in the `web` directory. The frontend communicates with
the server via a WebSocket connection. A WebSocket connection is used instead of
HTTP requests because it allows for bidirectional communication, and allows the
server to push updates to the frontend without the frontend needing to request
them. The frontend is built using the Vue CLI, and is configured to use the hash
history mode for routing. The frontend is built using the `npm run build`
command in the `web/` directory, and the resulting files are copied to the
`web/dist` directory. The `web/dist` directory is served by the server.

#### Design

SARTEAM is designed to be used by very non-technical users. The user interface
is designed to be as simple as possible, and to require as little training as
possible. An emphasis is placed on making the tool easy to use, and on making it
easy to find information quickly. Clear instructions are provided for each step
of the process, and the user is guided through each step. Failure modes are
handled gracefully, and clear error messages and troubleshooting steps are
provided to the user. Since SARTEAM is designed to be used in an offline
environment, it is important that the tool is usable even if the user does not
have an internet connection. This means that users may not be able to search for
information online in the event of a failure and therefore clear troubleshooting
instructions for all failure modes must be provided.

### Backend

The backend is a server written in Go. It is responsible for managing the data,
serving the frontend, and communicating with external tools. The server is
located in the `cmd/sarteam` directory. The server is built using the `make build` command in the root directory, and the resulting binary is located in the
`bin` directory.

#### Design

The server is designed to be as simple as possible. It is responsible for
managing the data, serving the frontend, and communicating with external tools.
It is not responsible for any other tasks, such as authentication or
authorization. The server is designed to be run on a secure network, and is not
designed to be exposed to the internet. All data is stored in memory as a tree
of structs, with `models.sarteam` as the root.

#### Storage

A storage directory is set using the config file. Within this directory,
incidents are stored in files with the `.incident` extension. These files are
designed to be human-readable, and can be edited using a text editor. The
filename should follow the format `[Training ]<date> <location>`, where `<date>`
is the date of the incident in the format `YYYY-MM-DD`, and `<location>` is the
location of the incident. For example, a training incident on January 1, 2020 at
Magdelena Point would be named `Training 2020-01-01 Magdelena Point.incident`.
The contents of the file are a timestamped log of mutations that were applied to
the incident. These mutations are applied to the incident when the file is
loaded. The mutations are stored in the file in the following format:
`<timestamp> <mutation type> <mutation data>`. This format is the same as the
commands sent from clients over a websocket connection, with the addition of the
timestamp. This format has the following advantages:

- The file is human-readable.
- A timeline of the incident can be reconstructed from the file.
- Files are appended to, rather than overwritten, so they are less likely to be
  corrupted by a power failure or other failure.
- Corrupted incident files can be fixed by manually removing the corrupted
  lines.
- The naming convention allows for easy sorting of incident files.
