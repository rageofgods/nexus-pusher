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

* 

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
  bindAddress: "0.0.0.0"
  port: "8181"
  concurrency: 1
  credentials:
    test: "test"
  tls:
    auto: false
    domainName: "somedomain.org"
    enabled: false
    keyPath: "key"
    certPath: "cert"
```
* **concurrency** - how many parallel workers will be spawn
* **credentials** - list of 'user/password' to server auth
* **tls.auto** - enable Let's Encrypt cert generation (following domainName)
* **tls.domainName** - domain name for Let's Encrypt cert generation
* **enabled** - enables TLS server listening
* **keyPath** - absolute location of private key file
* **certPath** - absolute location of certificate file

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

* 0.1
    * Initial Release

## License

This project is licensed under the GNU License - see the LICENSE.md file for details
