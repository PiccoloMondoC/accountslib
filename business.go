package accountslib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/uuid"
)

// Business represents the structure of a business.
type Business struct {
	ID            uuid.UUID `json:"id"`
	UserAccountID uuid.UUID `json:"user_account_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateBusinessAccountInput struct {
	UserID       uuid.UUID `json:"user_id"`
	BusinessName string    `json:"business_name"`
}

type UpdateBusinessAccountInput struct {
	UserID          uuid.UUID `json:"user_id"`
	BusinessName    string    `json:"business_name"`
	NewBusinessName string    `json:"new_business_name"`
	BusinessID      uuid.UUID `json:"business_id"`
}

type AddMemberToBusinessAccountInput struct {
	UserID     uuid.UUID `json:"user_id"`
	BusinessID uuid.UUID `json:"business_id"`
	RoleID     uuid.UUID `json:"role_id"`
}

type UpdateMemberRoleInBusinessAccountInput struct {
	BusinessID   uuid.UUID `json:"business_id"`
	MemberUserID uuid.UUID `json:"member_user_id"`
	NewRoleID    uuid.UUID `json:"new_role_id"`
}

// CreateBusinessAccount creates a new business account for a given user.
func (c *Client) CreateBusinessAccount(input CreateBusinessAccountInput) (*Business, error) {
	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	requestURL.Path = path.Join(requestURL.Path, "/api/v1/business/")

	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token)) // This assumes you're using Bearer token authentication
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create business account: %s", string(bodyBytes))
	}

	var newBusinessAccount Business
	err = json.NewDecoder(resp.Body).Decode(&newBusinessAccount)
	if err != nil {
		return nil, err
	}

	return &newBusinessAccount, nil
}

func (c *Client) GetBusinessAccountByID(businessID uuid.UUID) (*Business, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %w", err)
	}
	u.Path = path.Join(u.Path, "business", businessID.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("X-API-Key", c.ApiKey)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status back from server: %d (%s)", resp.StatusCode, string(body))
	}
	var business Business
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&business); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &business, nil
}

func (c *Client) GetBusinessAccountsByUserID(userID uuid.UUID) ([]Business, error) {
	// Parse BaseURL and create a new URL
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Create an endpoint path
	endpointPath := path.Join("users", userID.String(), "business_accounts")

	// Resolve to get the full endpoint URL
	endpointURL := baseURL.ResolveReference(&url.URL{Path: endpointPath})

	// Create a new HTTP request
	req, err := http.NewRequest("GET", endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check the HTTP response status
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from server: %s", res.Status)
	}

	// Decode the HTTP response body
	var businessAccounts []Business
	if err := json.NewDecoder(res.Body).Decode(&businessAccounts); err != nil {
		return nil, err
	}

	return businessAccounts, nil
}

func (u *UpdateBusinessAccountInput) Validate() error {
	// Perform validation on u fields
	if u.UserID == uuid.Nil {
		return errors.New("userID cannot be empty")
	}
	if u.NewBusinessName == "" {
		return errors.New("newBusinessName cannot be empty")
	}
	if u.BusinessID == uuid.Nil {
		return errors.New("businessID cannot be empty")
	}

	// Return nil if all checks pass
	return nil
}

func (c *Client) UpdateBusinessAccount(input UpdateBusinessAccountInput) error {
	// Validate the payload
	err := input.Validate()
	if err != nil {
		return err
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// Create the HTTP request
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", c.BaseURL, input.BusinessID), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("X-API-Key", c.ApiKey)

	// Execute the request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	// Handle the response
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (c *Client) DeleteBusinessAccount(businessID uuid.UUID) error {
	// Construct the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %v", err)
	}
	u.Path = path.Join(u.Path, "business", businessID.String())

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return fmt.Errorf("could not create HTTP request: %v", err)
	}

	// Add authorization header if you have one
	if c.Token != "" {
		req.Header.Add("Authorization", "Bearer "+c.Token)
	}

	// Add any other necessary headers like API Key
	if c.ApiKey != "" {
		req.Header.Add("X-Api-Key", c.ApiKey)
	}

	// Send the HTTP request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	return nil
}

func (c *Client) ListBusinessAccounts() ([]Business, error) {
	// Prepare a new request
	reqURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL: %w", err)
	}

	reqURL.Path = path.Join(reqURL.Path, "business-accounts")
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add authorization headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("x-api-key", c.ApiKey)

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response
	var businessAccounts []Business
	if err = json.NewDecoder(resp.Body).Decode(&businessAccounts); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return businessAccounts, nil
}

// Define the validation for AddMemberToBusinessAccountInput
func (input *AddMemberToBusinessAccountInput) Validate() error {
	if input.UserID == uuid.Nil {
		return errors.New("UserID cannot be empty")
	}
	if input.BusinessID == uuid.Nil {
		return errors.New("BusinessID cannot be empty")
	}
	if input.RoleID == uuid.Nil {
		return errors.New("RoleID cannot be empty")
	}

	// Return nil if all checks pass
	return nil
}

// Update the AddMemberToBusinessAccount function
func (c *Client) AddMemberToBusinessAccount(input AddMemberToBusinessAccountInput) error {
	// Validate the input
	err := input.Validate()
	if err != nil {
		return err
	}

	requestURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	requestURL.Path = path.Join(requestURL.Path, "businesses", input.BusinessID.String(), "members")

	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, requestURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	// Optionally, we can read and return the response
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoveMemberFromBusinessAccount(businessID, memberID uuid.UUID) error {
	// Create the URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, "api", "businesses", businessID.String(), "members", memberID.String())

	// Create the request
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	// Send the request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response from server: %s (%d)", body, resp.StatusCode)
	}

	return nil
}

func (c *Client) GetMembersOfBusinessAccount(businessId uuid.UUID) ([]AccountMembership, error) {
	// Create a new URL from the base url of the client
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	// Specify the path for the businesses API
	u.Path = path.Join(u.Path, "api/v1/businesses", businessId.String(), "members")

	// Create a new request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	// Send the request via the HTTP client
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response from server: %s", string(bodyBytes))
	}

	// Decode the response
	var memberships []AccountMembership
	err = json.NewDecoder(resp.Body).Decode(&memberships)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Return the memberships
	return memberships, nil
}

// Validate validates input data for UpdateMemberRoleInBusinessAccountInput
func (input *UpdateMemberRoleInBusinessAccountInput) Validate() error {
	if input.BusinessID == uuid.Nil {
		return errors.New("BusinessID cannot be empty")
	}
	if input.MemberUserID == uuid.Nil {
		return errors.New("MemberUserID cannot be empty")
	}
	if input.NewRoleID == uuid.Nil {
		return errors.New("NewRoleID cannot be empty")
	}

	// Return nil if all checks pass
	return nil
}

// UpdateMemberRoleInBusinessAccount sends a request to the REST API endpoint to update a member's role in a business account.
func (c *Client) UpdateMemberRoleInBusinessAccount(input UpdateMemberRoleInBusinessAccountInput) error {
	// Validate the input
	err := input.Validate()
	if err != nil {
		return err
	}

	updateURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base url: %w", err)
	}

	updateURL.Path = path.Join(updateURL.Path, "api/v1/businesses", input.BusinessID.String(), "members", input.MemberUserID.String())

	reqBody := &UpdateAccountMembershipEvent{
		AccountType: "business",
		AccountID:   input.BusinessID,
		UserID:      input.MemberUserID,
		Role:        input.NewRoleID.String(),
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, updateURL.String(), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create a request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send a request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
