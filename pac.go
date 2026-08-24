package cfdi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PACConfig holds the settings required to reach a PAC provider via SOAP.
type PACConfig struct {
	// URL is the PAC webservice endpoint. It is used as the request URL, as
	// the SOAP operation namespace in the envelope, and as the base for the
	// auto-derived SOAPAction header.
	URL string
	// Username is the SOAP user, sent inside the SOAP Auth header.
	Username string
	// Password is the SOAP password, sent inside the SOAP Auth header.
	Password string
	// InvoicePassword is the certificate (FIEL/CSD) password, sent as the
	// "password" parameter of the generateInvoiceVersion4 body.
	InvoicePassword string
	// SOAPAction is the HTTP SOAPAction header value. When empty it
	// defaults to DefaultSOAPAction, the value accepted by cfdi360.
	SOAPAction string
	// Timeout is the request timeout in seconds.
	Timeout int
}

// DefaultSOAPAction is the SOAPAction header value proven against the
// cfdi360 generateInvoiceVersion4 operation.
const DefaultSOAPAction = "urn:document#documentServer#generateInvoiceVersion4"

// DefaultConfig returns a PACConfig with sensible defaults. The default URL
// points to a placeholder provider; set URL and credentials explicitly in
// production. SOAPAction is left empty so DefaultSOAPAction is used.
func DefaultConfig() *PACConfig {
	return &PACConfig{
		URL:     "https://api.cfdi_pack_provider.com/cfdi",
		Timeout: 600,
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
// response. docType is the SOAP documentType; when empty, "1" is used.
// fileType is the SOAP fileType ("0"=TXT, "1"=XML); when empty, "1" is used.
// commercio is the SOAP commercio flag; when empty, "0" is used.
// If cfg is nil, DefaultConfig is used.
func StampCFDI(cfg *PACConfig, xmlContent, docType, fileType, commercio string) (*PACResponse, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if docType == "" {
		docType = "1"
	}
	if fileType == "" {
		fileType = "1"
	}
	if commercio == "" {
		commercio = "0"
	}

	payload := BuildSOAPPayload(cfg.URL, xmlContent, docType, fileType, commercio, cfg.Username, cfg.Password, cfg.InvoicePassword)

	req, err := http.NewRequest("POST", cfg.URL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	soapAction := cfg.SOAPAction
	if soapAction == "" {
		soapAction = DefaultSOAPAction
	}
	req.Header.Set("SOAPAction", soapAction)

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