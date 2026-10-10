package cfdi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateIngresoXML(t *testing.T) {
	req := &CFDIRequest{
		Version:           "4.0",
		Serie:             "A",
		Folio:             "100",
		Fecha:             "2026-07-01T10:00:00",
		CondicionesDePago: "Pago a 30 dias",
		SubTotal:          "1000.00",
		Descuento:         "100.00",
		Moneda:            "MXN",
		Total:             "1160.00",
		TipoDeComprobante: "I",
		Exportacion:       "01",
		MetodoPago:        "PPD",
		FormaPago:         "99",
		LugarExpedicion:   "06000",
		Confirmacion:      "ABC123",
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
				ClaveProdServ:    "84111506",
				NoIdentificacion: "PZA-001",
				Cantidad:         "1.00",
				ClaveUnidad:      "E48",
				Unidad:           "Pieza",
				Descripcion:      "Servicios de ejemplo",
				ValorUnitario:    "1000.00",
				Importe:          "1000.00",
				Descuento:        "50.00",
				ObjetoImp:        "02",
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
		{"Descuento=\"100.00\"", func() bool { return strings.Contains(xmlOut, `Descuento="100.00"`) }},
		{"CondicionesDePago", func() bool { return strings.Contains(xmlOut, `CondicionesDePago="Pago a 30 dias"`) }},
		{"Confirmacion=\"ABC123\"", func() bool { return strings.Contains(xmlOut, `Confirmacion="ABC123"`) }},
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
		{"NoIdentificacion", func() bool { return strings.Contains(xmlOut, `NoIdentificacion="PZA-001"`) }},
		{"Unidad", func() bool { return strings.Contains(xmlOut, `Unidad="Pieza"`) }},
		{"concept-level Descuento", func() bool { return strings.Contains(xmlOut, `Descuento="50.00"`) }},
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
								EquivalenciaDR:  "1",
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
		{"TipoCambioP defaults to 1 for MXN", func() bool { return strings.Contains(xmlOut, `TipoCambioP="1"`) }},
		{"FechaPago", func() bool { return strings.Contains(xmlOut, `FechaPago="2026-07-15T14:25:00"`) }},
		{"DoctoRelacionado element", func() bool { return strings.Contains(xmlOut, "pago20:DoctoRelacionado") }},
		{"IdDocumento", func() bool { return strings.Contains(xmlOut, `IdDocumento="UUID-DE-LA-FACTURA-ORIGINAL-PPD"`) }},
		{"EquivalenciaDR", func() bool { return strings.Contains(xmlOut, `EquivalenciaDR="1"`) }},
		{"no legacy EquivMonedaDR", func() bool { return !strings.Contains(xmlOut, `EquivMonedaDR`) }},
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

func TestGenerateIngresoXMLTipoCambio(t *testing.T) {
	req := &CFDIRequest{
		Version:           "4.0",
		Serie:             "A",
		Folio:             "100",
		Fecha:             "2026-07-01T10:00:00",
		SubTotal:          "1000.00",
		Moneda:            "USD",
		TipoCambio:        "17.50",
		Total:             "1160.00",
		TipoDeComprobante: "I",
		Exportacion:       "01",
		MetodoPago:        "PPD",
		FormaPago:         "99",
		LugarExpedicion:   "06000",
		Emisor:            Emisor{Rfc: "XAXX010101000", Nombre: "E", RegimenFiscal: "601"},
		Receptor:          Receptor{Rfc: "XEXX010101000", Nombre: "C", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos:         []Concepto{{ClaveProdServ: "84111506", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "S", ValorUnitario: "1000", Importe: "1000", ObjetoImp: "02"}},
	}

	xmlOut := req.ParseCFDItoXML()
	if !strings.Contains(xmlOut, `Moneda="USD"`) {
		t.Error("expected Moneda=\"USD\"")
	}
	if !strings.Contains(xmlOut, `TipoCambio="17.50"`) {
		t.Error("expected TipoCambio=\"17.50\"")
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
		Conceptos: []Concepto{{
			ClaveProdServ: "84111506", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "S",
			ValorUnitario: "1000.00", Importe: "1000.00", ObjetoImp: "02",
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{{Base: "1000.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "160.00"}},
				},
			},
		}},
	}

	valid, errs := req.Validate()
	if !valid {
		t.Errorf("Expected valid I-type request, got errors: %v", errs)
	}
}

func TestValidateTipoCambio(t *testing.T) {
	req := &CFDIRequest{
		TipoDeComprobante: "I",
		Version:           "4.0",
		Serie:             "A",
		Folio:             "100",
		Fecha:             "2026-07-01T10:00:00",
		SubTotal:          "1000.00",
		Moneda:            "USD",
		Total:             "1160.00",
		Exportacion:       "01",
		MetodoPago:        "PPD",
		FormaPago:         "99",
		LugarExpedicion:   "06000",
		Emisor:            Emisor{Rfc: "XAXX010101000", Nombre: "E", RegimenFiscal: "601"},
		Receptor:          Receptor{Rfc: "XEXX010101000", Nombre: "C", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos: []Concepto{{
			ClaveProdServ: "84111506", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "S",
			ValorUnitario: "1000.00", Importe: "1000.00", ObjetoImp: "02",
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{{Base: "1000.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "160.00"}},
				},
			},
		}},
	}

	valid, errs := req.Validate()
	if valid {
		t.Error("Expected validation to fail when Moneda is not MXN and TipoCambio is missing")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "TipoCambio") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about TipoCambio, got: %v", errs)
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
	const (
		url             = "https://cfdi.pac-provider.com/webservice"
		user            = "login-user"
		pass            = "login-pass"
		invoicePassword = "cert-pass"
	)
	doc := `<cfdi:Comprobante Serie="FA"><cfdi:Emisor/></cfdi:Comprobante>`
	payload := BuildSOAPPayload(url, doc, "1", "1", "0", user, pass, invoicePassword)

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"SOAP-ENV envelope", func() bool { return strings.Contains(payload, "SOAP-ENV:Envelope") }},
		{"SOAP-ENV header", func() bool { return strings.Contains(payload, "SOAP-ENV:Header") }},
		{"SOAP-ENV body", func() bool { return strings.Contains(payload, "SOAP-ENV:Body") }},
		{"operation generateInvoiceVersion4", func() bool { return strings.Contains(payload, "ns1:generateInvoiceVersion4") }},
		{"no legacy GenerateInvoice", func() bool { return !strings.Contains(payload, "GenerateInvoice") }},
		{"webservice URL as operation namespace", func() bool { return strings.Contains(payload, `xmlns:ns1="`+url+`"`) }},
		{"Auth header without extra namespace", func() bool { return strings.Contains(payload, "<ns1:Auth>") }},
		{"userName plain in Auth header", func() bool { return strings.Contains(payload, "<userName>"+user+"</userName>") }},
		{"password plain in Auth header", func() bool { return strings.Contains(payload, "<password>"+pass+"</password>") }},
		{"document escaped", func() bool { return strings.Contains(payload, `&lt;cfdi:Comprobante Serie="FA"&gt;`) }},
		{"quotes not escaped inside text content", func() bool { return strings.Contains(payload, `Serie="FA"`) }},
		{"documentType param", func() bool { return strings.Contains(payload, "<documentType>1</documentType>") }},
		{"fileType=1", func() bool { return strings.Contains(payload, "<fileType>1</fileType>") }},
		{"cer empty self-closing", func() bool { return strings.Contains(payload, "<cer/>") }},
		{"key empty self-closing", func() bool { return strings.Contains(payload, "<key/>") }},
		{"certificate password param", func() bool { return strings.Contains(payload, "<password>"+invoicePassword+"</password>") }},
		{"commercio=0", func() bool { return strings.Contains(payload, "<commercio>0</commercio>") }},
		{"params ordered: cer before key", func() bool { return strings.Index(payload, "<cer/>") < strings.Index(payload, "<key/>") }},
		{"params ordered: key before password", func() bool { return strings.Index(payload, "<key/>") < strings.Index(payload, "<password>"+invoicePassword) }},
		{"params ordered: password before commercio", func() bool { return strings.Index(payload, "<commercio>") > strings.Index(payload, "<password>"+invoicePassword) }},
	}

	for _, check := range checks {
		if !check.fn() {
			t.Errorf("FAIL: %s", check.name)
		}
	}
}

func TestBuildSOAPPayloadGolden(t *testing.T) {
	const (
		url             = "https://cfdi.pac-provider.com/webservice"
		user            = "user"
		pass            = "pass"
		invoicePassword = "certpass"
	)
	doc := `<cfdi:Comprobante><cfdi:Emisor/></cfdi:Comprobante>`

	got := BuildSOAPPayload(url, doc, "1", "1", "0", user, pass, invoicePassword)

	want := `<?xml version="1.0" encoding="utf-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="https://cfdi.pac-provider.com/webservice">
  <SOAP-ENV:Header>
    <ns1:Auth>
      <userName>user</userName>
      <password>pass</password>
    </ns1:Auth>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <ns1:generateInvoiceVersion4>
      <document>&lt;cfdi:Comprobante&gt;&lt;cfdi:Emisor/&gt;&lt;/cfdi:Comprobante&gt;</document>
      <documentType>1</documentType>
      <fileType>1</fileType>
      <cer/>
      <key/>
      <password>certpass</password>
      <commercio>0</commercio>
    </ns1:generateInvoiceVersion4>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	if got != want {
		t.Errorf("SOAP envelope mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestStampCFDIHeaders(t *testing.T) {
	const url = "https://cfdi.pac-provider.com/webservice"

	var lastReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastReq = r
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write([]byte(`<soap:Envelope><soap:Body><generateInvoiceVersion4Response><generateInvoiceVersion4Result>SUCCESS:https://example.com/xml,https://example.com/pdf</generateInvoiceVersion4Result></generateInvoiceVersion4Response></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	cfg := &PACConfig{
		URL:             server.URL,
		Username:        "user",
		Password:        "pass",
		InvoicePassword: "certpass",
	}

	resp, err := StampCFDI(cfg, `<cfdi:Comprobante/>`, "", "", "")
	if err != nil {
		t.Fatalf("StampCFDI failed: %s", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}

	if lastReq == nil {
		t.Fatal("no request captured")
	}
	if ct := lastReq.Header.Get("Content-Type"); ct != "text/xml; charset=utf-8" {
		t.Errorf("expected Content-Type text/xml; charset=utf-8, got %s", ct)
	}
	wantAction := DefaultSOAPAction
	if sa := lastReq.Header.Get("SOAPAction"); sa != wantAction {
		t.Errorf("expected default SOAPAction %q, got %q", wantAction, sa)
	}
	if auth := lastReq.Header.Get("Authorization"); auth != "" {
		t.Errorf("expected no Authorization header after removing Basic Auth, got %q", auth)
	}
}

func TestStampCFDICustomSOAPAction(t *testing.T) {
	var lastReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastReq = r
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write([]byte(`<soap:Envelope><soap:Body><generateInvoiceVersion4Response><generateInvoiceVersion4Result>SUCCESS:https://example.com/xml,https://example.com/pdf</generateInvoiceVersion4Result></generateInvoiceVersion4Response></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	cfg := &PACConfig{
		URL:        server.URL,
		Username:   "user",
		Password:   "pass",
		SOAPAction: `"custom-action"`,
	}

	if _, err := StampCFDI(cfg, `<cfdi:Comprobante/>`, "", "", ""); err != nil {
		t.Fatalf("StampCFDI failed: %s", err)
	}
	if sa := lastReq.Header.Get("SOAPAction"); sa != `"custom-action"` {
		t.Errorf("expected configured SOAPAction, got %q", sa)
	}
}

func TestStampCFDIDocumentType(t *testing.T) {
	const soapResp = `<soap:Envelope><soap:Body><generateInvoiceVersion4Response><generateInvoiceVersion4Result>SUCCESS:https://example.com/xml,https://example.com/pdf</generateInvoiceVersion4Result></generateInvoiceVersion4Response></soap:Body></soap:Envelope>`

	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write([]byte(soapResp))
	}))
	defer server.Close()

	cfg := &PACConfig{URL: server.URL}

	if _, err := StampCFDI(cfg, `<cfdi:Comprobante/>`, "", "", ""); err != nil {
		t.Fatalf("StampCFDI failed: %s", err)
	}
	if !strings.Contains(string(body), "<documentType>1</documentType>") {
		t.Errorf("expected default documentType 1, got body:\n%s", body)
	}

	if _, err := StampCFDI(cfg, `<cfdi:Comprobante/>`, "2", "", ""); err != nil {
		t.Fatalf("StampCFDI failed: %s", err)
	}
	if !strings.Contains(string(body), "<documentType>2</documentType>") {
		t.Errorf("expected documentType 2, got body:\n%s", body)
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
	if len(imp.Traslados.Traslado) != 1 {
		t.Fatalf("Expected traslados grouped into 1 node, got %d", len(imp.Traslados.Traslado))
	}
	g := imp.Traslados.Traslado[0]
	if g.Base != "1500.00" {
		t.Errorf("Expected grouped Base 1500.00, got %s", g.Base)
	}
	if g.Importe != "240.00" {
		t.Errorf("Expected grouped Importe 240.00, got %s", g.Importe)
	}

	// Distinct rates must remain separate groups, in first-seen order.
	conceptos = append(conceptos, Concepto{
		Impuestos: &ConceptoImpuestos{
			Traslados: &TrasladosWrapper{
				Traslado: []Traslado{
					{Base: "1000.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.080000", Importe: "80.00"},
				},
			},
		},
	})
	imp = buildComprobanteImpuestos(conceptos)
	if len(imp.Traslados.Traslado) != 2 {
		t.Fatalf("Expected 2 groups after adding 8%% rate, got %d", len(imp.Traslados.Traslado))
	}
	if imp.Traslados.Traslado[0].TasaOCuota != "0.160000" || imp.Traslados.Traslado[1].TasaOCuota != "0.080000" {
		t.Errorf("Unexpected group order/rates: %+v", imp.Traslados.Traslado)
	}
	if imp.TotalImpuestosTrasladados != "320.00" {
		t.Errorf("Expected total trasladados 320.00, got %s", imp.TotalImpuestosTrasladados)
	}
}
