# 说明
使用纯真数据库来查询ip

# 启动服务

## 配置项说明件说明

```shell
-b string
    base uri (default "/")
-c string
    ip数据文件路径也可以是一个url地址 (default "./qqwry.dat")
-cc string
    中国省市数据 (default "./china_city.json")
-p string
    服务端口 (default ":8080")

```

## systemd 方法启动
> vi /etc/systemd/system/ipservice.service


``` shell
[Unit]
Description=ipservice
Wants=network.target
After=network.target

[Service]
Type=simple
#执行用户
User=www
#文件执行路径
ExecStart=/path/to/program -b '/' -c '/path/to/qqwry.dat' -cc '/path/to/china_city.json' -p ":8080"
ExecReload=/bin/kill -USR1 $MAINPID
Restart=always
RestartSec=3
#打开文件句柄限制
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target

```


 启动服务

> systemctl start ipservice.service


 停止服务

> systemctl stop ipservice.service


 重启服务
> systemctl restart ipservice.service 


重新加载ip数据库
> systemctl reload ipservice.service 

完整systemd文档参见[systemd](https://wiki.archlinux.org/index.php/systemd)


## sysctl.conf


``` conf

net.ipv4.ip_forward = 0
net.ipv4.conf.default.rp_filter = 1
net.ipv4.conf.default.accept_source_route = 0
kernel.sysrq = 0

kernel.core_uses_pid = 1

net.ipv4.tcp_syncookies = 1

kernel.msgmnb = 65536

kernel.msgmax = 65536

kernel.shmmax = 68719476736

kernel.shmall = 4294967296

fs.file-max = 512000

net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
net.core.netdev_max_backlog = 250000
net.core.somaxconn = 3240000

#net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 10
net.ipv4.ip_local_port_range = 10000 65000
net.ipv4.tcp_max_syn_backlog = 262144
net.ipv4.tcp_max_tw_buckets = 20000
net.ipv4.tcp_fastopen = 3
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864
net.ipv4.tcp_mtu_probing = 1
net.ipv4.tcp_congestion_control = hybla

net.ipv4.tcp_timestamps=0
net.ipv4.tcp_max_orphans=262144
net.ipv4.tcp_synack_retries = 1
net.ipv4.tcp_syn_retries = 1
net.ipv4.tcp_keepalive_time = 30
fs.inotify.max_user_watches=512000

```

# 使用方法

> http://localhost:8080/location/114.114.114.144


``` json
{
"city": "江苏省南京市", //具体省市
"country": "中国",   //中国 | 外国
"area": "南京信风网络科技有限公司"
}

```



# 写给开发人员

## Required go 1.11+

## 编译

``` golang
//build linux amd64 二进制包
 make linux
//本地运行
 make dev

```


# 感谢

- 感谢[yinheli](https://github.com/yinheli)的[qqwry](https://github.com/yinheli/qqwry)项目，为我提供了纯真ip库文件格式算法