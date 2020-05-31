# 共享文件系统

面向对象技术与方法作业2。实现网络文件共享系统，其主要需求描述如下：

本系统中拥有一个中心文件服务器，系统的所有用户都可以向这个中心文件服务器上传文件，也能查询当前文件清单，从中选择文件下载。

## 编译

项目用Go语言编写，首先安装[Go](https://golang.org/doc/install)

下载项目

```
git clone https://github.com/lishengye/sfs.git
```

安装依赖

```
go mod vendor
```

建立输出目录

```
mkdir output
```

服务端

```
go build -o ./output/sfs  ./cmd/server
```

复制配置文件

```
cp config/sfs.json  output/
```

客户端

```
go build -o ./output/sfc  ./cmd/client
```

## 运行

#### 服务端

运行：

```
sfs -c sfs.json
```

sfs为配置文件，指定运行端口和工作目录

运行为daemon，Linux：

```
nohup ./sfs -c sfs.json > log.txt 2>&1 &
```

#### 客户端

运行

```
sfc -s remote_ip -p 6679 -u usename

-s：服务端ip

-p：服务端运行端口，默认6679

-u：用户名
```

进入交互命令行，支持的命令有

```
list，download，upload，exit，help
```

输入help查看帮助

## 功能

#### 已实现

1. 获取文件列表
2. 文件上传，下载

#### 待实现

1. 多线程下载
2. 断点续传