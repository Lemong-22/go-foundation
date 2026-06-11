Here is a clean step-by-step setup for **fresh Go on WSL Ubuntu**.

I recommend installing from the **official Go tarball**, not `apt`, because Ubuntu’s `apt` version can lag behind. The official Go docs also recommend removing any old `/usr/local/go` before extracting a new one. ([Go][1])

## 1. Open WSL Ubuntu

In Windows Terminal:

```bash
wsl
```

Or open your Ubuntu app directly.

## 2. Update Ubuntu packages

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget tar build-essential git
```

## 3. Remove old Go installation if any

```bash
sudo rm -rf /usr/local/go
```

Also check whether Go was installed via `apt`:

```bash
which go
go version
```

If it shows an old apt-installed Go, remove it:

```bash
sudo apt remove -y golang-go golang
sudo apt autoremove -y
```

## 4. Download latest Go for Linux AMD64

As of the current official Go downloads page, the latest stable Linux x86-64 archive shown is:

```bash
go1.26.3.linux-amd64.tar.gz
```

([Go][2])

Download it:

```bash
cd /tmp
wget https://go.dev/dl/go1.26.3.linux-amd64.tar.gz
```

## 5. Extract Go into `/usr/local`

```bash
sudo tar -C /usr/local -xzf go1.26.3.linux-amd64.tar.gz
```

This creates:

```bash
/usr/local/go
```

## 6. Add Go to your PATH

For WSL Ubuntu, edit your shell config.

If you use Bash:

```bash
nano ~/.bashrc
```

Add this at the bottom:

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

Save and reload:

```bash
source ~/.bashrc
```

The official install guide says `/usr/local/go/bin` needs to be added to `PATH`, then verified with `go version`. ([Go][1])

## 7. Verify installation

```bash
go version
```

Expected output should look like:

```bash
go version go1.26.3 linux/amd64
```

Also verify paths:

```bash
which go
go env GOPATH
go env GOROOT
```

Expected:

```bash
/usr/local/go/bin/go
/home/YOUR_USER/go
/usr/local/go
```

## 8. Create a test project

```bash
mkdir -p ~/code/go-hello
cd ~/code/go-hello
go mod init example.com/go-hello
```

Create `main.go`:

```bash
nano main.go
```

Paste:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello from Go on WSL Ubuntu")
}
```

Run it:

```bash
go run .
```

Expected:

```bash
Hello from Go on WSL Ubuntu
```

Build it:

```bash
go build -o app
./app
```

## 9. Install common Go dev tools

```bash
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

Verify:

```bash
gopls version
dlv version
staticcheck -version
```

## 10. VS Code setup for WSL

Open your Go project from inside WSL:

```bash
cd ~/code/go-hello
code .
```

Install these VS Code extensions:

```text
Go
WSL
```

The Go extension should detect `gopls`.

## 11. Recommended Go project layout

For simple apps:

```text
my-app/
├── go.mod
├── go.sum
├── main.go
└── README.md
```

For a backend/service:

```text
my-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   ├── service/
│   └── repository/
├── pkg/
├── go.mod
├── go.sum
└── README.md
```

## 12. Useful commands

```bash
go mod tidy        # clean dependencies
go test ./...      # run all tests
go fmt ./...       # format code
go vet ./...       # basic static checks
go build ./...     # compile all packages
```

## Clean reinstall later

When upgrading Go later, use the same pattern:

```bash
sudo rm -rf /usr/local/go
cd /tmp
wget https://go.dev/dl/goX.Y.Z.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf goX.Y.Z.linux-amd64.tar.gz
go version
```

Do **not** extract a new Go archive over an existing `/usr/local/go`; the official docs warn this can produce broken installations. ([Go][1])

[1]: https://go.dev/doc/install "Download and install - The Go Programming Language"
[2]: https://go.dev/dl/ "All releases - The Go Programming Language"
