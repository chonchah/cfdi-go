package cfdi

import (
	"fmt"
	"strings"
)

// BuildSOAPPayload escapes the XML document and wraps it in a SOAP 1.1
// envelope byte-compatible with the legacy PHP SoapClient request that the
// cfdi360 sandbox accepted (verified end-to-end):
//   - "SOAP-ENV" envelope prefix;
//   - Auth header children (<userName>/<password>) carry no extra namespace;
//   - <cer/> and <key/> are emitted self-closing and empty;
//   - only &, < and > are escaped inside <document>; double quotes are left
//     untouched (legal inside element text content).
func BuildSOAPPayload(url, document, docType, fileType, commercio, user, password, invoicePassword string) string {
	escaped := strings.ReplaceAll(document, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")

	if docType == "" {
		docType = "1"
	}
	if fileType == "" {
		fileType = "1"
	}
	if commercio == "" {
		commercio = "0"
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="%s">
  <SOAP-ENV:Header>
    <ns1:Auth>
      <userName>%s</userName>
      <password>%s</password>
    </ns1:Auth>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <ns1:generateInvoiceVersion4>
      <document>%s</document>
      <documentType>%s</documentType>
      <fileType>%s</fileType>
      <cer/>
      <key/>
      <password>%s</password>
      <commercio>%s</commercio>
    </ns1:generateInvoiceVersion4>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, url, user, password, escaped, docType, fileType, invoicePassword, commercio)
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