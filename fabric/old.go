package fabric

import (
	"encoding/json"
	"errors"
	"fabricclient/logger"
	"fabricclient/util"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	Authorization = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1Mzk2NDk5NTksInVzZXJuYW1lIjoiSmltIiwib3JnTmFtZSI6Ik9yZzEiLCJpYXQiOjE1Mzk2MTM5NTl9.hL_5l4lV2CGJQY5DPZrPlTYds4P8eVfZSL6QUXcD6oo"
)

func (f *FabricClient) initValue(addr, num string) error {
	type FabricReq struct {
		Peers []string `json:"peers"`
		Fcn   string   `json:"fcn"`
		Args  []string `json:"args"`
	}

	peers := []string{"peer0.org1.example.com", "peer0.org2.example.com"}
	args := []string{addr, num}

	fabricReq := &FabricReq{
		Peers: peers,
		Fcn:   "initValue",
		Args:  args,
	}

	data, err := json.Marshal(fabricReq)
	if err != nil {
		logger.Error(err)
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:4000/channels/mychannel/chaincodes/mycc", strings.NewReader(string(data)))
	if err != nil {
		logger.Error(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Authorization)

	resp, err := f.cli.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug(string(body))

	return nil
}

func (f *FabricClient) move(from, to, num string) error {
	type FabricReq struct {
		Peers []string `json:"peers"`
		Fcn   string   `json:"fcn"`
		Args  []string `json:"args"`
	}

	peers := []string{"peer0.org1.example.com", "peer0.org2.example.com"}
	args := []string{from, to, num}

	fabricReq := &FabricReq{
		Peers: peers,
		Fcn:   "move",
		Args:  args,
	}

	data, err := json.Marshal(fabricReq)
	if err != nil {
		logger.Error(err)
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:4000/channels/mychannel/chaincodes/mycc", strings.NewReader(string(data)))
	if err != nil {
		logger.Error(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Authorization)

	resp, err := f.cli.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Debug(string(body))

	return nil
}

func (f *FabricClient) query(addr string) (string, error) {
	req, err := http.NewRequest("GET", "http://localhost:4000/channels/mychannel/chaincodes/mycc?peer=peer0.org1.example.com&fcn=query&args=['"+addr+"']", nil)
	if err != nil {
		logger.Error(err)
		return "0", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+Authorization)

	resp, err := f.cli.Do(req)
	if err != nil {
		logger.Error(err)
		return "0", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return "0", err
	}

	if len(body) == 0 {
		logger.Error("Query fail, addr :", addr)
		return "0", errors.New("Query fail, addr : " + addr)
	}

	logger.Debug("addr :", string(body))

	return string(body), nil
}

func (f *FabricClient) GetWallets(walletCount int) ([]*Wallet, []*Wallet, error) {
	richWallets := []*Wallet{}
	airWallets := []*Wallet{}

	for i := 0; i < walletCount; i++ {
		w := &Wallet{
			Address: util.GetUUID(),
			PrivKey: "",
		}
		richWallets = append(richWallets, w)

		w = &Wallet{
			Address: util.GetUUID(),
			PrivKey: "",
		}
		airWallets = append(airWallets, w)
	}

	return richWallets, airWallets, nil
}

func (f *FabricClient) fabricTest(richWallets, airWallets []*Wallet) error {
	/*  addrs := []string{"vsadv", "gnbdmkbjsk"}

	    err := f.initValue(addrs[0], "1000")
	    if err != nil {
	        logger.Error(err)
	        return err
	    }

	    b, err := f.query(addrs[0])
	    if err != nil {
	        logger.Error(err)
	        return err
	    }
	    logger.Info(addrs[0], b)

	    err = f.move(addrs[0], addrs[1], "56")
	    if err != nil {
	        logger.Error(err)
	        return err
	    }

	    b, err = f.query(addrs[0])
	    if err != nil {
	        logger.Error(err)
	        return err
	    }
	    logger.Info(addrs[0], b)

	    b, err = f.query(addrs[1])
	    if err != nil {
	        logger.Error(err)
	        return err
	    }
	    logger.Info(addrs[1], b)

	    return nil
	*/
	/*  for _, w := range richWallets {
	        b, err := f.query(w.Address)
	        if err != nil {
	            logger.Error(err)
	            continue
	        }

	        logger.Info("richWallets", w.Address, b)
	    }
	    return nil
	*/
	var wg sync.WaitGroup

	for i := 0; i < len(richWallets); i++ {
		wg.Add(1)

		go func(addr string) {
			defer wg.Done()

			time.Sleep(time.Second * 2)

			logger.Info(addr, "init value start")

			err := f.initValue(addr, "360")
			if err != nil {
				logger.Error(err)
				return
			}

			logger.Info(addr, "init value end")
		}(richWallets[i].Address)
	}

	wg.Wait()

	for _, w := range richWallets {
		b, err := f.query(w.Address)
		if err != nil {
			logger.Error(err)
			continue
		}

		logger.Info("richWallets", w.Address, b)
	}

	for i := 0; i < len(richWallets); i++ {
		wg.Add(1)

		go func(addr, recvAddr string) {
			defer wg.Done()

			time.Sleep(time.Second * 2)

			logger.Info(addr, "transfer start", recvAddr)

			err := f.move(addr, recvAddr, "1")
			if err != nil {
				logger.Error(err)
				return
			}

			logger.Info(addr, "transfer end", recvAddr)
		}(richWallets[i].Address, airWallets[i].Address)
	}

	wg.Wait()

	for _, w := range airWallets {
		b, err := f.query(w.Address)
		if err != nil {
			logger.Error(err)
			continue
		}

		logger.Info("airWallets", w.Address, b)
	}

	for _, w := range richWallets {
		b, err := f.query(w.Address)
		if err != nil {
			logger.Error(err)
			continue
		}

		logger.Info("richWallets", w.Address, b)
	}

	return nil
}
