# MapReduce GO

This project is a lightweight distributed MapReduce framework implemented in Go.  
It supports a single-Master, multi-Worker architecture and uses RPC communication to collaboratively complete large-scale text WordCount tasks.

---

## 🚀 Environment Setup (Ubuntu/Linux)

Before running the project, make sure your system has a working Go toolchain and CGO support configured.

### 1. Install Basic Dependencies

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install GCC (required to compile .so plugins)
sudo apt install gcc -y
```

### 2. Install Go

```bash
# Download Go 1.26.1
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz

# Configure environment variables (append to .bashrc)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

---

## 🛠️ Project Initialization and Build

### 1. Clone the Repository

```bash
git clone https://github.com/Jeremy4455/MapReduce_GO.git

# Make sure to run from this directory
cd MapReduce_GO/src/main/
```

### 2. Build the WordCount Plugin

Because Go plugin mode has strict build-environment requirements, make sure this command is executed on the machine where the Worker runs:

```bash
GO111MODULE=off go build -buildmode=plugin -o wc.so ../mrapps/wc.go
```

### 3. Configure the Master IP

Edit `config.json` and set the address to the public/private IP of the Master node:

```bash
{
    "masterAddr": "<Master_IP_Address>:1234"
}
```

---

## 💻 Demo Run

### Step 1: Start the Master Node

The Master listens for RPC requests, maintains the task queue, and monitors Worker status.

```bash
GO111MODULE=off go run mrmaster.go data/pg-*.txt
```

<img src="./assets/images/master.png" alt="master" style="zoom: 50%;" />

---

### Step 2: Start Worker Nodes

You can launch Workers simultaneously in multiple terminals or on multiple cloud servers. Workers will automatically register with the Master and fetch tasks.

```bash
GO111MODULE=off go run mrworker.go wc.so
```

<img src="./assets/images/worker1.png" alt="master" style="zoom: 50%;" />

<img src="./assets/images/worker2.png" alt="master" style="zoom: 50%;" />

---

### Step 3: Verify Results

After the Master outputs `All tasks finished`, the system will automatically aggregate the results.

```bash
# List generated output files
ls output/

# Show top 10 by word frequency
cat output/mr-out-* | sort -k2nr | head -n 10
```

---

## ⚠️ Notes

1. **Firewall**: Make sure port `1234` on the Master node is allowed in the security group/firewall.
2. **Cleanup**: Before rerunning, it is recommended to execute `rm -f mr-out-*` to remove intermediate files.

You may adjust field names in `config.json` as needed (for example, from `config` to `masterAddr`) to match your actual implementation.

If you modify source code, remember to rebuild `wc.so` afterward.
