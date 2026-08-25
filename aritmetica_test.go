package cfdi

import (
	"strings"
	"testing"
	"time"
)

// z4Conceptos mirrors the payload that stamped successfully against the
// cfdi360 sandbox: importes are the ROUND_HALF_UP of cantidad × valorUnitario
// and base × tasaOCuota, so float summation would drift to 266.21 while the
// exact centavo sum is 266.20.
func z4Request() *CFDIRequest {
	traslado := func(base, importe string) *ConceptoImpuestos {
		return &ConceptoImpuestos{
			Traslados: &TrasladosWrapper{
				Traslado: []Traslado{{
					Base: base, Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: importe,
				}},
			},
		}
	}
	return &CFDIRequest{
		Version:           "4.0",
		Serie:             "D",
		Folio:             "951514-ST",
		Fecha:             time.Now().Format("2006-01-02T15:04:05"),
		SubTotal:          "1663.79",
		Moneda:            "MXN",
		Total:             "1929.99",
		TipoDeComprobante: "I",
		Exportacion:       "01",
		MetodoPago:        "PPD",
		FormaPago:         "99",
		LugarExpedicion:   "64000",
		Emisor:            Emisor{Rfc: "JES900109Q90", Nombre: "JIMENEZ ESTRADA SALAS A A", RegimenFiscal: "601"},
		Receptor:          Receptor{Rfc: "SIM111027F90", Nombre: "SIM CARD SA DE CV", DomicilioFiscalReceptor: "64000", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos: []Concepto{
			{
				ClaveProdServ: "81111500", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "Servicio A",
				ValorUnitario: "603.45", Importe: "603.45", ObjetoImp: "02",
				Impuestos: traslado("603.45", "96.55"),
			},
			{
				ClaveProdServ: "81111500", Cantidad: "1", ClaveUnidad: "E48", Descripcion: "Servicio B",
				ValorUnitario: "1060.34", Importe: "1060.34", ObjetoImp: "02",
				Impuestos: traslado("1060.34", "169.65"),
			},
		},
	}
}

func TestZ4ScenarioGoldenXML(t *testing.T) {
	req := z4Request()
	xmlOut := req.ParseCFDItoXML()

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"seal attrs always emitted", func() bool {
			return strings.Contains(xmlOut, `Sello=""`) &&
				strings.Contains(xmlOut, `NoCertificado=""`) &&
				strings.Contains(xmlOut, `Certificado=""`)
		}},
		{"grouped global traslado", func() bool { return strings.Contains(xmlOut, `Base="1663.79"`) }},
		{"global importe is exact cent sum 266.20", func() bool {
			return strings.Contains(xmlOut, `TotalImpuestosTrasladados="266.20"`)
		}},
		{"total 1929.99 preserved", func() bool { return strings.Contains(xmlOut, `Total="1929.99"`) }},
	}
	for _, c := range checks {
		if !c.fn() {
			t.Errorf("FAIL: %s\n%s", c.name, xmlOut)
		}
	}
}

func TestZ4ScenarioValidates(t *testing.T) {
	valid, errs := z4Request().Validate()
	if !valid {
		t.Errorf("expected valid request, got errors: %v", errs)
	}
}

func TestValidateRejectsGlobalTotalDrift(t *testing.T) {
	// The original production bug: totals derived from unrounded sums
	// (266.2064 -> 266.21) while items are declared rounded.
	req := z4Request()
	req.Total = "1930.00"
	valid, errs := req.Validate()
	if valid {
		t.Fatal("expected validation failure for total drift")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "Total") && strings.Contains(e, "1929.99") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected total-mismatch error mentioning 1929.99, got: %v", errs)
	}
}

func TestValidateRejectsConceptoImporteMismatch(t *testing.T) {
	req := z4Request()
	req.Conceptos[0].Importe = "603.44"
	valid, errs := req.Validate()
	if valid {
		t.Error("expected validation failure for importe != round(cantidad*valorUnitario)")
	}
	if len(errs) == 0 || !strings.Contains(errs[0], "Cantidad por el ValorUnitario") {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestValidateRejectsTrasladoImporteMismatch(t *testing.T) {
	req := z4Request()
	req.Conceptos[1].Impuestos.Traslados.Traslado[0].Importe = "169.66"
	valid, errs := req.Validate()
	if valid {
		t.Error("expected validation failure for traslado importe != round(base*tasa)")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "Base por TasaOCuota") {
			found = true
		}
	}
	if !found {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestValidateRejectsSubTotalMismatch(t *testing.T) {
	req := z4Request()
	req.SubTotal = "1663.80"
	valid, errs := req.Validate()
	if valid {
		t.Error("expected validation failure for subtotal mismatch")
	}
	if len(errs) == 0 || !strings.Contains(errs[0], "SubTotal") {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestValidateRejectsMoreThanTwoDecimals(t *testing.T) {
	// Monto fields (Importe, Descuento, SubTotal, Total, impuestos) still max 2 decimals
	req := z4Request()
	req.Conceptos[0].Importe = "603.455"
	valid, errs := req.Validate()
	if valid {
		t.Error("expected validation failure for Importe with 3 decimals")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "decimales") {
			found = true
		}
	}
	if !found {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestValidateAllowsSixDecimalValorUnitario(t *testing.T) {
	// 1 * 100.123456 = 100.123456 -> round to 100.12
	// IVA 16%: 100.12 * 0.16 = 16.0192 -> round to 16.02
	// SubTotal = 100.12, Total = 100.12 + 16.02 = 116.14
	req := &CFDIRequest{
		Version: "4.0", Serie: "A", Folio: "1",
		Fecha: time.Now().Format("2006-01-02T15:04:05"),
		SubTotal: "100.12", Moneda: "MXN", Total: "116.14",
		TipoDeComprobante: "I", Exportacion: "01", MetodoPago: "PUE", FormaPago: "01",
		LugarExpedicion: "64000",
		Emisor:          Emisor{Rfc: "XAXX010101000", Nombre: "E", RegimenFiscal: "601"},
		Receptor:        Receptor{Rfc: "XEXX010101000", Nombre: "C", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos: []Concepto{{
			Cantidad: "1", ClaveProdServ: "84111506", ClaveUnidad: "E48", Descripcion: "x",
			ValorUnitario: "100.123456", Importe: "100.12", ObjetoImp: "02",
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{{Base: "100.12", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "16.02"}},
				},
			},
		}},
	}
	valid, errs := req.Validate()
	if !valid {
		t.Errorf("expected valid request with 6-decimal ValorUnitario, got: %v", errs)
	}
}

func TestValidateRejectsSevenDecimalValorUnitario(t *testing.T) {
	req := z4Request()
	req.Conceptos[0].ValorUnitario = "603.4555555"
	valid, errs := req.Validate()
	if valid {
		t.Error("expected validation failure for ValorUnitario with 7 decimals")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "decimales") {
			found = true
		}
	}
	if !found {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestValidateAllowsSixDecimalQuantity(t *testing.T) {
	req := &CFDIRequest{
		Version: "4.0", Serie: "A", Folio: "1",
		Fecha: time.Now().Format("2006-01-02T15:04:05"),
		SubTotal: "100.00", Moneda: "MXN", Total: "116.00",
		TipoDeComprobante: "I", Exportacion: "01", MetodoPago: "PUE", FormaPago: "01",
		LugarExpedicion: "64000",
		Emisor:          Emisor{Rfc: "XAXX010101000", Nombre: "E", RegimenFiscal: "601"},
		Receptor:        Receptor{Rfc: "XEXX010101000", Nombre: "C", DomicilioFiscalReceptor: "06500", RegimenFiscalReceptor: "601", UsoCFDI: "G03"},
		Conceptos: []Concepto{{
			Cantidad: "0.000001", ClaveProdServ: "84111506", ClaveUnidad: "E48", Descripcion: "x",
			ValorUnitario: "100000000.00", Importe: "100.00", ObjetoImp: "02",
			Impuestos: &ConceptoImpuestos{
				Traslados: &TrasladosWrapper{
					Traslado: []Traslado{{Base: "100.00", Impuesto: "002", TipoFactor: "Tasa", TasaOCuota: "0.160000", Importe: "16.00"}},
				},
			},
		}},
	}
	valid, errs := req.Validate()
	if !valid {
		t.Errorf("expected valid request with 6-decimal Cantidad, got: %v", errs)
	}
}

func TestDescuentoZeroOmitted(t *testing.T) {
	req := z4Request()
	req.Descuento = "0.00"
	xmlOut := req.ParseCFDItoXML()
	if strings.Contains(xmlOut, "Descuento") {
		t.Errorf("expected zero Descuento omitted, got:\n%s", xmlOut)
	}

	req.Descuento = ""
	xmlOut = req.ParseCFDItoXML()
	if strings.Contains(xmlOut, "Descuento") {
		t.Errorf("expected empty Descuento omitted, got:\n%s", xmlOut)
	}

	req.Descuento = "50.00"
	xmlOut = req.ParseCFDItoXML()
	if !strings.Contains(xmlOut, `Descuento="50.00"`) {
		t.Errorf("expected non-zero Descuento kept, got:\n%s", xmlOut)
	}
}

func TestFechaRefresh(t *testing.T) {
	now := time.Now()

	stale := now.Add(-96 * time.Hour).Format("2006-01-02T15:04:05")
	req := z4Request()
	req.Fecha = stale
	if got := ensureFreshFecha(req.Fecha); got == stale {
		t.Errorf("expected stale fecha (%s) to be refreshed", stale)
	}
	if out := req.ParseCFDItoXML(); strings.Contains(out, stale) {
		t.Errorf("stale fecha leaked into XML")
	}

	fresh := now.Add(-time.Hour).Format("2006-01-02T15:04:05")
	if got := ensureFreshFecha(fresh); got != fresh {
		t.Errorf("expected fresh fecha kept, got %s want %s", got, fresh)
	}

	if got := ensureFreshFecha(""); !strings.HasPrefix(got, now.Format("2006-01-02")) {
		t.Errorf("expected empty fecha filled with today, got %s", got)
	}

	req.Fecha = ""
	out := req.ParseCFDItoXML()
	if !strings.Contains(out, `Fecha="`+now.Format("2006-01-02")) {
		t.Errorf("expected auto-filled Fecha in XML, got:\n%s", out)
	}
}

func TestParseScaled(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"0", 0, true},
		{"1", 1000000, true},
		{"123.45", 123450000, true},
		{"-0.03", -30000, true},
		{"+7.5", 7500000, true},
		{"0.160000", 160000, true},
		{"", 0, false},
		{"abc", 0, false},
		{"1.2.3", 0, false},
		{"1.2345678", 0, false}, // too precise for internal scale
		{"--5", 0, false},
	}
	for _, c := range cases {
		got, ok := parseScaled(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("parseScaled(%q) = %d,%v; want %d,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestRoundToCentsHalfUp(t *testing.T) {
	cases := []struct {
		scaled int64 // scaled by 10^6
		want   int64 // cents
	}{
		{96552000, 9655},  // 96.552 -> 96.55
		{96555000, 9656},  // 96.555 -> 96.56 (half up)
		{169654400, 16965}, // 169.6544 -> 169.65
		{-96552000, -9655},
		{-96555000, -9656},
		{500000, 50},     // 0.50 -> 0.50
		{500, 0},         // 0.0005 -> 0.00
		{1500, 0},        // 0.0015 -> 0.00
		{250000, 25},     // 0.25 -> 0.25
	}
	for _, c := range cases {
		if got := roundToCents(c.scaled); got != c.want {
			t.Errorf("roundToCents(%d) = %d, want %d", c.scaled, got, c.want)
		}
	}
}

func TestMulRoundCents(t *testing.T) {
	cases := []struct {
		a, b, want int64 // scaled inputs, cents output
	}{
		// 603.45 x 0.16 = 96.552 -> 9655
		{603450000, 160000, 9655},
		// 1060.34 x 0.16 = 169.6544 -> 16965
		{1060340000, 160000, 16965},
		// 1 x 603.4483 = 603.4483 -> 60345
		{1000000, 603448300, 60345},
		// 3 x 33.333333... style: 0.333333 x 3 = 0.999999 -> 100.00? no: 1.00
		{333333, 3000000, 100},
	}
	for _, c := range cases {
		if got := mulRoundCents(c.a, c.b); got != c.want {
			t.Errorf("mulRoundCents(%d,%d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestFormatCents(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0.00"},
		{5, "0.05"},
		{-3, "-0.03"},
		{26620, "266.20"},
		{192999, "1929.99"},
		{123456789, "1234567.89"},
	}
	for _, c := range cases {
		if got := formatCents(c.in); got != c.want {
			t.Errorf("formatCents(%d) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestIsZeroAmount(t *testing.T) {
	for _, s := range []string{"0", "0.00", "-0.000"} {
		if !isZeroAmount(s) {
			t.Errorf("isZeroAmount(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "0.01", "x"} {
		if isZeroAmount(s) {
			t.Errorf("isZeroAmount(%q) = true, want false", s)
		}
	}
}

func TestHasMaxDec(t *testing.T) {
	if !hasMaxDec("12.34", 2) || hasMaxDec("12.345", 2) || !hasMaxDec("12", 2) || !hasMaxDec(".5", 1) {
		t.Error("hasMaxDec misbehaving")
	}
}
