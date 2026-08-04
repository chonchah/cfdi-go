package cfdi

// Validate checks required fields in the request.
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
	if req.Fecha == "" {
		errs = append(errs, "Fecha is required")
	}
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
		if req.Total == "" {
			errs = append(errs, "Total is required for TipoDeComprobante I")
		}
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