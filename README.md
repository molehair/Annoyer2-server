# Annoyer2-server

Annoyer requires a server for timely notifications. This page is for the ones who want to deploy their own servers, not using the default one.

## Installation

> REQUIRED: The server requires to have a **https url**. Please prepare one.

Follow [this](https://firebase.google.com/docs/admin/setup/#go), download `service-account.json`, and choose one of the options below.

#### Using docker

1. Prepare two directories: `logs` and `config`.
2. Put `service-account.json` into `config`.
3. Run

```bash
docker run -v <your-logs-dir>:/logs -v <your-config-dir>/config:/config -p 8080:8080/tcp molehair/annoyer2server:latest
```

The server will listen on port 8080.

4. Update the server address at `lib/config.dart` on the client project.
5. Build and run the client.

#### Without docker

1. Install [Go](https://go.dev/).
1. Create `config` directory on the root of project.
2. Put `service-account.json` into `config`.
3. Run

```bash
go run .
```

The server will listen on port 8080.

4. Update the server address at `lib/config.dart` on the client project.
5. Build and run the client.