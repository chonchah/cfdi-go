package cfdi

import (
	"fmt"
	"strings"
)

// BuildSOAPPayload escapes the XML document and wraps it in a SOAP envelope
// with the GenerateInvoice operation understood by the PAC provider.
func BuildSOAPPayload(document string, docType string) string {
	escaped := strings.ReplaceAll(document, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\"", "&quot;")

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GenerateInvoice xmlns="http://cfdi_pack_provider.com/">
      <document>%s</document>
      <documentType>%s</documentType>
    </GenerateInvoice>
  </soap:Body>
</soap:Envelope>`, escaped, docType)
}

// ExtractPACResponseText extracts the meaningful text from a SOAP response.
// It finds content containing "SUCCESS" between XML tags and returns it.
// Expected format: "SUCCESS:url_xml,url_pdf".
func ExtractPACResponseText(soapResponse string) (string, error) {
	if !strings.Contains(soapResponse, "SUCCESS") {
		return "", fmt.Errorf("PAC returned error: %s", soapResponse)
	}
	start := strings.Index(soapResponse, "SUCCESS")
	if start == -1 {
		return "", fmt.Errorf("no SUCCESS marker in SOAP response: %s", soapResponse)
	}
	before := strings.LastIndex(soapResponse[:start], ">")
	if before == -1 {
		return strings.TrimSpace(soapResponse), nil
	}
	contentStart := before + 1
	end := strings.Index(soapResponse[contentStart:], "<")
	if end == -1 {
		return strings.TrimSpace(soapResponse[contentStart:]), nil
	}
	return strings.TrimSpace(soapResponse[contentStart : contentStart+end]), nil
}