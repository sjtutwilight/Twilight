package contracts

import (
	"log"

	"twilight-go/internal/config"
	"twilight-go/internal/listener"

	"github.com/ethereum/go-ethereum/common"
)

// ContractManager 管理区块链合约
type ContractManager struct {
	contracts map[common.Address]*listener.Contract
}

// NewContractManager 创建新的合约管理器
func NewContractManager() *ContractManager {
	return &ContractManager{
		contracts: make(map[common.Address]*listener.Contract),
	}
}

// LoadContracts 从部署配置加载合约
func (cm *ContractManager) LoadContracts(deployment *config.DeploymentConfig) {
	// 添加工厂合约
	factoryAddr := common.HexToAddress(deployment.Factory)
	factoryContract := listener.NewContract(factoryAddr, "Factory")
	factoryContract.AddEventType("PairCreated", common.HexToHash("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"))
	cm.contracts[factoryAddr] = factoryContract
	log.Printf("Added factory contract at %s", deployment.Factory)

	// 添加交易对合约
	for _, pair := range deployment.Pairs {
		pairAddr := common.HexToAddress(pair.Address)
		pairContract := listener.NewContract(pairAddr, "Pair")
		pairContract.AddEventType("Swap", common.HexToHash("0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"))
		pairContract.AddEventType("Mint", common.HexToHash("0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"))
		pairContract.AddEventType("Burn", common.HexToHash("0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496"))
		pairContract.AddEventType("Sync", common.HexToHash("0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1"))
		pairContract.AddEventType("Transfer", common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"))
		cm.contracts[pairAddr] = pairContract
		log.Printf("Added pair contract at %s", pair.Address)
	}

	// 添加代币合约
	for _, token := range deployment.Tokens {
		tokenAddr := common.HexToAddress(token.Address)
		tokenContract := listener.NewContract(tokenAddr, "Token")
		tokenContract.AddEventType("Transfer", common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"))
		tokenContract.AddEventType("Approval", common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"))
		cm.contracts[tokenAddr] = tokenContract
		log.Printf("Added token contract at %s", token.Address)
	}
}

// LoadDefaultContract 加载默认合约（从config.yaml）
func (cm *ContractManager) LoadDefaultContract(address common.Address) {
	pairContract := listener.NewContract(address, "TWSwapPair")
	pairContract.AddEventType("Swap", common.HexToHash("0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"))
	pairContract.AddEventType("Mint", common.HexToHash("0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"))
	pairContract.AddEventType("Burn", common.HexToHash("0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496"))
	pairContract.AddEventType("Sync", common.HexToHash("0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1"))
	pairContract.AddEventType("Transfer", common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"))
	cm.contracts[address] = pairContract
	log.Printf("Added default pair contract at %s", address.Hex())
}

// RegisterContractsWithListener 将合约注册到监听器
func (cm *ContractManager) RegisterContractsWithListener(listenerInstance *listener.Listener) {
	for addr, contract := range cm.contracts {
		listenerInstance.AddContract(addr, contract)
	}
}
