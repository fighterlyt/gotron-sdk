package free

import (
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fighterlyt/gotron-sdk/pkg/address"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/client/transaction"
	"github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/fighterlyt/gotron-sdk/pkg/keystore"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/api"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/core"
	"github.com/shopspring/decimal"

	"github.com/pkg/errors"
)

/*UnFreeze 解冻质押TRX,解冻只能一次性解冻
参数:
*	client       	*client.GrpcClient	grc 客户端
*	from         	string            	质押来源钱包地址
*	privateKeyHex	string            	质押来源钱包私钥
*	to           	string              质押收益钱包
*	resource     	core.ResourceCode 	质押资源
返回值:
*	error        	error             	错误
*/
func UnFreeze(client *client.GrpcClient, from, privateKeyHex, to string, resource core.ResourceCode) (txID string, err error) { //nolint:lll
	var (
		account    keystore.Account
		privateKey *ecdsa.PrivateKey
		tx         *api.TransactionExtention
	)

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return ``, errors.Wrapf(err, "解析私钥错误")
	}

	fromAddress := address.PubkeyToAddress(privateKey.PublicKey)

	k := keystore.NewKeyStore("keystore", keystore.LightScryptN, keystore.LightScryptP)

	if k.HasAddress(fromAddress) {
		if account, err = k.Find(keystore.Account{Address: fromAddress}); err != nil {
			return ``, errors.Wrap(err, "加载账号")
		}
	} else {
		if account, err = k.ImportECDSA(privateKey, ""); err != nil {
			return ``, errors.Wrap(err, "导入私钥")
		}
	}

	if err = k.Unlock(account, ""); err != nil {
		return ``, errors.Wrap(err, "unlock钱包错误")
	}

	if tx, err = client.UnfreezeBalance(from, to, resource); err != nil {
		return ``, errors.Wrap(err, `操作`)
	}

	controller := transaction.NewController(client, k, &account, tx.Transaction)

	if err = controller.ExecuteTransaction(); err != nil {
		return ``, err
	}

	return strings.TrimPrefix(common.BytesToHexString(tx.GetTxid()), "0x"), nil
}

/*Freeze 冻结质押TRX
参数:
*	client       	*client.GrpcClient	grc 客户端
*	from         	string            	质押来源钱包地址
*	privateKeyHex	string            	质押来源钱包私钥
*	to           	string              质押收益钱包
*	resource     	core.ResourceCode 	质押资源
*	frozenBalance	int64             	质押金额,除以10^6是真正的金额
返回值:
*	error        	error             	错误
*/
func Freeze(client *client.GrpcClient, from, privateKeyHex, to string, resource core.ResourceCode, frozenBalance int64) (txID string, err error) { //nolint:lll
	var (
		account    keystore.Account
		privateKey *ecdsa.PrivateKey
		tx         *api.TransactionExtention
	)

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return ``, errors.Wrapf(err, "解析私钥错误")
	}

	fromAddress := address.PubkeyToAddress(privateKey.PublicKey)

	k := keystore.NewKeyStore("keystore", keystore.LightScryptN, keystore.LightScryptP)

	if k.HasAddress(fromAddress) {
		if account, err = k.Find(keystore.Account{Address: fromAddress}); err != nil {
			return ``, errors.Wrap(err, "加载账号")
		}
	} else {
		if account, err = k.ImportECDSA(privateKey, ""); err != nil {
			return ``, errors.Wrap(err, "导入私钥")
		}
	}

	if err = k.Unlock(account, ""); err != nil {
		return ``, errors.Wrap(err, "unlock钱包错误")
	}

	if tx, err = client.FreezeBalance(from, to, resource, frozenBalance); err != nil {
		return ``, errors.Wrap(err, `操作`)
	}

	controller := transaction.NewController(client, k, &account, tx.Transaction)

	if err = controller.ExecuteTransaction(); err != nil {
		return ``, err
	}

	return strings.TrimPrefix(common.BytesToHexString(tx.GetTxid()), "0x"), nil
}

/*Resource 账号资源
参数:
*	client   	*client.GrpcClient	客户端
*	address  	string            	地址
返回值:
*	bandwidth	int64             	带宽
*	energy   	int64             	能量
*	err      	error             	错误
*/
func Resource(client *client.GrpcClient, address string) (bandwidth, energy int64, err error) {
	var (
		message *api.AccountResourceMessage
	)

	if message, err = client.GetAccountResource(address); err != nil {
		return 0, 0, errors.Wrap(err, `获取资源信息`)
	}

	bandwidth = message.NetLimit + message.FreeNetLimit // 免费+质押获得
	bandwidth -= message.FreeNetUsed                    // 免费已使用
	bandwidth -= message.NetUsed                        // 质押已使用

	return bandwidth, message.EnergyLimit - message.EnergyUsed, nil
}

// DelegatedResource 代理/托管的资源
type DelegatedResource struct {
	From                      string          `json:"from,omitempty"`                         // 来源，质押TRX的一方
	To                        string          `json:"to,omitempty"`                           // 获利的一方
	FrozenBalanceForBandwidth decimal.Decimal `json:"frozen_balance_for_bandwidth,omitempty"` // 用于获取带宽的冻结TRX
	FrozenBalanceForEnergy    decimal.Decimal `json:"frozen_balance_for_energy,omitempty"`    // 用于获取能量的冻结TRX
	ExpireTimeForBandwidth    time.Time       `json:"expire_time_for_bandwidth,omitempty"`    // 带宽过期时间
	ExpireTimeForEnergy       time.Time       `json:"expire_time_for_energy,omitempty"`       // 能量过期时间
}

func NewDelegatedResource(from, to []byte, frozenBalanceForBandwidth, frozenBalanceForEnergy, expireTimeForBandwidth, expireTimeForEnergy int64) *DelegatedResource {
	return &DelegatedResource{
		From:                      Base58ToAddress(common.BytesToHexString(from)),
		To:                        Base58ToAddress(common.BytesToHexString(to)),
		FrozenBalanceForBandwidth: decimal.New(frozenBalanceForBandwidth, -6),
		FrozenBalanceForEnergy:    decimal.New(frozenBalanceForEnergy, -6),
		ExpireTimeForBandwidth:    time.Unix(expireTimeForBandwidth, 0),
		ExpireTimeForEnergy:       time.Unix(expireTimeForEnergy/1000, 0),
	}
}

func FreezeResource(client *client.GrpcClient, address string) (resources []*DelegatedResource, err error) {
	var (
		list []*api.DelegatedResourceList
	)

	if list, err = client.GetDelegatedResources(address); err != nil {
		return nil, errors.Wrap(err, `获取托管资源信息`)
	}

	for _, item := range list {
		for _, resource := range item.GetDelegatedResource() {
			resources = append(resources, NewDelegatedResource(resource.From, resource.To, resource.FrozenBalanceForBandwidth, resource.FrozenBalanceForEnergy,
				resource.ExpireTimeForBandwidth, resource.ExpireTimeForEnergy))
		}
	}

	return resources, nil
}

func Base58ToAddress(s string) string {
	s = strings.TrimPrefix(s, `0x`)
	addr := address.HexToAddress(s)

	return addr.String()
}
