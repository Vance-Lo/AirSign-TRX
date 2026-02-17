package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time" // 引入 time 包处理时间

	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

// 波场主网 USDT 合约
const USDT_CONTRACT = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
const TRON_GRID_URL = "grpc.trongrid.io:50051"

var reader = bufio.NewReader(os.Stdin)
var tronClient *client.GrpcClient

func main() {
	// API Key 读取逻辑 (保持不变)
	apiKey := ""
	keyBytes, err := os.ReadFile("apikey.txt")
	if err == nil {
		cleanKey := strings.TrimSpace(string(keyBytes))
		if len(cleanKey) > 10 {
			apiKey = cleanKey
			fmt.Println("🔑 已加载自定义 API Key")
		}
	} else {
		fmt.Println("🌐 未找到 apikey.txt，使用公共节点")
	}

	fmt.Println("正在连接波场主网...")
	tronClient = client.NewGrpcClient(TRON_GRID_URL)
	if apiKey != "" {
		tronClient.SetAPIKey(apiKey)
	}

	err = tronClient.Start(grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("❌ 连接失败:", err)
		return
	}
	fmt.Println("✅ 连接成功！")

	for {
		fmt.Println("\n==================================================")
		fmt.Println("       neoui 联网桥接器 V1.3 (Long-Life)          ")
		fmt.Println("       [修复过期错误 | 12小时有效期]              ")
		fmt.Println("==================================================")
		fmt.Println("1. 💰  查询余额")
		fmt.Println("2. 📝  生成 TRX 订单 (有效期 12h)")
		fmt.Println("3. 💵  生成 USDT 订单 (有效期 12h)")
		fmt.Println("4. 📡  广播签名交易")
		fmt.Println("q. 👋  退出")
		fmt.Println("==================================================")
		fmt.Print("👉 请输入指令: ")

		input := readInput()

		switch input {
		case "1":
			runCheckBalance()
		case "2":
			runCreateTrxOrder()
		case "3":
			runCreateUsdtOrder()
		case "4":
			runBroadcast()
		case "q":
			return
		default:
			fmt.Println("❌ 无效指令")
		}
	}
}

// --- 核心修复工具：延长有效期 ---
func extendExpiration(tx *api.TransactionExtention) {
	// 波场最大允许 24 小时，这里我们设置 12 小时，足够走路往返了
	// Expiration 单位是毫秒
	newExpiration := time.Now().Add(12*time.Hour).UnixNano() / 1e6
	tx.Transaction.RawData.Expiration = newExpiration
	fmt.Println("✅ 订单有效期已延长至 12 小时。")
}

// --- 功能 2: 生成 TRX 订单 ---
func runCreateTrxOrder() {
	fmt.Println("\n[生成 TRX 转账订单]")
	fmt.Print("发款地址 (From): ")
	from := readInput()
	fmt.Print("收款地址 (To): ")
	to := readInput()
	fmt.Print("金额 (TRX): ")
	amt := readInput()

	var f float64
	fmt.Sscanf(amt, "%f", &f)

	tx, err := tronClient.Transfer(from, to, int64(f*1000000))
	if err != nil {
		fmt.Println("❌ 失败:", err)
		return
	}

	// ⚡️ 修复点：调用延时函数
	extendExpiration(tx)

	saveRequestFile(tx.Transaction)
}

// --- 功能 3: 生成 USDT 订单 ---
func runCreateUsdtOrder() {
	fmt.Println("\n[生成 USDT 转账订单]")
	fmt.Print("发款地址 (From): ")
	from := readInput()
	fmt.Print("收款地址 (To): ")
	to := readInput()
	fmt.Print("金额 (USDT): ")
	amountStr := readInput()

	amountFloat := 0.0
	fmt.Sscanf(amountStr, "%f", &amountFloat)
	amountInt := int64(amountFloat * 1000000)

	// 使用 TRC20Send (V1.2修复版逻辑)
	tx, err := tronClient.TRC20Send(from, to, USDT_CONTRACT, big.NewInt(amountInt), 50000000)
	if err != nil {
		fmt.Println("❌ 失败:", err)
		return
	}

	// ⚡️ 修复点：调用延时函数
	extendExpiration(tx)

	saveRequestFile(tx.Transaction)
}

// --- 功能 4: 广播 (保持 V1.2 逻辑) ---
func runBroadcast() {
	fmt.Println("\n[广播签名交易]")
	content, err := os.ReadFile("signed.txt")
	if err != nil {
		fmt.Println("❌ 未找到 signed.txt")
		return
	}

	signedHex := strings.TrimSpace(string(content))
	txBytes, _ := hex.DecodeString(signedHex)

	if len(txBytes) == 0 {
		fmt.Println("❌ 文件为空")
		return
	}

	tx := &core.Transaction{}
	if err := proto.Unmarshal(txBytes, tx); err != nil {
		fmt.Println("❌ 解析失败:", err)
		return
	}
	// [新增] 广播前再次确认哈希
	rawData, _ := proto.Marshal(tx.GetRawData())
	hash := sha256.Sum256(rawData)
	fmt.Printf("📝 [Debug] 待广播哈希: %x\n", hash)

	fmt.Println("正在广播...")
	result, err := tronClient.Broadcast(tx)
	if err != nil {
		fmt.Println("❌ 网络错误:", err)
		return
	}

	if result.Code != 0 {
		// 仍然显示具体错误，以便排查
		fmt.Printf("❌ 广播失败: %s (Code: %d)\n", string(result.Message), result.Code)
	} else {
		rawData, _ := proto.Marshal(tx.GetRawData())
		hash := sha256.Sum256(rawData)
		txid := hex.EncodeToString(hash[:])

		fmt.Println("\n✅ 广播成功！")
		fmt.Println("交易哈希 (TXID):", txid)
	}
	pause()
}

// --- 辅助功能 (保持不变) ---
func runCheckBalance() {
	fmt.Print("\n查询地址: ")
	addr := readInput()
	acc, err := tronClient.GetAccount(addr)
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	fmt.Printf("TRX 余额: %.6f\n", float64(acc.Balance)/1000000.0)
	pause()
}

func saveRequestFile(tx *core.Transaction) {
	bytes, _ := proto.Marshal(tx)
	// 2. [新增] 计算并打印 SHA256 哈希，作为“对暗号”的基准
	rawData, _ := proto.Marshal(tx.GetRawData())
	h := sha256.Sum256(rawData)
	fmt.Printf("\n📝 [Debug] 生成的订单哈希 (SHA256): %x\n", h)
	fmt.Println("👉 请记下前 4 位，去断网电脑核对！")

	os.WriteFile("request.txt", []byte(hex.EncodeToString(bytes)), 0644)
	fmt.Println("🎉 request.txt 已生成！(有效期 12h)")
	pause()
}

func readInput() string {
	str, _ := reader.ReadString('\n')
	return strings.TrimSpace(str)
}
func pause() {
	fmt.Println("\n按回车键继续...")
	reader.ReadString('\n')
}

// --- UI 美化工具 (放在文件最下方) ---
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

// 替换原来的 fmt.Println("✅ ...")
func printSuccess(msg string) {
	fmt.Println(ColorGreen + msg + ColorReset)
}

// 替换原来的 fmt.Println("❌ ...")
func printError(msg string) {
	fmt.Println(ColorRed + msg + ColorReset)
}

// 替换原来的 fmt.Println("⚠️ ...")
func printWarn(msg string) {
	fmt.Println(ColorYellow + msg + ColorReset)
}

// 用于打印哈希值或标题
func printInfo(msg string) {
	fmt.Println(ColorCyan + msg + ColorReset)
}
