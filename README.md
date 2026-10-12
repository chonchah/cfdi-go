# CFDI Go

[![Go Reference](https://pkg.go.dev/badge/github.com/chonchah/cfdi-go.svg)](https://pkg.go.dev/github.com/chonchah/cfdi-go)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](go.mod)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/chonchah/cfdi-go/total)


Dependencia-free (stdlib only) Go library for **Mexican electronic invoicing — CFDI 4.0 (SAT México)**. Build, validate and stamp CFDI documents.

## Features

- Build a CFDI 4.0 XML document from structured data (Ingreso and Pago with Pagos 2.0 complement)
- Validate a CFDI request against required tax fields
- Build and parse SOAP payloads exchanged with a PAC
- Stamp a CFDI via a PAC over plain HTTP

## Installation

```bash
go get github.com/chonchah/cfdi-go
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/chonchah/cfdi-go"
)

func main() {
	req := &cfdi.CFDIRequest{
		Version:           "4.0",
		Serie:             "A",
		Folio:             "100",
		Fecha:             "2026-07-01T10:00:00",
		SubTotal:          "1000.00",
		Moneda:            "MXN",
		Total:             "1160.00",
		TipoDeComprobante: "I",
		Exportacion:       "01",
		MetodoPago:        "PPD",
		FormaPago:         "99",
		LugarExpedicion:   "06000",
		Emisor: cfdi.Emisor{
			Rfc: "XAXX010101000", Nombre: "EMISORA EJEMPLO SA DE CV", RegimenFiscal: "601",
		},
		Receptor: cfdi.Receptor{
			Rfc: "XEXX010101000", Nombre: "CLIENTE EJEMPLO SA", DomicilioFiscalReceptor: "06500",
			RegimenFiscalReceptor: "601", UsoCFDI: "G03",
		},
		Conceptos: []cfdi.Concepto{
			{
				ClaveProdServ: "84111506", Cantidad: "1.00", ClaveUnidad: "E48",
				Descripcion: "Servicios de ejemplo", ValorUnitario: "1000.00", Importe: "1000.00",
				ObjetoImp: "02",
				Impuestos: &cfdi.ConceptoImpuestos{
					Traslados: &cfdi.TrasladosWrapper{Traslado: []cfdi.Traslado{
						{Base: "1000.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "160.00"},
					}},
				},
			},
		},
	}

	// Validate
	if ok, errs := req.Validate(); !ok {
		fmt.Println("validation errors:", errs)
		return
	}

	// Generate XML
	fmt.Println(req.ParseCFDItoXML())

	// Stamp via PAC (optional)
	resp, err := cfdi.StampCFDI(&cfdi.PACConfig{
		URL:             "https://your-pac.example/webservice",
		Username:        "user",
		Password:        "pass",
		InvoicePassword: "cert-password",
	}, req.ParseCFDItoXML(), req.DocumentType)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Success, *resp.PACResponse)
}
```

Parsing JSON input is also supported:

```go
var req cfdi.CFDIRequest
req.FromJSON([]byte(`{...}`))
```

## API

| Symbol | Description |
|---|---|
| `CFDIRequest` | Structured input for building a CFDI |
| `CFDIRequest.ParseCFDItoXML()` | Build the CFDI 4.0 XML string |
| `CFDIRequest.Validate()` | Validate required fields |
| `CFDIRequest.FromJSON()` | Decode from JSON |
| `BuildSOAPPayload(url, document, docType, fileType, commercio, user, password, invoicePassword)` | Wrap a document in the PAC SOAP envelope |
| `ExtractPACResponseText(soapResponse)` | Extract the PAC result from a SOAP response |
| `PACConfig` / `PACResponse` | PAC client configuration and result |
| `StampCFDI(cfg, xmlContent, docType, fileType, commercio)` | Send a document to a PAC for stamping (`docType` defaults to `"1"`, `fileType` to `"1"`, `commercio` to `"0"` when empty) |

## SOAP / PAC

`StampCFDI` talks to the PAC over SOAP 1.1, replicating the request format of a
legacy PHP client. Authentication is carried inside the SOAP `Auth` header
(not HTTP Basic Auth), and the operation is `generateInvoiceVersion4`.

Request envelope:

```xml
<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="{URL}">
  <soap:Header>
    <ns1:Auth xmlns:ns2="namespace">
      <ns2:userName>{user}</ns2:userName>
      <ns2:password>{password}</ns2:password>
    </ns1:Auth>
  </soap:Header>
  <soap:Body>
    <ns1:generateInvoiceVersion4>
      <document>{escaped CFDI XML}</document>
      <documentType>{docType}</documentType>
      <fileType>{fileType}</fileType>
      <cer></cer>
      <key></key>
      <password>{certificate password}</password>
      <commercio>{commercio}</commercio>
    </ns1:generateInvoiceVersion4>
  </soap:Body>
</soap:Envelope>
```

The operation namespace is the webservice `URL`. The CFDI document is escaped
(`&`, `<`, `>`, `"` → entities) inside `<document>`.

The `<documentType>` parameter is configurable per request via the optional
`documentType` field on `CFDIRequest` (JSON key `documentType`). It is passed
to `StampCFDI` as the `docType` argument; when empty (or omitted from the
JSON), the default `"1"` (Ingreso) is used. Likewise, `fileType`
(`"0"`=TXT, `"1"`=XML, default `"1"`) and `commercio` (default `"0"`) are
configurable via the `fileType` and `commercio` fields on `CFDIRequest`.

`PACConfig` fields:

| Field | Meaning |
|---|---|
| `URL` | PAC webservice endpoint; used as request URL and SOAP operation namespace |
| `Username` | SOAP user, sent in the SOAP `Auth` header |
| `Password` | SOAP password, sent in the SOAP `Auth` header |
| `InvoicePassword` | Certificate (FIEL/CSD) password, sent as the body `password` parameter |
| `SOAPAction` | HTTP `SOAPAction` header; if empty it is derived as `URL + "#generateInvoiceVersion4"` |
| `Timeout` | Request timeout in seconds |

The `Auth` header children (`userName`/`password`) and the derived `SOAPAction`
assume the format the PHP client sends. If your PAC rejects the request, capture
the real PHP request with `$client->__getLastRequest()` /
`__getLastRequestHeaders()` and adjust the assumptions accordingly.

## Example service

An HTTP example consuming the library is included under [`cmd/make-cfdi`](cmd/make-cfdi). Run it locally:

```bash
go run ./cmd/make-cfdi
# POST http://localhost:8080/make-cfdi
```

Build the container image:

```bash
docker build -t cfdi-go .
```

## Tests

```bash
go test ./... -v
go vet ./...
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
