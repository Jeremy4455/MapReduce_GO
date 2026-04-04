```bash
# 1. 更新系统
sudo apt update && sudo apt upgrade -y

wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz

# 4. 配置环境变量（永久生效）
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc

source ~/.bashrc

# 5. 验证安装
go version

git clone https://github.com/Jeremy4455/MapReduce_GO.git

cd MapReduce_GO/src/main/

sudo apt install gcc

GO111MODULE=off go build -buildmode=plugin -o wc.so ../mrapps/wc.go
```



更改worker的config.json

```json
{
	config: "100.31.46.217:1234"
}
```

运行master

```bash
GO111MODULE=off go run mrmaster.go data/pg-*.txt
```

运行worker

```bash
GO111MODULE=off go run mrworker.go wc.so
```

