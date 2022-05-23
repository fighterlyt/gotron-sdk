package free

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/core"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	from       = ``
	privateKey = ``
	configFile = `./config.json`
	to         = `TYjBaCYBgngDA3nMpBD76Qk7qBx8twvDqY`
)

type Data struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

func TestLoad(t *testing.T) {
	file, err := os.Open(configFile)
	require.NoError(t, err, `打开配置文件`)

	defer func() {
		_ = file.Close()
	}()

	data := &Data{}

	require.NoError(t, json.NewDecoder(file).Decode(data), `解码`)

	from = data.Address
	privateKey = data.PrivateKey
}
func TestGrpcClient_FreezeBalance(t *testing.T) {
	TestLoad(t)

	g := client.NewGrpcClient(`grpc.trongrid.io:50051`)

	if err := g.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(err.Error())
	}

	txID, err := Freeze(g, from, privateKey, to, core.ResourceCode_ENERGY, 10000000)

	require.NoError(t, err)
	t.Log(txID)
}

func TestGrpcClient_UnFreezeBalance(t *testing.T) {
	TestLoad(t)

	g := client.NewGrpcClient(`grpc.trongrid.io:50051`)

	if err := g.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(err.Error())
	}

	txID, err := UnFreeze(g, from, privateKey, from, core.ResourceCode_ENERGY)

	require.NoError(t, err)
	t.Log(txID)
}

func TestResource(t *testing.T) {
	TestLoad(t)

	g := client.NewGrpcClient(`grpc.trongrid.io:50051`)

	if err := g.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(err.Error())
	}

	var (
		bandwidth, energy int64
		err               error
	)

	bandwidth, energy, err = Resource(g, to)

	require.NoError(t, err, `GetAccountResource`)
	t.Log(bandwidth, energy)
}

func TestFreezeResource(t *testing.T) {
	TestLoad(t)

	g := client.NewGrpcClient(`grpc.trongrid.io:50051`)

	if err := g.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(err.Error())
	}

	var (
		resources []*DelegatedResource
		err       error
	)

	resources, err = FreezeResource(g, from)

	require.NoError(t, err, `FreezeResource`)
	spew.Println(resources)
}

func TestFreezeResource1(t *testing.T) {
	TestLoad(t)

	g := client.NewGrpcClient(`grpc.trongrid.io:50051`)

	if err := g.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(err.Error())
	}

	var (
		resources []*DelegatedResource
		err       error
	)

	resources, err = FreezeResource(g, from)

	require.NoError(t, err, `FreezeResource`)
	spew.Println(resources)
}
