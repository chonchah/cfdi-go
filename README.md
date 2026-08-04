# CFDI Go

[![Go Reference](https://pkg.go.dev/badge/github.com/chonchah/cfdi-go.svg)](https://pkg.go.dev/github.com/chonchah/cfdi-go)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](go.mod)

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
		URL:      "https://your-pac.example/cfdi",
		Username: "user",
		Password: "pass",
	}, req.ParseCFDItoXML())
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
| `BuildSOAPPayload(document, docType)` | Wrap a document in a SOAP envelope |
| `ExtractPACResponseText(soapResponse)` | Extract the PAC result from a SOAP response |
| `PACConfig` / `PACResponse` | PAC client configuration and result |
| `StampCFDI(cfg, xmlContent)` | Send a document to a PAC for stamping |

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