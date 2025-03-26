# AIO Script Updater Made in Go

## This script basically establishes an SSH connection with the provided details and either uploads a file and replaces it remotely or downloads one and replaces it locally.

<br>   

### <strong>NOTE: the following is now optional, please see below for command line usage</strong>
#### Ensure you set the following global variables in `main.go` before you `go build` or `go run`:

```
-- localPath (filepath of file on local pc),

-- remotePath (filepath of file on remote server),

-- remoteUser (username of remote server),

-- remoteIP (IP of remote server),

-- sshPath (filepath for sshpass)

-- (optional) preemptiveTM (transfer mode)
```
### Regarding command line usage, please run `go build main.go` and following that `./main -h`