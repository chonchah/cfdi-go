package cfdi

import (
	"fmt"
	"strings"
)

// Validate checks required fields and arithmetic consistency of the
// request. Arithmetic errors are reported with SAT-style Spanish messages
// so callers can surface them directly to users (the HTTP handler maps
// them to a 400 response).
func (req *CFDIRequest) Validate() (bool, []string) {
	var errs []string

	if req.Version == "" {
		errs = append(errs, "Version is required")
	}
	if req.Serie == "" {
		errs = append(errs, "Serie is required")
	}
	if req.Folio == "" {
		errs = append(errs, "Folio is required")
	}
	// Fecha may be empty: ParseCFDItoXML refreshes it automatically.
	if req.TipoDeComprobante == "" {
		errs = append(errs, "TipoDeComprobante is required")
	} else if req.TipoDeComprobante != "I" && req.TipoDeComprobante != "P" {
		errs = append(errs, "TipoDeComprobante must be I (Ingreso) or P (Pago)")
	}
	if req.Exportacion == "" {
		errs = append(errs, "Exportacion is required")
	}
	if req.LugarExpedicion == "" {
		errs = append(errs, "LugarExpedicion is required")
	}
	if req.Emisor.Rfc == "" {
		errs = append(errs, "Emisor.Rfc is required")
	}
	if req.Emisor.Nombre == "" {
		errs = append(errs, "Emisor.Nombre is required")
	}
	if req.Emisor.RegimenFiscal == "" {
		errs = append(errs, "Emisor.RegimenFiscal is required")
	}
	if req.Receptor.Rfc == "" {
		errs = append(errs, "Receptor.Rfc is required")
	}
	if req.Receptor.Nombre == "" {
		errs = append(errs, "Receptor.Nombre is required")
	}
	if req.Receptor.DomicilioFiscalReceptor == "" {
		errs = append(errs, "Receptor.DomicilioFiscalReceptor is required")
	}
	if req.Receptor.RegimenFiscalReceptor == "" {
		errs = append(errs, "Receptor.RegimenFiscalReceptor is required")
	}
	if req.Receptor.UsoCFDI == "" {
		errs = append(errs, "Receptor.UsoCFDI is required")
	}
	if len(req.Conceptos) == 0 {
		errs = append(errs, "At least one Concepto is required")
	}

	if req.TipoDeComprobante == "I" {
		if req.MetodoPago == "" {
			errs = append(errs, "MetodoPago is required for TipoDeComprobante I")
		}
		if req.FormaPago == "" {
			errs = append(errs, "FormaPago is required for TipoDeComprobante I")
		}
		if req.SubTotal == "" {
			errs = append(errs, "SubTotal is required for TipoDeComprobante I")
		}
		if req.Moneda == "" {
			errs = append(errs, "Moneda is required for TipoDeComprobante I")
		}
		if req.Moneda != "" && req.Moneda != "MXN" && req.TipoCambio == "" {
			errs = append(errs, "TipoCambio is required when Moneda is not MXN")
		}
		if req.Total == "" {
			errs = append(errs, "Total is required for TipoDeComprobante I")
		}

		errs = append(errs, req.validateAritmetica()...)
	}

	if req.TipoDeComprobante == "P" {
		if req.Complemento == nil || req.Complemento.Pagos == nil {
			errs = append(errs, "Complemento.Pagos is required for TipoDeComprobante P")
		}
		if req.Receptor.UsoCFDI != "CP01" {
			errs = append(errs, "Receptor.UsoCFDI must be CP01 for TipoDeComprobante P")
		}
	}

	return len(errs) == 0, errs
}

// validateAritmetica verifies centavo-level consistency of an Ingreso
// comprobante using exact fixed-point arithmetic (no floats):
//
//   - monetary fields carry at most two decimals;
//   - Concepto.Importe equals round(Cantidad × ValorUnitario);
//   - each Traslado/Retencion Importe equals round(Base × TasaOCuota);
//   - SubTotal equals the sum of Concepto.Importe − Concepto.Descuento;
//   - Total equals SubTotal − Descuento + trasladados − retenidos,
//     matching the totals derived by buildComprobanteImpuestos.
func (req *CFDIRequest) validateAritmetica() []string {
	var errs []string

	// amount validates a monetary attribute (≤2 decimals) and returns it
	// scaled by 10^decScale; optional empty values yield 0 without error.
	amount := func(label, val string, optional bool) int64 {
		v := strings.TrimSpace(val)
		if v == "" {
			if optional {
				return 0
			}
			errs = append(errs, fmt.Sprintf("El campo %s es requerido", label))
			return 0
		}
		if !hasMaxDec(v, maxFracDec) {
			errs = append(errs, fmt.Sprintf("El campo %s tiene más de %d decimales: \"%s\"", label, maxFracDec, val))
			return 0
		}
		parsed, ok := parseScaled(v)
		if !ok {
			errs = append(errs, fmt.Sprintf("El campo %s no es un importe válido: \"%s\"", label, val))
			return 0
		}
		return parsed
	}
	// conceptField validates a concepto-level field (Cantidad, ValorUnitario)
	// allowing up to 6 decimals per SAT rules.
	conceptField := func(label, val string, optional bool) int64 {
		v := strings.TrimSpace(val)
		if v == "" {
			if optional {
				return 0
			}
			errs = append(errs, fmt.Sprintf("El campo %s es requerido", label))
			return 0
		}
		if !hasMaxDec(v, decScale) {
			errs = append(errs, fmt.Sprintf("El campo %s tiene más de %d decimales: \"%s\"", label, decScale, val))
			return 0
		}
		parsed, ok := parseScaled(v)
		if !ok {
			errs = append(errs, fmt.Sprintf("El campo %s no es un valor válido: \"%s\"", label, val))
			return 0
		}
		return parsed
	}
	rate := func(label, val string) (int64, bool) {
		parsed, ok := parseScaled(val)
		if !ok {
			errs = append(errs, fmt.Sprintf("El campo %s no es un valor válido: \"%s\"", label, val))
			return 0, false
		}
		return parsed, true
	}
	fail := func(format string, args ...any) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	var sumSub, sumTraslados, sumRetenciones int64 // all scaled

	for i, c := range req.Conceptos {
		n := i + 1
		importe := amount(fmt.Sprintf("Importe del Concepto %d", n), c.Importe, false)
		desc := amount(fmt.Sprintf("Descuento del Concepto %d", n), c.Descuento, true)
		sumSub += importe - desc

		if cant, ok := rate(fmt.Sprintf("Cantidad del Concepto %d", n), c.Cantidad); ok {
			vu := conceptField(fmt.Sprintf("ValorUnitario del Concepto %d", n), c.ValorUnitario, false)
			if expected := mulRoundCents(cant, vu); expected != roundToCents(importe) {
				fail("El campo Importe del Concepto %d (%s) no es igual al redondeo de la Cantidad por el ValorUnitario (%s)",
					n, formatCents(roundToCents(importe)), formatCents(expected))
			}
		}

		if c.Impuestos != nil {
			if c.Impuestos.Traslados != nil {
				for j, t := range c.Impuestos.Traslados.Traslado {
					sumTraslados += validateImpuestoItem(&errs, amount, rate, fail,
						fmt.Sprintf("del Traslado %d del Concepto %d", j+1, n), t.Base, t.TasaOCuota, t.Importe, t.TipoFactor)
				}
			}
			if c.Impuestos.Retenciones != nil {
				for j, r := range c.Impuestos.Retenciones.Retencion {
					sumRetenciones += validateImpuestoItem(&errs, amount, rate, fail,
						fmt.Sprintf("de la Retención %d del Concepto %d", j+1, n), r.Base, r.TasaOCuota, r.Importe, r.TipoFactor)
				}
			}
		}
	}

	subTotal := amount("SubTotal", req.SubTotal, false)
	if subTotal != sumSub {
		fail("El campo SubTotal (%s) no es igual a la suma de los importes menos los descuentos de los conceptos (%s)",
			formatCents(roundToCents(subTotal)), formatCents(roundToCents(sumSub)))
	}

	total := amount("Total", req.Total, false)
	descGlobal := amount("Descuento", req.Descuento, true)
	expectedTotal := subTotal - descGlobal + sumTraslados - sumRetenciones
	if total != expectedTotal {
		fail("El campo Total (%s) no es igual a SubTotal menos Descuento más los impuestos trasladados menos los retenidos (%s)",
			formatCents(roundToCents(total)), formatCents(roundToCents(expectedTotal)))
	}

	return errs
}

// validateImpuestoItem checks one concepto-level traslado/retención and
// returns its Importe (scaled) for aggregation into the comprobante total.
func validateImpuestoItem(
	errs *[]string,
	amount func(label, val string, optional bool) int64,
	rate func(label, val string) (int64, bool),
	fail func(format string, args ...any),
	label, baseStr, tasaStr, importeStr, tipoFactor string,
) int64 {
	base := amount("El campo Base "+label, baseStr, false)
	importe := amount("El campo Importe "+label, importeStr, false)
	if tipoFactor != "Tasa" && tipoFactor != "Cuota" {
		return importe // Exento y otros factores no llevan cálculo de importe
	}
	if tasa, ok := rate("El campo TasaOCuota "+label, tasaStr); ok {
		if expected := mulRoundCents(tasa, base); expected != roundToCents(importe) {
			fail("El campo Importe %s (%s) no es igual al redondeo de Base por TasaOCuota (%s)",
				label, formatCents(roundToCents(importe)), formatCents(expected))
		}
	}
	return importe
}
