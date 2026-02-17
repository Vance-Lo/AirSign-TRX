# 🛡️ 安全校验指南 (SHA256)

为了确保您的资金安全，请在运行程序前对比文件指纹：

### 💻 Windows (PowerShell)
> CertUtil -hashfile AirSign-TRX_Windows_x64.zip SHA256

### 🍎 macOS (Terminal)
> shasum -a 256 AirSign-TRX_macOS_Apple.tar.gz

### 🐧 Linux
> sha256sum AirSign-TRX_Linux_x64.tar.gz

**核对方法**: 将输出的哈希字符串与 `SHA256SUMS` 文件中的内容对比，完全一致方可使用。