package cfdi

import (
	"fmt"
	"strings"
)

// BuildSOAPPayload escapes the XML document and wraps it in a SOAP 1.1
// envelope replicating the legacy PHP client's generateInvoiceVersion4
// request format. The operation namespace is the PAC webservice URL, and
// authentication is carried by a SOAP Auth header whose children use the
// literal "namespace" namespace (as the PHP SoapHeader('namespace', 'Auth',
// ...) call sends it). If the PAC rejects the request, verify against
// $client->__getLastRequest() / __getLastRequestHeaders() from PHP.
func BuildSOAPPayload(url, document, docType, fileType, commercio, user, password, invoicePassword string) string {
	escaped := strings.ReplaceAll(document, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\"", "&quot;")

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
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="%s">
  <soap:Header>
    <ns1:Auth xmlns:ns2="namespace">
      <ns2:userName>%s</ns2:userName>
      <ns2:password>%s</ns2:password>
    </ns1:Auth>
  </soap:Header>
  <soap:Body>
    <ns1:generateInvoiceVersion4>
      <document>%s</document>
      <documentType>%s</documentType>
      <fileType>%s</fileType>
      <cer></cer>
      <key></key>
      <password>%s</password>
      <commercio>%s</commercio>
    </ns1:generateInvoiceVersion4>
  </soap:Body>
</soap:Envelope>`, url, user, password, escaped, docType, fileType, invoicePassword, commercio)
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