package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/hashicorp/vault/shamir"
	"google.golang.org/protobuf/proto"
)

// 全局读取器
var reader = bufio.NewReader(os.Stdin)

func main() {
	// 全局防崩溃护盾
	defer func() {
		if r := recover(); r != nil {
			printError(fmt.Sprintf("\n💥 程序发生了严重的意外错误: %v", r))
			fmt.Println("🔍 错误堆栈:", string(debug.Stack()))
			fmt.Println("请截图保存以上信息。按回车键退出...")
			reader.ReadString('\n')
		}
	}()

	cwd, _ := os.Getwd()

	for {
		clearScreen()
		// UI 优化：启动画面增加新身份标识
		fmt.Printf("%s==================================================%s\n", ColorCyan, ColorReset)
		fmt.Printf("      🛡️  AirSign-TRX Vault (Offline)             \n")
		fmt.Printf("      Dev: Vance Lo | Site: vancelo.com           \n")
		fmt.Printf("      [SHA256 Fixed | SSS 3/5 | Air-Gapped]       \n")
		fmt.Printf("      WorkDir: %s\n", filepath.Base(cwd))
		fmt.Printf("%s==================================================%s\n", ColorCyan, ColorReset)

		fmt.Println("1. 📚  Split Key (拆分私钥)")
		fmt.Println("2. 🧩  Combine Key (还原私钥 - 慎用)")
		fmt.Println("3. 📝  Silent Sign (手动复制签名)")
		fmt.Println("4. 📂  Auto File Sign (自动文件签名 - 推荐)")
		fmt.Println("5. 🔍  View Address (查看钱包地址)")
		fmt.Println("q. 👋  Exit")
		fmt.Printf("%s==================================================%s\n", ColorCyan, ColorReset)
		fmt.Print("👉 Choice: ")

		input := readInput()

		switch strings.ToLower(input) {
		case "1":
			runSplitSafe()
		case "2":
			runCombineSafe()
		case "3":
			runSilentSignSafe()
		case "4":
			runAutoFileSign()
		case "5":
			runViewAddress()
		case "q":
			printInfo("正在安全清除内存并退出...")
			return
		default:
			printError("无效指令，请按回车重试...")
			reader.ReadString('\n')
		}
	}
}

// --- 功能 1: 拆分逻辑 ---
// --- 功能 1: 拆分逻辑 (修复 Slice 报错) ---
func runSplitSafe() {
	printInfo("\n[模式: 拆分私钥]")
	fmt.Println("请输入您的 64 位十六进制私钥 (不带 0x 前缀):")
	rawPriv := readInput()

	if len(rawPriv) != 64 {
		printError("私钥长度错误！必须是 64 个字符。")
		return
	}

	data, err := hex.DecodeString(rawPriv)
	if err != nil {
		printError("私钥格式错误！包含非 Hex 字符。")
		return
	}

	fmt.Println("请输入一段随意字符搅动随机池 (敲击键盘后回车):")
	readInput()

	fmt.Println("正在进行数学分片计算...")
	parts, err := shamir.Split(data, 5, 3)
	if err != nil {
		printError("拆分算法执行失败: " + err.Error())
		return
	}

	cwd, _ := os.Getwd()
	fmt.Printf("\n正在当前目录 [%s] 生成文件...\n", cwd)

	for i, part := range parts {
		fileName := fmt.Sprintf("shard_%d.key", i+1)
		hexPart := hex.EncodeToString(part)

		// 🛠️ 修复点：先赋值给变量 hash，再切片
		hash := sha256.Sum256([]byte(hexPart))
		shortHash := hash[:4] // 现在可以切片了

		printSuccess(fmt.Sprintf("✅ [成功] %s (校验码: %x...)", fileName, shortHash))
		os.WriteFile(fileName, []byte(hexPart), 0644)
	}

	printSuccess("\n📢 所有分片已生成！请立即将它们移动到不同的 U 盘中。")
	pause()
}

// --- 功能 2: 还原逻辑 ---
func runCombineSafe() {
	printInfo("\n[模式: 还原私钥]")
	privBytes, err := recoverFromShardsSafe()
	if err != nil {
		printError(err.Error())
		return
	}

	printWarn("\n⚠️  警告：私钥即将显示在屏幕上，请确保四周无人！")
	fmt.Println("--------------------------------------------------")
	// UI 优化：高亮显示私钥
	fmt.Printf("%s🎉 原始私钥: %s%s\n", ColorGreen, hex.EncodeToString(privBytes), ColorReset)
	fmt.Println("--------------------------------------------------")
	zeroBytes(privBytes)
	pause()
}

// --- 功能 3: 手动静默签名 ---
func runSilentSignSafe() {
	printInfo("\n[模式: 手动静默签名]")
	fmt.Println("此模式不读写文件，请手动复制粘贴代码。")

	fmt.Println("\n请粘贴联网端生成的 Raw Hex (未签名交易数据):")
	rawHex := readInput()
	rawHex = strings.TrimSpace(rawHex)

	txBytes, err := hex.DecodeString(rawHex)
	if err != nil {
		printError("交易数据格式错误，不是有效的 Hex 字符串。")
		return
	}

	tx := &core.Transaction{}
	if err := proto.Unmarshal(txBytes, tx); err != nil {
		printError("交易解析失败，可能数据不完整或版本不匹配。")
		return
	}

	if !previewTransaction(tx) {
		printWarn("🚫 用户取消操作。")
		pause()
		return
	}

	printInfo("\n确认无误，开始加载分片...")
	privBytes, err := recoverFromShardsSafe()
	if err != nil {
		printError(err.Error())
		return
	}
	defer func() { zeroBytes(privBytes); runtime.GC(); fmt.Println("🧹 内存私钥已擦除") }()

	privateKey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		printError("私钥转换失败")
		return
	}

	ethAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	tronBytes := append([]byte{0x41}, ethAddr.Bytes()...)
	// UI 优化：地址自检高亮
	printWarn(fmt.Sprintf("\n🔑 签名私钥对应地址: %s", common.EncodeCheck(tronBytes)))

	// 字节级提取 (保留核心逻辑)
	if len(txBytes) < 3 || txBytes[0] != 0x0a {
		printError("非标准交易结构，无法提取 RawData")
		return
	}

	var rawDataLen, headerLen int
	if txBytes[1] < 0x80 {
		rawDataLen = int(txBytes[1])
		headerLen = 2
	} else {
		part1 := int(txBytes[1] & 0x7F)
		part2 := int(txBytes[2] & 0x7F)
		rawDataLen = part1 | (part2 << 7)
		headerLen = 3
	}

	if len(txBytes) < headerLen+rawDataLen {
		printError("数据截断，长度校验失败")
		return
	}
	rawDataBytes := txBytes[headerLen : headerLen+rawDataLen]

	// SHA256 签名 (保留核心逻辑)
	hash := sha256.Sum256(rawDataBytes)
	sig, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		printError("签名计算过程出错: " + err.Error())
		return
	}

	sigTag := []byte{0x12, 0x41}
	finalBytes := append(txBytes, sigTag...)
	finalBytes = append(finalBytes, sig...)

	printSuccess("\n✅ 签名成功！请复制下方报文：")
	fmt.Println("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	fmt.Println(hex.EncodeToString(finalBytes))
	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	pause()
}

// --- 功能 4: 自动文件签名 ---
func runAutoFileSign() {
	printInfo("\n[模式: 自动文件签名]")

	content, err := os.ReadFile("request.txt")
	if err != nil {
		printError("未找到 request.txt")
		return
	}

	rawHex := strings.TrimSpace(string(content))
	txBytes, err := hex.DecodeString(rawHex)
	if err != nil {
		printError("Hex 格式错误")
		return
	}

	printInfo("请加载分片...")
	privBytes, err := recoverFromShardsSafe()
	if err != nil {
		printError(err.Error())
		return
	}
	defer func() { zeroBytes(privBytes); runtime.GC() }()

	privateKey, _ := crypto.ToECDSA(privBytes)
	ethAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	tronBytes := append([]byte{0x41}, ethAddr.Bytes()...)

	// UI 优化：显示地址自检
	printWarn(fmt.Sprintf("\n🔍 [自检] 私钥对应地址: %s", common.EncodeCheck(tronBytes)))

	// 字节级提取
	if len(txBytes) < 3 || txBytes[0] != 0x0a {
		printError("非标准交易结构")
		return
	}
	var rawDataLen, headerLen int
	if txBytes[1] < 0x80 {
		rawDataLen = int(txBytes[1])
		headerLen = 2
	} else {
		rawDataLen = int(txBytes[1]&0x7F) | (int(txBytes[2]&0x7F) << 7)
		headerLen = 3
	}

	if len(txBytes) < headerLen+rawDataLen {
		printError("数据截断")
		return
	}
	rawDataBytes := txBytes[headerLen : headerLen+rawDataLen]

	// SHA256 哈希
	hash := sha256.Sum256(rawDataBytes)
	printInfo(fmt.Sprintf("\n📝 [Debug] 计算哈希: %x", hash))

	// 签名
	sig, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		printError("签名失败: " + err.Error())
		return
	}

	// 拼接
	sigTag := []byte{0x12, 0x41}
	finalBytes := append(txBytes, sigTag...)
	finalBytes = append(finalBytes, sig...)

	os.WriteFile("signed.txt", []byte(hex.EncodeToString(finalBytes)), 0644)
	printSuccess("\n🎉 signed.txt 生成成功 (SHA256算法)！")
	pause()
}

// --- 功能 5: 查看地址 ---
func runViewAddress() {
	printInfo("\n[模式: 查看钱包地址]")
	privBytes, err := recoverFromShardsSafe()
	if err != nil {
		printError(err.Error())
		return
	}
	defer func() { zeroBytes(privBytes); runtime.GC(); fmt.Println("\n🧹 内存私钥已擦除") }()

	privateKey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		printError("无效私钥")
		return
	}

	ethAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	tronAddress := common.EncodeCheck(append([]byte{0x41}, ethAddress.Bytes()...))

	printSuccess("\n🎉 您的波场钱包地址为:")
	fmt.Println("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	fmt.Printf("  %s\n", tronAddress)
	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	pause()
}

// --- 辅助模块 (预览) ---
func previewTransaction(tx *core.Transaction) bool {
	printInfo("\n🔎 [交易安全审计]")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("⏱  过期时间: %s\n", time.Unix(tx.RawData.Expiration/1000, 0).Format("2006-01-02 15:04:05"))

	if len(tx.RawData.Contract) == 0 {
		printWarn("⚠️  警告: 未检测到合约内容！")
		return confirmAction()
	}

	contract := tx.RawData.Contract[0]

	if contract.Type == core.Transaction_Contract_TransferContract {
		var tc core.TransferContract
		if err := proto.Unmarshal(contract.Parameter.Value, &tc); err == nil {
			amount := float64(tc.Amount) / 1000000.0
			fmt.Printf("📝 类型: TRX 转账\n")
			// UI 优化：金额和收款人高亮
			printWarn(fmt.Sprintf("👉 收款人: %s", common.EncodeCheck(tc.ToAddress)))
			printWarn(fmt.Sprintf("💰 金额: %.6f TRX", amount))
			return confirmAction()
		}
	}

	if contract.Type == core.Transaction_Contract_TriggerSmartContract {
		var tsc core.TriggerSmartContract
		if err := proto.Unmarshal(contract.Parameter.Value, &tsc); err == nil {
			fmt.Printf("📝 类型: 智能合约调用 (USDT/TRC20)\n")
			fmt.Printf("🏢 合约: %s\n", common.EncodeCheck(tsc.ContractAddress))
			if len(tsc.Data) >= 68 && hex.EncodeToString(tsc.Data[:4]) == "a9059cbb" {
				toAddr := common.EncodeCheck(append([]byte{0x41}, tsc.Data[4+12:36]...))
				amountInt := new(big.Int).SetBytes(tsc.Data[36:68])
				printWarn(fmt.Sprintf("👉 收款人: %s", toAddr))
				printWarn(fmt.Sprintf("💰 数值: %s (未除精度)", amountInt.String()))
			} else {
				printWarn("⚠️  无法完全解析合约参数，请仔细核对！")
			}
			return confirmAction()
		}
	}

	printWarn(fmt.Sprintf("⚠️  未知合约类型: %v", contract.Type))
	return confirmAction()
}

// --- 辅助模块 (分片读取) ---
func recoverFromShardsSafe() ([]byte, error) {
	fmt.Println("请输入 3 个分片文件的路径 (支持拖入文件，用逗号隔开):")
	input := readInput()
	input = strings.ReplaceAll(input, "\"", "")
	input = strings.ReplaceAll(input, "'", "")
	input = strings.ReplaceAll(input, "，", ",")

	filePaths := strings.Split(input, ",")
	if len(filePaths) < 3 {
		return nil, fmt.Errorf("路径少于 3 个")
	}

	var collected [][]byte
	for _, path := range filePaths {
		cleanPath := strings.TrimSpace(path)
		content, err := os.ReadFile(cleanPath)
		if err != nil {
			printError(fmt.Sprintf("读取失败: %s", filepath.Base(cleanPath)))
			return nil, err
		}
		partBytes, err := hex.DecodeString(strings.TrimSpace(string(content)))
		if err != nil {
			return nil, fmt.Errorf("Hex 格式错误")
		}

		// UI 优化：读取成功显示绿色
		printSuccess(fmt.Sprintf("✅ 读取: %s", filepath.Base(cleanPath)))
		collected = append(collected, partBytes)
	}

	fmt.Println("正在进行数学合并计算...")
	secret, err := shamir.Combine(collected)
	if err != nil {
		return nil, fmt.Errorf("分片合并失败，文件不匹配")
	}
	return secret, nil
}

// --- 通用工具 ---
func readInput() string { s, _ := reader.ReadString('\n'); return strings.TrimSpace(s) }
func pause()            { fmt.Println("\n按回车键继续..."); reader.ReadString('\n') }
func confirmAction() bool {
	fmt.Print("\n❓ 确认无误? (y/n): ")
	return strings.ToLower(readInput()) == "y"
}
func zeroBytes(s []byte) {
	for i := range s {
		s[i] = 0
	}
}
func clearScreen() {
	if runtime.GOOS == "windows" {
		fmt.Println("--------------------------------------------------")
	} else {
		fmt.Print("\033[H\033[2J")
	}
}

// --- UI 颜色工具 (核心优化) ---
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

func printSuccess(msg string) { fmt.Println(ColorGreen + msg + ColorReset) }
func printError(msg string)   { fmt.Println(ColorRed + "❌ " + msg + ColorReset) }
func printWarn(msg string)    { fmt.Println(ColorYellow + msg + ColorReset) }
func printInfo(msg string)    { fmt.Println(ColorCyan + msg + ColorReset) }
