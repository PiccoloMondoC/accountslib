// sky-accounts/pkg/clientlib/accountslib/sanctioned_countries.go
package accountslib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// SanctionedCountry represents the structure of a sanctioned country.
type SanctionedCountry struct {
	ID          uuid.UUID `json:"id"`
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	AddedAt     time.Time `json:"added_at"`
}

func (c *Client) IsCountrySanctioned(countryCode string) (bool, error) {
	// Validate the country code
	if len(countryCode) < 2 || len(countryCode) > 3 {
		return false, errors.New("invalid country code")
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/sanctions/countries/%s", c.BaseURL, url.PathEscape(countryCode)), nil)
	if err != nil {
		return false, fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	// Decode the response body
	var result struct {
		IsSanctioned bool `json:"isSanctioned"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("unable to decode response body: %w", err)
	}

	return result.IsSanctioned, nil
}

func (c *Client) AddSanctionedCountry(countryCode string, countryName string) error {
	// Create the payload
	payload := SanctionedCountry{
		CountryCode: countryCode,
		CountryName: countryName,
		AddedAt:     time.Now(),
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/sanctioned-countries", c.BaseURL), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}

func (c *Client) RemoveSanctionedCountry(countryCode string) error {
	// Validation
	if len(countryCode) < 2 || len(countryCode) > 3 {
		return errors.New("invalid country code")
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/sanctioned-countries/%s", c.BaseURL, countryCode), nil)
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-API-Key", c.ApiKey)

	// Send the HTTP request
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: got %v, body: %s", res.StatusCode, body)
	}

	return nil
}
