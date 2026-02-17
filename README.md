# AirSign-TRX: Air-Gapped Cold Wallet for TRON

> 🛡️ **Zero-Knowledge, Sharded Key, Air-Gapped Transaction Signing.**

![Go Version](https://img.shields.io/badge/Go-1.16%2B-blue) ![License](https://img.shields.io/badge/license-MIT-green) ![Author](https://img.shields.io/badge/dev-Vance%20Lo-orange)

**AirSign-TRX** 是一套极简主义的波场（TRON）冷钱包解决方案。它严格遵循**私钥不触网**原则，利用 Shamir's Secret Sharing (SSS) 技术将私钥粉碎为多份分片，确保资产的极致安全与可继承性。

**Developer**: Vance Lo  
**Contact**: vance.dev@proton.me  
**Website**: [vancelo.com](https://vancelo.com)

---

## 🌟 核心特性 (Features)

* **⚡️ 物理隔离 (Air-Gapped)**: 
    * **Vault (冷端)**: 运行在断网电脑上，负责维护私钥和签名 。
    * **Bridge (热端)**: 运行在联网电脑上，负责生成订单和广播交易 。
* **🧩 SSS 分片技术**: 私钥生成后即被拆分为 5 份（3/5 门限），任何单一物理介质丢失都不会导致资产失窃 。
* **🛡️ 字节级签名**: 采用 SHA256 原始字节提取算法，解决 Protobuf 跨平台序列化哈希不一致问题，杜绝盲签风险 。
* **👁️ 智能反欺诈**: 冷端在签名时自动解析交易内容，直观显示收款地址与金额 。
* **💵 全币种支持**: 原生支持 TRX 转账及 USDT (TRC20) 合约调用 。

---

### 📥 下载编译好的程序 (Recommended)
如果您不想安装 Go 环境，可以直接在 [Releases](https://github.com/Vance-Lo/AirSign-TRX/releases/) 页面下载对应系统的压缩包：
* **Windows**: `AirSign-TRX_Win.zip`
* **macOS**: `AirSign-TRX_Mac.zip` (支持 M1/M2/M3)
* **Linux**: `AirSign-TRX_Linux.tar.gz`

---

## 🚀 快速开始 (Quick Start)

### 1. 编译 (Build)

你需要安装 Go 1.16+ 环境 。

```bash
# 1. 克隆项目
git clone https://github.com/AirSign-TRX/AirSign-TRX.git
cd AirSign-TRX
```
**或者**
```bash
# 1. GitHub CLI 克隆项目
如果您安装了 GitHub 官方命令行工具，只需输入：
gh repo clone Vance-Lo/AirSign-TRX
```

```bash
# 2. 安装并整理依赖 (必须执行) [cite: 2026-02-17]
go mod tidy [cite: 2026-02-17]
```
```bash
# 3. 编译联网端 (Bridge)
go build -o bin/bridge ./cmd/bridge
# 3. 编译断网端 (Vault)
go build -o bin/vault ./cmd/vault
```
### 2. 初始化 (Setup)

1. 准备两个 U 盘：**冷盘 (Key USB)** 和 **热盘 (Transfer USB)**。
2. 将 `bin/vault` 复制到 **断网电脑**。
3. 将 `bin/bridge` 留在 **联网电脑**。

## 📖 使用手册 (SOP)

### 阶段一：创建金库 (仅需一次)

* **启动**: 在 **断网电脑** 运行 `vault`。
* **操作**: 选择 `[1] 拆分私钥`。
* **输入**: 粘贴你的 64位 Hex 私钥 。
* **结果**: 程序生成 5 个 `.key` 分片，请分开放置 。
* **安全**: 完成后建议重启断网电脑清空内存 。

### 阶段二：日常转账 / 资产提取

#### 1. 生成订单 (联网端)
* **运行**: 执行 `bridge` 程序。
* **选择**: 根据需求选 `[2] TRX` 或 `[3] USDT` 转账。
* **输入**: 按照提示填入发款人、收款人及金额。
* **结果**: 生成 `request.txt` (有效期 12 小时) 。
* **动作**: 将 `request.txt` 拷入 **热盘 U 盘**。

#### 2. 离线签名 (断网端)
* **准备**: 插入 **热盘 U 盘** 和 **冷盘 U 盘** (包含至少 3 个分片)。
* **运行**: 执行 `vault`，选择 `[4] 全自动文件签名` 。
* **核对**: 屏幕将解析并显示交易详情，确认无误后输入 `y` 。
* **结果**: 生成已签名的 `signed.txt` 。
* **动作**: 将 `signed.txt` 拷回 **热盘 U 盘**。

#### 3. 广播交易 (联网端)
* **连接**: 将 **热盘 U 盘** 插回联网电脑。
* **执行**: 运行 `bridge`，选择 `[4] 广播交易` 。
* **完成**: 交易上链，资金到账。

---

## 📂 目录结构

```text
AirSign-TRX/
├── cmd/
│   ├── bridge/      # 联网端源码 (Hot Wallet)
│   └── vault/       # 断网端源码 (Cold Vault)
├── go.mod           # 模块依赖管理
├── LICENSE          # MIT 开源协议
├── README.md        # 项目说明文档
├── build.sh         # 自动化脚本 
├── SOP.md           # 离线操作指南 
└── VERIFY.md        # 校验命令速查 
```
---

## ⚠️ 免责声明 (Disclaimer)

> 本软件按“原样”提供，不提供任何形式的明示或暗示保证 。作者不对因使用本软件而导致的资金丢失、私钥泄露或交易失败承担任何责任 。

* 请务必在小额测试通过后再进行大额资产管理 。
* 请务必保管好您的物理分片文件 。
* **严禁在联网设备上运行 Vault 程序** 。

---

Copyright © 2026 Vance Lo. All rights reserved.

