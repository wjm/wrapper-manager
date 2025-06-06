# wrapper-manager
A tool for managing multiple Wrapper instances

Only supports linux x86_64 arch

## Features
- Multi-instance management
- Add accounts at runtime (support 2FA)
- Multi-connection decryption
- gRPC API
- Get lyrics without an account
- Automatic region detection

## Deploy
```shell
# mkdir wrapper-manager && cd wrapper-manager
# go build github.com/WorldObservationLog/wrapper-manager
# ./wrapper --host localhost --port 8080
```
