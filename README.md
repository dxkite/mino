# 米诺

基于Go的http代理，支持网站访问权限限制，转发请求

## 特性

- [x] 代理访问
    - [x] HTTP/HTTPS
    - [x] Socks5
- [x] 自动PAC设置
- [x] Web服务
- [x] 开机自启
- [x] 自动更新
- [x] 热更新配置
- [x] 支持线路切换
- [x] 控制
    - [x] 支持域名黑名单允许拒绝访问特定域名
    - [x] 支持配置域名使用特定远程服务器
    - [x] 支持全部或者部分使用远程服务器
- [ ] 权限验证
    - [ ] 启用IP验证
    - [ ] 用户名认证
- [x] Web面板(desktop)
    - [x] 本地访问不验证权限 
    - [x] 配置界面
    - [x] 日志展示
    - [ ] 流量实时显示

## 移动端支持

- [ ] android **计划中**
- [ ] ios

## 多平台支持

从v0.2.1-alpha版本起，增加了对macOS的适配，并且原生支持M1！

## 使用

### 安装

```bash
go install dxkite.cn/mino/cmd/mino
```
### Docker支持
Docker一键启动服务端
```bash
docker run -d -p 28648:28648 w4ter/mino:2.0.2
```
### 命令行

`-addr :1080` 监听 `1080` 端口 支持 http/socks5 协议
`-upstream mino://127.0.0.1:8080`
`-pac_file conf/pac.txt` 启用PAC文件，自动设置系统Pac(windows)
```
mino -addr :1080 -pac_file conf/pac.txt -upstream mino://127.0.0.1:8080
```

`-addr :8080` 监听 `8080` 端口，支持 http/socks5/mino协议（需要配置加密密钥）
直连网络
使用公钥 `-cert_file conf/server.crt` 私钥 `-key_file conf/server.key` 加密连接
```
mino -addr :8080 -cert_file conf/server.crt  -key_file conf/server.key
```

### 使用配置

直接运行会加载  `mino.yml` 作为配置文件

```
mino
```

- 默认配置名 `mino.yml`

指定配置文件：
```
mino -conf config.yaml
```

### 配置文件示例

**客户端**

```yaml
address: ":1080"
upstream: "mino://127.0.0.1:28648"
```

**服务端**
```yaml
address: ":28648"
tls_cert_file: "conf/server.crt"
tls_key_file: "conf/server.key"
```
