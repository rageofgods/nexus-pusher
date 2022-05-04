# Nexus-pusher

Synchronise Sonatype Nexus repositories.

## Description

Nexus-pusher is a simple client\server application written in Go, to synchronise Sonatype Nexus repositories\
between two Nexus servers in split environments.

The goal is to be able to do repo sync without actually send assets data from one server to another\
to save bandwidth.

![Diagram](nexus-pusher.drawio.png)

### How it works
1. Nexus-pusher client is requesting full list of assets for specified repositories from Nexus Server 1 and Nexus Server 2 and find differences between them.
2. Nexus-pusher client sends diff from step one to Nexus-pusher server.
3. Nexus-pusher server analyze diff and download all assets from external repository (i.e https://registry.npmjs.org/, etc).
4. Nexus-pusher server upload all downloaded assets to Nexus Server 2

## Getting Started

### Supported repository types:
* NPM

### Installing

* How/where to download your program
* Any modifications needed to be made to files/folders

### Executing program

* How to run the program
* Step-by-step bullets
```
code blocks for commands
```

### Configuration examples
#### Server:
```yaml
server:
  enabled: true
  bindAddress: "0.0.0.0"
  port: "8181"
  concurrency: 1
  credentials:
    test: "test"
  tls:
    enabled: true
    keyPath: "key"
    certPath: "cert"
```
#### Client:
```yaml
client:
    daemon:
        enabled: true
        syncEveryMinutes: 30
    server: "http://X.X.X.X:8181"
    serverAuth:
        user: "test"
        pass: "test"
    syncConfigs:
        - srcServerConfig:
            server: "https://nexus.some"
            user: "user"
            pass: "pass"
            repoName: "repo1"
          dstServerConfig:
            server: "https://nexus.some"
            user: "user"
            pass: "pass"
            repoName: "repo2"
          format: "npm"
```

## Help

Any advise for common problems or issues.
```
command to run if program contains helper info
```

## Authors

Contributors names and contact info

ex. Eugene Khokhlov
ex. [@rageofgods](https://github.com/rageofgods)

## Version History

* 0.2
    * Various bug fixes and optimizations
    * See [commit change]() or See [release history]()
* 0.1
    * Initial Release

## License

This project is licensed under the [NAME HERE] License - see the LICENSE.md file for details
