# 链接测试

## 客户端这一侧需要修改能链接的端口数

```bash
# 查看可用的链接端口数
sysctl net.ipv4.tcp_max_tw_buckets net.ipv4.ip_local_port_range
#
#net.ipv4.tcp_max_tw_buckets = 65536
#net.ipv4.ip_local_port_range = 32768	60999

# 临时修改
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"
```

## 服务端这一侧需要修改最能打开的文件数

vim /etc/security/limits.conf
```
* soft nofile 4096
* hard nofile 4096
```

## Nginx


## 以下内容无影响 

vim /etc/sysctl.conf

```
net.ipv4.tcp_mem  =   379008       505344  758016
net.ipv4.tcp_wmem = 4096        16384   4194304
net.ipv4.tcp_rmem = 4096          87380   4194304
net.core.wmem_default = 8388608
net.core.rmem_default = 8388608
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
```
