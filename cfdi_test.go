package cfdi

import (
	"strings"
	"testing"
)

func TestGenerateIngresoXML(t *testing.T) {
	req := &CFDIRequest{
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
		Emisor: Emisor{
			Rfc:           "XAXX010101000",
			Nombre:        "EMISORA EJEMPLO SA DE CV",
			RegimenFiscal: "601",
		},
		Receptor: Receptor{
			Rfc:                     "XEXX010101000",
			Nombre:                  "CLIENTE EJEMPLO SA",
			DomicilioFiscalReceptor: "06500",
			RegimenFiscalReceptor:   "601",
			UsoCFDI:                 "G03",
		},
		Conceptos: []Concepto{
			{
				ClaveProdServ: "84111506",
				Cantidad:      "1.00",
				ClaveUnidad:   "E48",
				Descripcion:   "Servicios de ejemplo",
				ValorUnitario: "1000.00",
				Importe:       "1000.00",
				ObjetoImp:     "02",
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{
						{
							Base:       "1000.00",
							Impuesto:   "002",
							TipoFactor: "Tasa",
							TasaOCuota: "0.160000",
							Importe:    "160.00",
						},
					},
				},
			},
			},
		},
	}

	xmlOut := req.ParseCFDItoXML()

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"root element cfdi:Comprobante", func() bool { return strings.Contains(xmlOut, "cfdi:Comprobante") }},
		{"Version=\"4.0\"", func() bool { return strings.Contains(xmlOut, `Version="4.0"`) }},
		{"Serie=\"A\"", func() bool { return strings.Contains(xmlOut, `Serie="A"`) }},
		{"Folio=\"100\"", func() bool { return strings.Contains(xmlOut, `Folio="100"`) }},
		{"TipoDeComprobante=\"I\"", func() bool { return strings.Contains(xmlOut, `TipoDeComprobante="I"`) }},
		{"MetodoPago=\"PPD\"", func() bool { return strings.Contains(xmlOut, `MetodoPago="PPD"`) }},
		{"FormaPago=\"99\"", func() bool { return strings.Contains(xmlOut, `FormaPago="99"`) }},
		{"SubTotal=\"1000.00\"", func() bool { return strings.Contains(xmlOut, `SubTotal="1000.00"`) }},
		{"Total=\"1160.00\"", func() bool { return strings.Contains(xmlOut, `Total="1160.00"`) }},
		{"Exportacion=\"01\"", func() bool { return strings.Contains(xmlOut, `Exportacion="01"`) }},
		{"LugarExpedicion=\"06000\"", func() bool { return strings.Contains(xmlOut, `LugarExpedicion="06000"`) }},
		{"Emisor element", func() bool { return strings.Contains(xmlOut, "cfdi:Emisor") }},
		{"Emisor Rfc", func() bool { return strings.Contains(xmlOut, `Rfc="XAXX010101000"`) }},
		{"RegimenFiscal", func() bool { return strings.Contains(xmlOut, `RegimenFiscal="601"`) }},
		{"Receptor element", func() bool { return strings.Contains(xmlOut, "cfdi:Receptor") }},
		{"UsoCFDI=\"G03\"", func() bool { return strings.Contains(xmlOut, `UsoCFDI="G03"`) }},
		{"Conceptos element", func() bool { return strings.Contains(xmlOut, "cfdi:Conceptos") }},
		{"Concepto element", func() bool { return strings.Contains(xmlOut, "cfdi:Concepto") }},
		{"ClaveProdServ", func() bool { return strings.Contains(xmlOut, `ClaveProdServ="84111506"`) }},
		{"ObjetoImp=\"02\"", func() bool { return strings.Contains(xmlOut, `ObjetoImp="02"`) }},
		{"Traslados element", func() bool { return strings.Contains(xmlOut, "cfdi:Traslados") }},
		{"TasaOCuota", func() bool { return strings.Contains(xmlOut, `TasaOCuota="0.160000"`) }},
		{"comprobante-level Impuestos with total", func() bool { return strings.Contains(xmlOut, `TotalImpuestosTrasladados="160.00"`) }},
		{"no pago20 namespace for I-type", func() bool { return !strings.Contains(xmlOut, `xmlns:pago20`) }},
	}

	for _, check := range checks {
		if !check.fn() {
			t.Errorf("FAIL: %s", check.name)
		}
	}
}

func TestGeneratePagoXML(t *testing.T) {
	req := &CFDIRequest{
		Version:           "4.0",
		Serie:             "P",
		Folio:             "50",
		Fecha:             "2026-07-15T14:30:00",
		SubTotal:          "0",
		Moneda:            "MXN",
		Total:             "0",
		TipoDeComprobante: "P",
		Exportacion:       "01",
		LugarExpedicion:   "06000",
		Emisor: Emisor{
			Rfc:           "XAXX010101000",
			Nombre:        "EMISORA EJEMPLO SA DE CV",
			RegimenFiscal: "601",
		},
		Receptor: Receptor{
			Rfc:                     "XEXX010101000",
			Nombre:                  "CLIENTE EJEMPLO SA",
			DomicilioFiscalReceptor: "06500",
			RegimenFiscalReceptor:   "601",
			UsoCFDI:                 "CP01",
		},
		Conceptos: []Concepto{
			{
				ClaveProdServ: "84111506",
				Cantidad:      "1.00",
				ClaveUnidad:   "ACT",
				Descripcion:   "Pago",
				ValorUnitario: "0",
				Importe:       "0",
				ObjetoImp:     "01",
			},
		},
		Complemento: &Complemento{
			Pagos: &Pagos{
				Version: "2.0",
				Totales: Totales{
					TotalTrasladosBaseIVA16:    "1000.00",
					TotalTrasladosImpuestoIVA16: "160.00",
					MontoTotalPagos:            "1160.00",
				},
				Pago: []Pago{
					{
						FechaPago:    "2026-07-15T14:25:00",
						FormaDePagoP: "03",
						MonedaP:      "MXN",
						Monto:        "1160.00",
						DoctoRelacionado: []DoctoRelacionado{
							{
								IdDocumento:     "UUID-DE-LA-FACTURA-ORIGINAL-PPD",
								Serie:           "A",
								Folio:           "100",
								MonedaDR:        "MXN",
								EquivMonedaDR:   "1",
								NumParcialidad:  "1",
								ImpSaldoAnt:     "1160.00",
								ImpPagado:       "1160.00",
								ImpSaldoInsoluto: "0.00",
								ObjetoImpDR:     "02",
							ImpuestosDR: &DoctoImpuestosDR{
								TrasladosDR: &TrasladosDRWrapper{
									TrasladoDR: []TrasladoDR{
										{
											BaseDR:      "1000.00",
											ImpuestoDR:  "002",
											TipoFactorDR: "Tasa",
											TasaOCuotaDR: "0.160000",
											ImporteDR:   "160.00",
										},
									},
								},
							},
							},
						},
					},
				},
			},
		},
	}

	xmlOut := req.ParseCFDItoXML()

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"TipoDeComprobante=\"P\"", func() bool { return strings.Contains(xmlOut, `TipoDeComprobante="P"`) }},
		{"pago20 namespace", func() bool { return strings.Contains(xmlOut, `xmlns:pago20="http://www.sat.gob.mx/Pagos20"`) }},
		{"Pagos element", func() bool { return strings.Contains(xmlOut, "pago20:Pagos") }},
		{"Pagos Version=\"2.0\"", func() bool { return strings.Contains(xmlOut, `Version="2.0"`) }},
		{"Totales element", func() bool { return strings.Contains(xmlOut, "pago20:Totales") }},
		{"MontoTotalPagos", func() bool { return strings.Contains(xmlOut, `MontoTotalPagos="1160.00"`) }},
		{"Pago element", func() bool { return strings.Contains(xmlOut, "pago20:Pago") }},
		{"FechaPago", func() bool { return strings.Contains(xmlOut, `FechaPago="2026-07-15T14:25:00"`) }},
		{"DoctoRelacionado element", func() bool { return strings.Contains(xmlOut, "pago20:DoctoRelacionado") }},
		{"IdDocumento", func() bool { return strings.Contains(xmlOut, `IdDocumento="UUID-DE-LA-FACTURA-ORIGINAL-PPD"`) }},
		{"ImpSaldoAnt", func() bool { return strings.Contains(xmlOut, `ImpSaldoAnt="1160.00"`) }},
		{"ImpPagado", func() bool { return strings.Contains(xmlOut, `ImpPagado="1160.00"`) }},
		{"ImpSaldoInsoluto", func() bool { return strings.Contains(xmlOut, `ImpSaldoInsoluto="0.00"`) }},
		{"ImpuestosDR element", func() bool { return strings.Contains(xmlOut, "pago20:ImpuestosDR") }},
		{"TrasladosDR element", func() bool { return strings.Contains(xmlOut, "pago20:TrasladosDR") }},
		{"TrasladoDR element", func() bool { return strings.Contains(xmlOut, "pago20:TrasladoDR") }},
		{"TasaOCuotaDR", func() bool { return strings.Contains(xmlOut, `TasaOCuotaDR="0.160000"`) }},
		{"UsoCFDI=\"CP01\"", func() bool { return strings.Contains(xmlOut, `UsoCFDI="CP01"`) }},
		{"Complemento element", func() bool { return strings.Contains(xmlOut, "cfdi:Complemento") }},
		{"no MetodoPago for P-type", func() bool { return !strings.Contains(xmlOut, `MetodoPago=`) }},
		{"no FormaPago for P-type", func() bool { return !strings.Contains(xmlOut, `FormaPago=`) }},
		{"no TotalImpuestosTrasladados for P-type", func() bool { return !strings.Contains(xmlOut, `TotalImpuestosTrasladados`) }},
	}

	for _, check := range checks {
		if !check.fn() {
			t.Errorf("FAIL: %s", check.name)
		}
	}
}

func TestValidateIngreso(t *testing.T) {
	req := &CFDIRequest{
		TipoDeComprobante: "I",
		Version:           "4.0",
		Serie:             "A",
		Folio:             "100",
		Fecha:             "2026-07-01T10:00:00",
		SubTotal:          "1000.00",
		Moneda:            "MXN",
		Total:             "1160.00",
		Exportacion:       "01",
		MetodoPago:        "PPD",
		FormaPago:         "99",
		LugarExpedicion:   "06000",
		Emisor:            Emisor{Rfc: "XAXX010101000", Nombre: "E", RegimenFiscal: "601"},
		Receptor:          Receptor{Rfc: "XEXX010101000", Nombre: "C", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos:         []Concepto{{ClaveProdServ: "84111506", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "S", ValorUnitario: "1000", Importe: "1000", ObjetoImp: "02"}},
	}

	valid, errs := req.Validate()
	if !valid {
		t.Errorf("Expected valid I-type request, got errors: %v", errs)
	}
}

func TestValidatePago(t *testing.T) {
	req := &CFDIRequest{
		TipoDeComprobante: "P",
		Version:           "4.0",
		Serie:             "P",
		Folio:             "50",
		Fecha:             "2026-07-15T14:30:00",
		SubTotal:          "0",
		Total:             "0",
		Exportacion:       "01",
		LugarExpedicion:   "06000",
		Emisor:            Emisor{Rfc: "XAXX010101000", Nombre: "E", RegimenFiscal: "601"},
		Receptor:          Receptor{Rfc: "XEXX010101000", Nombre: "C", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "CP01"},
		Conceptos:         []Concepto{{ClaveProdServ: "84111506", Cantidad: "1", ClaveUnidad: "ACT", Descripcion: "Pago", ValorUnitario: "0", Importe: "0", ObjetoImp: "01"}},
		Complemento: &Complemento{
			Pagos: &Pagos{
				Version: "2.0",
				Totales: Totales{TotalTrasladosBaseIVA16: "1000", TotalTrasladosImpuestoIVA16: "160", MontoTotalPagos: "1160"},
				Pago:    []Pago{{FechaPago: "2026-07-15T14:25:00", FormaDePagoP: "03", MonedaP: "MXN", Monto: "1160.00"}},
			},
		},
	}

	valid, errs := req.Validate()
	if !valid {
		t.Errorf("Expected valid P-type request, got errors: %v", errs)
	}
}

func TestValidateMissingFields(t *testing.T) {
	req := &CFDIRequest{}
	valid, errs := req.Validate()
	if valid {
		t.Error("Expected validation to fail for empty request")
	}
	if len(errs) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestValidateInvalidTipo(t *testing.T) {
	req := &CFDIRequest{
		TipoDeComprobante: "X",
		Version:           "4.0",
		Serie:             "A",
		Folio:             "1",
		Fecha:             "2026-01-01T00:00:00",
		Exportacion:       "01",
		LugarExpedicion:   "06000",
		Emisor:            Emisor{Rfc: "X", Nombre: "E", RegimenFiscal: "601"},
		Receptor:          Receptor{Rfc: "Y", Nombre: "R", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos:         []Concepto{{ClaveProdServ: "1", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "S", ValorUnitario: "1", Importe: "1", ObjetoImp: "02"}},
	}

	valid, errs := req.Validate()
	if valid {
		t.Error("Expected validation to fail for invalid TipoDeComprobante")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "must be I") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about TipoDeComprobante, got: %v", errs)
	}
}

func TestBuildSOAPPayload(t *testing.T) {
	doc := `<cfdi:Comprobante><cfdi:Emisor/></cfdi:Comprobante>`
	payload := BuildSOAPPayload(doc, "1")

	if !strings.Contains(payload, "soap:Envelope") {
		t.Error("Missing SOAP Envelope")
	}
	if !strings.Contains(payload, "GenerateInvoice") {
		t.Error("Missing GenerateInvoice element")
	}
	if !strings.Contains(payload, "&lt;cfdi:Comprobante&gt;") {
		t.Error("Document should be XML-escaped in SOAP payload")
	}
}

func TestExtractPACResponseText(t *testing.T) {
	soapResp := `<?xml version="1.0"?><soap:Envelope><soap:Body><GenerateInvoiceResponse><GenerateInvoiceResult>SUCCESS:https://example.com/xml,https://example.com/pdf</GenerateInvoiceResult></GenerateInvoiceResponse></soap:Body></soap:Envelope>`
	text, err := ExtractPACResponseText(soapResp)
	if err != nil {
		t.Fatalf("ExtractPACResponseText failed: %s", err)
	}
	if text != "SUCCESS:https://example.com/xml,https://example.com/pdf" {
		t.Errorf("Expected SUCCESS:https://example.com/xml,https://example.com/pdf, got %s", text)
	}
}

func TestExtractPACResponseTextError(t *testing.T) {
	_, err := ExtractPACResponseText("<soap:Envelope><soap:Body><GenerateInvoiceResponse><GenerateInvoiceResult>ERROR:Something went wrong</GenerateInvoiceResult></GenerateInvoiceResponse></soap:Body></soap:Envelope>")
	if err == nil {
		t.Error("Expected error for non-SUCCESS response")
	}
}

func TestExtractPACResponseTextNoTags(t *testing.T) {
	text, err := ExtractPACResponseText("SUCCESS:https://example.com/xml,https://example.com/pdf")
	if err != nil {
		t.Fatalf("ExtractPACResponseText failed: %s", err)
	}
	if text != "SUCCESS:https://example.com/xml,https://example.com/pdf" {
		t.Errorf("Expected SUCCESS:https://example.com/xml,https://example.com/pdf, got %s", text)
	}
}

func TestComprobanteImpuestosBuilder(t *testing.T) {
	conceptos := []Concepto{
		{
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{
						{Base: "1000.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "160.00"},
					},
				},
			},
		},
		{
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{
						{Base: "500.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "80.00"},
					},
				},
			},
		},
	}

	imp := buildComprobanteImpuestos(conceptos)
	if imp == nil {
		t.Fatal("Expected impuestos, got nil")
	}
	if imp.TotalImpuestosTrasladados != "240.00" {
		t.Errorf("Expected 240.00 total trasladados, got %s", imp.TotalImpuestosTrasladados)
	}
	if len(imp.Traslados.Traslado) != 2 {
		t.Errorf("Expected 2 traslados, got %d", len(imp.Traslados.Traslado))
	}
}
