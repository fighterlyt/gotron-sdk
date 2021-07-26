package transaction

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fighterlyt/gotron-sdk/pkg/address"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/keystore"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func TestAllow(t *testing.T) {
	g := client.NewGrpcClient(`grpc.shasta.trongrid.io:50051`)

	if err := g.Start(grpc.WithInsecure()); err != nil {
		panic(err.Error())
	}

	if amount, err := g.TRC20Allow(`TYjBaCYBgngDA3nMpBD76Qk7qBx8twvDqY`, `TEUmz9RLVXCBBg3ohoTWmN7dRDiAoXZygy`, `TRZTJyNpKVevp959982XBbkjm7qrLxYTWi`); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Log(amount.Uint64())
	}
}

func TestApprove(t *testing.T) {
	g := client.NewGrpcClient(`grpc.shasta.trongrid.io:50051`)

	if err := g.Start(grpc.WithInsecure()); err != nil {
		panic(err.Error())
	}

	if err := Approve(g, `TYjBaCYBgngDA3nMpBD76Qk7qBx8twvDqY`, `46a630a7169cd0f1a739f8ca6fb14ddf95717a82a97044c6c71773bcb898507f`, `TEUmz9RLVXCBBg3ohoTWmN7dRDiAoXZygy`,
		`TRZTJyNpKVevp959982XBbkjm7qrLxYTWi`, big.NewInt(100)); err != nil {
		t.Fatal(err.Error())
	}
}

func TestSendFrom(t *testing.T) {
	g := client.NewGrpcClient(`grpc.shasta.trongrid.io:50051`)

	if err := g.Start(grpc.WithInsecure()); err != nil {
		panic(err.Error())
	}

	if err := SendFrom(g, `TEUmz9RLVXCBBg3ohoTWmN7dRDiAoXZygy`, `TYjBaCYBgngDA3nMpBD76Qk7qBx8twvDqY`, `2fecd7b903137647917cc6dbbbbb92efbeb2a34d987a9dcbc5308022a3f0fc14`,
		`TEUmz9RLVXCBBg3ohoTWmN7dRDiAoXZygy`,
		`TRZTJyNpKVevp959982XBbkjm7qrLxYTWi`, big.NewInt(100)); err != nil {
		t.Fatal(err.Error())
	}
}

func Approve(client *client.GrpcClient, from, privateKeyHex, to string, contract string, amount *big.Int) error { //nolint:golint,lll
	var privateKey *ecdsa.PrivateKey

	var err error

	var (
		account keystore.Account
	)

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return errors.Wrapf(err, "解析私钥错误")
	}

	var fromAddress = address.PubkeyToAddress(privateKey.PublicKey)

	k := keystore.NewKeyStore("keystore", keystore.LightScryptN, keystore.LightScryptP)

	if k.HasAddress(fromAddress) {
		if account, err = k.Find(keystore.Account{Address: fromAddress}); err != nil {
			return errors.Wrap(err, "加载账号")
		}
	} else {
		if account, err = k.ImportECDSA(privateKey, ""); err != nil {
			return errors.Wrap(err, "导入私钥")
		}
	}

	if err = k.Unlock(account, ""); err != nil {
		return errors.Wrap(err, "unlock钱包错误")
	}

	var tx *api.TransactionExtention

	if tx, err = client.TRC20Approve(from, to, contract, amount, 100000000); err != nil { //nolint:golint,lll
		return errors.Wrap(err, "构建交易失败")
	}

	controller := NewController(client, k, &account, tx.Transaction)

	if err = controller.ExecuteTransaction(); err != nil {
		return err
	}

	return nil
}

func SendFrom(client *client.GrpcClient, from, spender, privateKeyHex, to string, contract string, amount *big.Int) error { //nolint:golint,lll
	var privateKey *ecdsa.PrivateKey

	var err error

	var (
		account keystore.Account
	)

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return errors.Wrapf(err, "解析私钥错误")
	}

	var fromAddress = address.PubkeyToAddress(privateKey.PublicKey)

	k := keystore.NewKeyStore("keystore", keystore.LightScryptN, keystore.LightScryptP)

	if k.HasAddress(fromAddress) {
		if account, err = k.Find(keystore.Account{Address: fromAddress}); err != nil {
			return errors.Wrap(err, "加载账号")
		}
	} else {
		if account, err = k.ImportECDSA(privateKey, ""); err != nil {
			return errors.Wrap(err, "导入私钥")
		}
	}

	if err = k.Unlock(account, ""); err != nil {
		return errors.Wrap(err, "unlock钱包错误")
	}

	var tx *api.TransactionExtention

	if tx, err = client.TRC20SendFrom(from, spender, to, contract, amount, 100000000); err != nil { //nolint:golint,lll
		return errors.Wrap(err, "构建交易失败")
	}

	controller := NewController(client, k, &account, tx.Transaction)

	if err = controller.ExecuteTransaction(); err != nil {
		return err
	}

	return nil
}
