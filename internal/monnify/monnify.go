package monnify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ReservedAccount struct {
	AccountReference     string `json:"accountReference"`
	AccountName          string `json:"accountName"`
	CurrencyCode         string `json:"currencyCode"`
	ContractCode         string `json:"contractCode"`
	CustomerEmail        string `json:"customerEmail"`
	BVN                  string `json:"bvn"`
	CustomerName         string `json:"customerName"`
	GetAllAvailableBanks bool   `json:"getAllAvailableBanks"`
}

func CreateReservedAccount(client *http.Client, req ReservedAccount) (*ReservedAccount, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"https://sandbox.monnify.com/api/v2/bank-transfer/reserved-accounts",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	//httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, body)
	}

	var result ReservedAccount
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
