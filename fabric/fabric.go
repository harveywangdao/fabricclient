package fabric

import (
	"encoding/json"
	"fabricclient/logger"
	"fabricclient/util"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type FabricClient struct {
	urlHead string
	cli     *http.Client
	tp      *TestParam
}

type Wallet struct {
	Address string `json:"address"`
	//PubKey  string `json:"pubKey"`
	PrivKey string `json:"privKey"`
}

type TestParam struct {
	Token1Wallet Wallet `json:"token1Wallet"`
	TokenID1     string `json:"tokenID1"`
	Token2Wallet Wallet `json:"token2Wallet"`
	TokenID2     string `json:"tokenID2"`
}

func (f *FabricClient) testTransfer() error {
	tp := f.tp

	for i := 0; i < 10; i++ {
		f.Transfer(tp.TokenID1, tp.Token1Wallet.Address, tp.Token1Wallet.PrivKey, tp.Token2Wallet.Address, "50")
		f.QueryBalance(tp.Token1Wallet.Address)
		f.QueryBalance(tp.Token2Wallet.Address)
	}

	return nil
}

func (f *FabricClient) genWallets(num int) []*Wallet {
	ws := []*Wallet{}

	for i := 0; i < num; i++ {
		w := &Wallet{}
		w.PrivKey, _, w.Address = util.GetNewAddress()
		ws = append(ws, w)
	}

	return ws
}

func (f *FabricClient) highConcurrent() error {
	walletMum := 50
	group1 := f.genWallets(walletMum)
	group2 := f.genWallets(walletMum)

	for i := 0; i < walletMum; i++ {
		_, err := f.Transfer(f.tp.TokenID1, f.tp.Token1Wallet.Address, f.tp.Token1Wallet.PrivKey, group1[i].Address, "1")
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	for i := 0; i < walletMum; i++ {
		f.QueryBalance(group1[i].Address)
	}

	var wg sync.WaitGroup

	for i := 0; i < walletMum; i++ {
		wg.Add(1)
		go func(fromAddr, fromPrivKey, toAddr string) {
			defer wg.Done()

			logger.Info("transfer start", fromAddr, toAddr)
			time.Sleep(time.Second * 2)

			_, err := f.Transfer(f.tp.TokenID1, fromAddr, fromPrivKey, toAddr, "1")
			if err != nil {
				logger.Error(err)
				return
			}

			logger.Info("transfer end", fromAddr, toAddr)
		}(group1[i].Address, group1[i].PrivKey, group2[i].Address)
	}

	wg.Wait()

	for i := 0; i < walletMum; i++ {
		f.QueryBalance(group1[i].Address)
	}

	for i := 0; i < walletMum; i++ {
		f.QueryBalance(group2[i].Address)
	}

	return nil
}

func (f *FabricClient) testApiInit() error {
	var err error

	if f.tp != nil {
		err = f.QueryToken(f.tp.TokenID1)
		if err == nil {
			return nil
		}
	}

	tp := &TestParam{}

	tp.Token1Wallet.PrivKey, _, tp.Token1Wallet.Address = util.GetNewAddress()
	tp.TokenID1, err = f.IssueToken(tp.Token1Wallet.Address, tp.Token1Wallet.PrivKey, "OCE", "10000")
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("tp.TokenID1 =", tp.TokenID1)

	err = f.QueryToken(tp.TokenID1)
	if err != nil {
		logger.Error(err)
		return err
	}

	tp.TokenID2, _ = f.IssueToken(tp.Token1Wallet.Address, tp.Token1Wallet.PrivKey, "OCE2", "20000")
	err = f.QueryBalance(tp.Token1Wallet.Address)
	if err != nil {
		logger.Error(err)
		return err
	}

	tp.Token2Wallet.PrivKey, _, tp.Token2Wallet.Address = util.GetNewAddress()
	txID, err := f.Transfer(tp.TokenID1, tp.Token1Wallet.Address, tp.Token1Wallet.PrivKey, tp.Token2Wallet.Address, "100")
	if err != nil {
		logger.Error(err)
		return err
	}

	err = f.QueryTx(txID)
	if err != nil {
		logger.Error(err)
		return err
	}

	f.QueryBalance(tp.Token1Wallet.Address)
	f.QueryBalance(tp.Token2Wallet.Address)

	f.tp = tp
	tpData, err := json.Marshal(tp)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = ioutil.WriteFile("conf/TestParam.json", tpData, os.ModePerm)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (f *FabricClient) testApi() error {
	if util.IsFileExist("conf/TestParam.json") {
		data, err := ioutil.ReadFile("conf/TestParam.json")
		if err != nil {
			logger.Error(err)
			return err
		}

		tp := &TestParam{}
		err = json.Unmarshal(data, tp)
		if err != nil {
			logger.Error(err)
			return err
		}

		f.tp = tp
	}

	err := f.testApiInit()
	if err != nil {
		logger.Error(err)
		return err
	}

	/*	err = f.testTransfer()
		if err != nil {
			logger.Error(err)
			return err
		}*/

	return nil
}

func (f *FabricClient) testing(wg *sync.WaitGroup) {
	defer wg.Done()
	var err error

	err = f.testApi()
	if err != nil {
		logger.Error(err)
		return
	}

	err = f.highConcurrent()
	if err != nil {
		logger.Error(err)
		return
	}
}

func NewFabricClient(ipport string, wg *sync.WaitGroup) (*FabricClient, error) {
	f := new(FabricClient)

	f.cli = &http.Client{}
	f.urlHead = "http://" + ipport

	go f.testing(wg)

	return f, nil
}
