package paystack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type DedicatedAccountRequest struct {
	Customer      interface{} `json:"customer"`
	PreferredBank string      `json:"preferred_bank,omitempty"`
}

type DedicatedAccountResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Bank struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
			Slug string `json:"slug"`
		} `json:"bank"`

		AccountName   string      `json:"account_name"`
		AccountNumber string      `json:"account_number"`
		Assigned      bool        `json:"assigned"`
		Currency      string      `json:"currency"`
		Metadata      interface{} `json:"metadata"`
		Active        bool        `json:"active"`
		ID            int         `json:"id"`
		CreatedAt     string      `json:"created_at"`
		UpdatedAt     string      `json:"updated_at"`

		Assignment struct {
			Integration  int    `json:"integration"`
			AssigneeID   int    `json:"assignee_id"`
			AssigneeType string `json:"assignee_type"`
			Expired      bool   `json:"expired"`
			AccountType  string `json:"account_type"`
			AssignedAt   string `json:"assigned_at"`
		} `json:"assignment"`

		Customer struct {
			ID           int    `json:"id"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
			Phone        string `json:"phone"`
			RiskAction   string `json:"risk_action"`
		} `json:"customer"`
	} `json:"data"`
}

type CustomerRequest struct {
	Email     string `json:"email"`
	Firstname string `json:"first_name"`
	Phone     string `json:"phone"`
	Lastname  string `json:"last_name"`
}

type CreateCustomerResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Email           string      `json:"email"`
		Integration     int         `json:"integration"`
		Domain          string      `json:"domain"`
		CustomerCode    string      `json:"customer_code"`
		ID              int         `json:"id"`
		Identified      bool        `json:"identified"`
		Identifications interface{} `json:"identifications"`
		CreatedAt       string      `json:"createdAt"`
		UpdatedAt       string      `json:"updatedAt"`
	} `json:"data"`
}

func CreateCustomer(customerReq CustomerRequest) (*CreateCustomerResponse, error) {
	body, err := json.Marshal(customerReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.paystack.co/customer",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+os.Getenv("PAYSTACK_SECRET_KEY"),
	)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateCustomerResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("paystack error: %s", result.Message)
	}

	return &result, nil
}

func CreateDedicatedAccount(customerCode string) (*DedicatedAccountResponse, error) {
	payload := DedicatedAccountRequest{
		Customer:      customerCode,
		PreferredBank: "test-bank",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.paystack.co/dedicated_account",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+os.Getenv("PAYSTACK_SECRET_KEY"),
	)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DedicatedAccountResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("paystack error: %s", result.Message)
	}

	return &result, nil
}
