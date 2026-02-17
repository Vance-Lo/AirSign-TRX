#!/bin/bash

# =================================================================
# Project: AirSign-TRX Auto Builder (Pro Version)
# Author: Vance Lo | Date: 2026-1-17
# =================================================================

PROJECT_NAME="AirSign-TRX"
VERSION="v1.0.0"
OUTPUT_DIR="bin"
APPS=("bridge" "vault")
PLATFORMS=("windows/amd64" "linux/amd64" "darwin/amd64" "darwin/arm64")

echo "🚀 开始构建 $PROJECT_NAME $VERSION ..."
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 1. 静态编译逻辑 [cite: 2026-01-16, 2026-02-17]
for APP in "${APPS[@]}"; do
    for PLATFORM in "${PLATFORMS[@]}"; do
        OS=${PLATFORM%/*}
        ARCH=${PLATFORM#*/}
        BINARY_NAME="${APP}_${OS}_${ARCH}"
        [ "$OS" == "windows" ] && BINARY_NAME="${BINARY_NAME}.exe"
        CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -o "${OUTPUT_DIR}/${BINARY_NAME}" "./cmd/${APP}"
    done
done

# 2. 自动化打包 (集成 SOP 和校验指南) [cite: 2026-02-17]
echo "🎁 正在打包并集成说明文档..."
cp README.md SOP.md VERIFY.md LICENSE $OUTPUT_DIR/
cd $OUTPUT_DIR

# 针对 Windows 打包
zip -q "${PROJECT_NAME}_${VERSION}_Windows_x64.zip" bridge_windows_amd64.exe vault_windows_amd64.exe SOP.md VERIFY.md LICENSE
rm bridge_windows_amd64.exe vault_windows_amd64.exe

# 针对 Linux 打包
tar -czf "${PROJECT_NAME}_${VERSION}_Linux_x64.tar.gz" bridge_linux_amd64 vault_linux_amd64 SOP.md VERIFY.md LICENSE
rm bridge_linux_amd64 vault_linux_amd64

# 针对 macOS 打包
tar -czf "${PROJECT_NAME}_${VERSION}_macOS_Intel.tar.gz" bridge_darwin_amd64 vault_darwin_amd64 SOP.md VERIFY.md LICENSE
tar -czf "${PROJECT_NAME}_${VERSION}_macOS_Apple.tar.gz" bridge_darwin_arm64 vault_darwin_arm64 SOP.md VERIFY.md LICENSE
rm bridge_darwin_amd64 vault_darwin_amd64 bridge_darwin_arm64 vault_darwin_arm64

# 3. 生成外置数字指纹 (用于下载前校验) [cite: 2026-02-17]
shasum -a 256 *.zip *.tar.gz > SHA256SUMS

# 清理临时拷贝的 MD 文件
rm README.md SOP.md VERIFY.md LICENSE
cd ..

echo "-------------------------------------------------"
echo "✅ 构建完成！文档已集成至所有压缩包中。"
ls -lh $OUTPUT_DIR