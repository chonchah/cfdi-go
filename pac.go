package cfdi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PACConfig holds the settings required to reach a PAC provider via SOAP.
type PACConfig struct {
	URL        string
	Username   string
	Password   string
	SOAPAction string
	Timeout    int
}

// DefaultConfig returns a PACConfig with sensible defaults. The default URL
// points to a placeholder provider; set URL and credentials explicitly in
// production.
func DefaultConfig() *PACConfig {
	return &PACConfig{
		URL:        "https://api.cfdi_pack_provider.com/cfdi",
		SOAPAction: "GenerateInvoice",
		Timeout:    600,
	}
}

// PACResponse is the outcome of a stamping request.
type PACResponse struct {
	Success     bool    `json:"success"`
	XML         string  `json:"xml,omitempty"`
	PACResponse *string `json:"pac_response"`
	Error       string  `json:"error,omitempty"`
}

// StampCFDI sends the CFDI XML to the PAC for stamping and returns the raw
// response. If cfg is nil, DefaultConfig is used.
func StampCFDI(cfg *PACConfig, xmlContent string) (*PACResponse, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	payload := BuildSOAPPayload(xmlContent, "1")

	req, err := http.NewRequest("POST", cfg.URL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", cfg.SOAPAction)

	if cfg.Username != "" && cfg.Password != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SOAP response: %w", err)
	}

	soapResponse := string(body)

	pacText, err := ExtractPACResponseText(soapResponse)
	if err != nil {
		return &PACResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &PACResponse{
		Success:     true,
		PACResponse: &pacText,
	}, nil
}