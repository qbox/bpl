BPL - Binary Processing Language
============

## 快速入门

了解 BPL 最快的方式是学习 qbpl 和 qbplproxy 两个实用程序：

### qbpl

qbpl 可用来分析任意的文件格式。使用方法如下：

```
qbpl [-p <protocol>.bpl -o <output>.log] <file>
```

多数情况下，你不需要指定 `-p <protocol>.bpl` 参数，我们根据文件后缀来确定应该使用何种 protocol 来解析这个文件。例如：

```
qbpl 1.gif
```

不过为了让 qbpl 能够找到所有的 protocols，我们需要先安装：

```
make install # 这将将所有的bpl文件拷贝到 ~/.qbpl/formats/
```

### qbplproxy

qbplproxy 可用来分析服务器和客户端之间的网络包。它通过代理要分析的服务，让客户端请求自己来分析请求包和返回包。使用方式如下：

```
qbplproxy -h <listenIp:port> -b <backendIp:port> [-p <protocol>.bpl -f <filter> -o <output>.log]
```

其中，`<listenIp:port>` 是 qbplproxy 自身监听的IP和端口，`<backendIp:port>` 是原始的服务。`-p <filter>` 是过滤条件，这个条件通过 BPL_FILTER 全局变量传递到 bpl 中。

多数情况下，你不需要指定 `-p <protocol>.bpl` 参数，qbplproxy 程序可以根据你监听的端口来猜测网络协议。例如：

```
mongod --port 37017
qbplproxy -h localhost:27017 -b localhost:37017
```

我们会依据端口 27017 知道你要分析的是 mongodb 的网络协议。


## BPL 文法

请参见 [BPL 文法](README_BPL.md)。


## 网络协议研究

### RTMP 协议

格式描述：

* [rtmp.bpl](https://github.com/qbox/bpl/blob/develop/formats/rtmp.bpl)

测试：

1) 启动一个 rtmp server，让其监听 1936 端口（而不是默认的 1935 端口）。比如我们可以用 [node-rtsp-rtmp-server](https://github.com/iizukanao/node-rtsp-rtmp-server)：

```
git clone git@github.com:iizukanao/node-rtsp-rtmp-server.git
cd node-rtsp-rtmp-server
修改 config.coffee，将：
  * rtmpServerPort: 1935 改为 rtmpServerPort: 1936；
  * serverPort: 80 改为 serverPort: 8080（这样就不用 sudo 来运行了）
coffee server.coffee
```

2) 启动 qbplproxy：

```
qbplproxy -h localhost:1935 -b localhost:1936 -p formats/rtmp.bpl | tee <output>.log
```

3) 推流：

```
ffmpeg -re -i test.m4v -c:v copy -c:a copy -f flv rtmp://localhost/live/123
```

4) 播流：

在 Mac 下可以考虑用 VLC Player，打开网址 rtmp://localhost/live/123 进行播放即可。

5) 选择性查看

有时候我们并不希望看到所有的信息，rtmp.bpl 支持以 flashVer 作为过滤条件。如：

```
qbplproxy -f 'flashVer=LNX 9,0,124,2' -h localhost:1935 -b localhost:1936 -p formats/rtmp.bpl | tee <output>.log
```

这样就可以只捕获 VLC Player 的播流过程了。

### WebRTC 协议

格式描述：

* [webrtc.bpl](https://github.com/qbox/bpl/blob/develop/formats/webrtc.bpl)


### MongoDB 协议

格式描述：

* [mongo.bpl](https://github.com/qbox/bpl/blob/develop/formats/mongo.bpl)

测试：

1) 启动 MongoDB，让其监听 37017 端口（而不是默认的 27017 端口）：

```
./mongod --port 37017
```

2) 启动 qbplproxy：

```
qbplproxy -h localhost:27017 -b localhost:37017 -p formats/mongo.bpl | tee <output>.log
```

3) 使用 MongoDB，比如通过 mongo shell 操作：

```
./mongo
```

## 文件格式研究

### MongoDB binlog 格式

TODO

### MySQL binlog 格式

TODO

### HLS TS 格式

格式描述：

* [ts.bpl](https://github.com/qbox/bpl/blob/develop/formats/ts.bpl)

测试：TODO

### FLV 格式

格式描述：

* [flv.bpl](https://github.com/qbox/bpl/blob/develop/formats/flv.bpl)

测试：TODO

### GIF 格式

格式描述：

* [gif.bpl](https://github.com/qbox/bpl/blob/develop/formats/gif.bpl)

测试：

```
qbpl -p formats/gif.bpl formats/1.gif
```
