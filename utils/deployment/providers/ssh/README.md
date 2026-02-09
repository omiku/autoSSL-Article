# SSH 部署提供者

这个部署提供者支持通过SSH协议将SSL证书部署到远程服务器，并支持通过跳板机（Jump Server/Bastion Host）进行连接。

## 功能特性

- ✅ 直接SSH连接
- ✅ 多级跳板机连接（支持跳板机链）
- ✅ 密码认证和密钥认证
- ✅ SCP/SFTP文件传输（自动回退）
- ✅ 详细的日志记录
- ✅ 超时控制
- ✅ 连接池优化

## 配置示例

### 直接连接

```json
{
  "provider": "ssh",
  "config": {
    "ssh_host": "target-server.com",
    "ssh_port": 22,
    "ssh_username": "deploy",
    "ssh_password": "password123",
    "cert_path": "/etc/ssl/certs/server.crt",
    "key_path": "/etc/ssl/private/server.key",
    "reload_command": "sudo systemctl reload nginx"
  }
}
```

### 单级跳板机

```json
{
  "provider": "ssh",
  "config": {
    "ssh_host": "target-server.internal",
    "ssh_port": 22,
    "ssh_username": "deploy",
    "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
    "cert_path": "/etc/ssl/certs/server.crt",
    "key_path": "/etc/ssl/private/server.key",
    "reload_command": "sudo systemctl reload nginx",
    "jump_servers": [
      {
        "ssh_host": "bastion.company.com",
        "ssh_port": 22,
        "ssh_username": "admin",
        "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
      }
    ]
  }
}
```

### 多级跳板机

```json
{
  "provider": "ssh",
  "config": {
    "ssh_host": "db-server.prod.internal",
    "ssh_port": 22,
    "ssh_username": "deploy",
    "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
    "cert_path": "/etc/ssl/certs/server.crt",
    "key_path": "/etc/ssl/private/server.key",
    "reload_command": "sudo systemctl reload postgresql",
    "jump_servers": [
      {
        "ssh_host": "bastion-1.company.com",
        "ssh_port": 22,
        "ssh_username": "admin",
        "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
      },
      {
        "ssh_host": "bastion-2.prod.internal",
        "ssh_port": 2222,
        "ssh_username": "deploy",
        "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
      }
    ]
  }
}
```

## 配置字段说明

### 主要配置

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `ssh_host` | string | 是 | 目标服务器地址 |
| `ssh_port` | int | 否 | 目标服务器端口，默认22 |
| `ssh_username` | string | 是 | SSH用户名 |
| `ssh_password` | string | 否 | SSH密码（与密钥二选一） |
| `ssh_key` | string | 否 | SSH私钥内容（与密码二选一） |
| `cert_path` | string | 是 | 证书文件在远程服务器上的路径 |
| `key_path` | string | 是 | 私钥文件在远程服务器上的路径 |
| `reload_command` | string | 否 | 证书部署后执行的重新加载命令 |

### 跳板机配置

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `jump_servers` | array | 否 | 跳板机配置数组 |
| `jump_servers[].ssh_host` | string | 是 | 跳板机地址 |
| `jump_servers[].ssh_port` | int | 否 | 跳板机端口，默认22 |
| `jump_servers[].ssh_username` | string | 是 | 跳板机用户名 |
| `jump_servers[].ssh_password` | string | 否 | 跳板机密码 |
| `jump_servers[].ssh_key` | string | 否 | 跳板机私钥 |

## 认证方式

### 1. 密码认证
```json
{
  "ssh_username": "deploy",
  "ssh_password": "your_password"
}
```

### 2. 密钥认证
```json
{
  "ssh_username": "deploy",
  "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nyour_private_key_here\n-----END OPENSSH PRIVATE KEY-----"
}
```

### 3. 混合认证（推荐）
```json
{
  "ssh_username": "deploy",
  "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
  "ssh_password": "fallback_password"
}
```

## 跳板机工作原理

1. **连接链建立**：从本地 -> 跳板机1 -> 跳板机2 -> ... -> 目标服务器
2. **隧道复用**：每个连接都会复用前一个连接的隧道
3. **超时控制**：每个连接步骤都有独立的超时设置
4. **错误处理**：任何一步失败都会返回详细的错误信息

## 日志输出

部署过程会输出详细的日志信息：

```
[INFO] 建立SSH连接 | 目标: target-server:22 | 跳板机数量: 2
[INFO] 通过跳板机建立连接 | 跳板机1: bastion-1:22
[INFO] 通过跳板机建立连接 | 跳板机2: bastion-2:2222
[SUCCESS] SSH连接建立成功 | 目标: target-server:22
[INFO] 开始部署证书...
[SUCCESS] 证书部署完成 | 远程路径: /etc/ssl/certs/server.crt
[SUCCESS] 私钥部署完成 | 远程路径: /etc/ssl/private/server.key
[INFO] 执行重载命令: sudo systemctl reload nginx
[SUCCESS] 部署流程完成
```

## 错误处理

### 常见错误及解决方案

1. **连接超时**
   - 检查网络连通性
   - 确认防火墙规则
   - 调整超时时间

2. **认证失败**
   - 验证用户名和密码
   - 检查私钥格式
   - 确认公钥已添加到服务器

3. **权限错误**
   - 检查目标路径权限
   - 确认用户有足够权限
   - 使用sudo命令

## 测试

运行测试：
```bash
go test -v ./utils/deployment/providers/ssh
```

注意：集成测试需要真实的SSH环境，默认会被跳过。