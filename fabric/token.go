package fabric

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fabricclient/logger"
	"fabricclient/util"
	"io/ioutil"
	"net/http"
	"strings"
)

func (f *FabricClient) IssueToken(addr, privKey, tokenName, totalNumber string) (string, error) {
	type Token struct {
		Address     string `json:"address"`
		TokenName   string `json:"tokenName"`
		TotalNumber string `json:"totalNumber"`
	}

	origin := Token{
		Address:     addr,
		TokenName:   tokenName,
		TotalNumber: totalNumber,
	}

	originJson, err := json.Marshal(&origin)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	originJsonHexStr := hex.EncodeToString(originJson)

	signatureHexStr, err := util.Sign(privKey, []byte(originJsonHexStr))
	if err != nil {
		logger.Error(err)
		return "", err
	}

	type SendData struct {
		PubKey    string `json:"pubKey"`
		Origin    string `json:"origin"`
		Signature string `json:"signature"`
	}

	pubKey, err := util.GetPubKeyByPrivKey(privKey)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	sendData := SendData{
		PubKey:    pubKey,
		Origin:    originJsonHexStr,
		Signature: signatureHexStr,
	}

	data, err := json.Marshal(&sendData)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:4000/ocean/v1/issueToken", strings.NewReader(string(data)))
	if err != nil {
		logger.Error(err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.cli.Do(req)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	logger.Info(string(body))

	type Response struct {
		Status  bool   `json:"status"`
		Msg     string `json:"message"`
		TokenID string `json:"tokenID"`
	}

	res := Response{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	if !res.Status {
		logger.Error(res.Msg)
		return "", errors.New(res.Msg)
	}

	logger.Info("Successfully IssueToken, tokenID =", res.TokenID)

	return res.TokenID, nil
}

func (f *FabricClient) QueryToken(tokenID string) error {
	resp, err := f.cli.Get("http://localhost:4000/ocean/v1/queryToken/" + tokenID)
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

	logger.Info(string(body))

	return nil
}

func (f *FabricClient) Tranfer() error {
	return nil
}

func (f *FabricClient) QueryBalance(address string) error {
	resp, err := f.cli.Get("http://localhost:4000/ocean/v1/queryBalance/" + address)
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

	logger.Info(string(body))

	return nil
}
