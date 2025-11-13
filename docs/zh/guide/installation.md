# 安装指南

本指南涵盖在不同平台上安装 BobaMixer 的所有方法。

## 系统要求

- **操作系统**: macOS、Linux 或 Windows (通过 WSL)
- **Go 版本**: 1.22+ (如果从源代码构建)
- **SQLite**: 3.x (通常预装)
- **磁盘空间**: 最少 ~50MB (加上数据库增长)
- **Git**: 可选,用于 git hooks 集成

## 安装方法

### 方法 1: 使用 Go Install (推荐)

如果你已安装 Go,这是最简单的方法:

```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

这会将 `boba` 二进制文件安装到 `$GOPATH/bin`。确保此目录在你的 PATH 中:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

将此行添加到你的 shell 配置文件 (`~/.bashrc`、`~/.zshrc` 等) 以使其永久生效。

### 方法 2: 使用 Homebrew (macOS/Linux)

对于 macOS 和 Linux 用户,Homebrew 提供简便的安装:

```bash
# 添加 BobaMixer tap
brew tap royisme/tap

# 安装 BobaMixer
brew install bobamixer
```

更新到最新版本:

```bash
brew upgrade bobamixer
```

### 方法 3: 下载预编译二进制文件

从[发布页面](https://github.com/royisme/BobaMixer/releases)下载适合你平台的二进制文件。

**Linux (amd64):**
```bash
# 下载
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Linux_x86_64.tar.gz

# 解压
tar -xzf bobamixer_Linux_x86_64.tar.gz

# 移动到 PATH
sudo mv boba /usr/local/bin/

# 验证
boba version
```

**macOS (Apple Silicon):**
```bash
# 下载
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Darwin_arm64.tar.gz

# 解压
tar -xzf bobamixer_Darwin_arm64.tar.gz

# 移动到 PATH
sudo mv boba /usr/local/bin/

# 验证
boba version
```

**macOS (Intel):**
```bash
# 下载
wget https://github.com/royisme/BobaMixer/releases/latest/download/bobamixer_Darwin_x86_64.tar.gz

# 解压
tar -xzf bobamixer_Darwin_x86_64.tar.gz

# 移动到 PATH
sudo mv boba /usr/local/bin/

# 验证
boba version
```

### 方法 4: 从源代码构建

对于开发者或如果你想要最新的开发版本:

```bash
# 克隆仓库
git clone https://github.com/royisme/BobaMixer.git
cd BobaMixer

# 构建
make build

# 安装到 GOPATH/bin
make install

# 或手动移动二进制文件
sudo mv dist/boba /usr/local/bin/
```

**开发设置:**

如果你计划贡献代码,设置开发环境:

```bash
# 安装依赖和 git hooks
make dev

# 运行测试
make test

# 运行 linter
make lint
```

详见[贡献指南](https://github.com/royisme/BobaMixer/blob/main/CONTRIBUTING.md)。

## 安装后设置

### 1. 创建配置目录

安装后,创建 BobaMixer 主目录:

```bash
mkdir -p ~/.boba/logs
chmod 700 ~/.boba
```

### 2. 初始化配置

运行 doctor 命令生成示例配置:

```bash
boba doctor
```

这会创建:
- `~/.boba/profiles.yaml` - 配置文件定义
- `~/.boba/routes.yaml` - 路由规则
- `~/.boba/pricing.yaml` - 模型定价信息
- `~/.boba/secrets.yaml` - API 密钥存储
- `~/.boba/usage.db` - SQLite 数据库

### 3. 保护 Secrets 文件

确保你的 secrets 文件具有适当的权限:

```bash
chmod 600 ~/.boba/secrets.yaml
```

如果 `secrets.yaml` 权限不正确,BobaMixer 将拒绝运行。

### 4. 验证安装

检查一切是否正常:

```bash
# 检查版本
boba version

# 运行健康检查
boba doctor

# 列出默认配置文件
boba ls --profiles
```

## 平台特定说明

### macOS

在 macOS 上首次运行 BobaMixer 时,你可能会看到安全警告。解决方法:

1. 尝试运行 `boba`
2. 打开**系统偏好设置** → **安全性与隐私**
3. 点击 BobaMixer 警告旁边的**仍要打开**
4. 再次运行 `boba`

或者,移除隔离属性:

```bash
xattr -d com.apple.quarantine /usr/local/bin/boba
```

### Linux

在某些 Linux 发行版上,你可能需要安装 SQLite:

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install sqlite3 libsqlite3-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install sqlite sqlite-devel
```

**Arch Linux:**
```bash
sudo pacman -S sqlite
```

### Windows (WSL)

BobaMixer 通过 WSL (Windows Subsystem for Linux) 在 Windows 上运行:

1. 安装 WSL 2:
   ```powershell
   wsl --install
   ```

2. 安装 Ubuntu 或你喜欢的发行版

3. 在 WSL 中遵循 Linux 安装说明

## 升级

### 使用 Homebrew

```bash
brew upgrade bobamixer
```

### 使用 Go Install

```bash
go install github.com/royisme/bobamixer/cmd/boba@latest
```

### 手动升级

下载最新二进制文件并替换现有文件:

```bash
# 备份当前版本
cp /usr/local/bin/boba /usr/local/bin/boba.backup

# 下载并安装新版本
# (遵循上述下载二进制文件说明)

# 验证
boba version
```

### 迁移说明

升级时,如果需要,BobaMixer 会自动迁移数据库架构。升级前始终备份数据:

```bash
# 备份配置和数据库
cp -r ~/.boba ~/.boba.backup.$(date +%Y%m%d)

# 升级 BobaMixer
# ... 遵循升级说明 ...

# 验证
boba doctor
```

## 卸载

### 使用 Homebrew

```bash
brew uninstall bobamixer
brew untap royisme/tap
```

### 手动卸载

```bash
# 移除二进制文件
sudo rm /usr/local/bin/boba

# 可选:移除配置 (这会删除所有数据!)
rm -rf ~/.boba
```

要保留数据以便将来重新安装,不要删除 `~/.boba`。

## 故障排除安装

### 找不到命令

如果安装后出现 "command not found":

1. 检查二进制文件是否在 PATH 中:
   ```bash
   which boba
   ```

2. 如果为空,添加到 PATH:
   ```bash
   export PATH=$PATH:/usr/local/bin
   ```

3. 通过添加到 shell 配置文件使其永久生效

### 权限被拒绝

如果出现 "permission denied":

```bash
# 使其可执行
chmod +x /usr/local/bin/boba

# 或使用 sudo 重新安装
sudo cp boba /usr/local/bin/
```

### SQLite 问题

如果你看到数据库错误:

```bash
# 检查 SQLite 版本
sqlite3 --version

# 应该是 3.x 或更高
# 如果不是,安装/升级 SQLite
```

### Git Hooks 不工作

如果 git hooks 安装失败:

```bash
# 检查 git 版本
git --version

# 手动安装 hooks
cd your-project
cp ~/.boba/git-hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

## 下一步

安装后,继续:

- **[快速开始](/zh/guide/getting-started)** - 配置你的第一个配置文件
- **[配置指南](/zh/guide/configuration)** - 了解所有配置选项
- **[CLI 参考](/zh/reference/cli)** - 探索所有可用命令

## 获取帮助

如果遇到安装问题:

1. 查看[故障排除指南](/zh/advanced/troubleshooting)
2. 搜索 [GitHub Issues](https://github.com/royisme/BobaMixer/issues)
3. 提交新问题,包括:
   - 你的操作系统和版本
   - 使用的安装方法
   - 错误消息
   - `boba doctor` 的输出
