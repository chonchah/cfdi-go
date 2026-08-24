package cfdi

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"
)

// CFDIRequest is the structured input used to generate a CFDI 4.0.
type CFDIRequest struct {
	Timbrar           bool         `json:"timbrar"`
	Version           string       `json:"Version"`
	Serie             string       `json:"Serie"`
	Folio             string       `json:"Folio"`
	Fecha             string       `json:"Fecha"`
	CondicionesDePago string       `json:"CondicionesDePago,omitempty"`
	SubTotal          string       `json:"SubTotal,omitempty"`
	Descuento         string       `json:"Descuento,omitempty"`
	Moneda            string       `json:"Moneda,omitempty"`
	TipoCambio        string       `json:"TipoCambio,omitempty"`
	Total             string       `json:"Total,omitempty"`
	TipoDeComprobante string       `json:"TipoDeComprobante"`
	Exportacion       string       `json:"Exportacion"`
	MetodoPago        string       `json:"MetodoPago,omitempty"`
	FormaPago         string       `json:"FormaPago,omitempty"`
	LugarExpedicion   string       `json:"LugarExpedicion"`
	Confirmacion      string       `json:"Confirmacion,omitempty"`
	Emisor            Emisor       `json:"Emisor"`
	Receptor          Receptor     `json:"Receptor"`
	Conceptos         []Concepto   `json:"Conceptos"`
	Complemento       *Complemento `json:"Complemento,omitempty"`
	// DocumentType is the SOAP documentType sent to the PAC. Empty means the
	// default "1" (Ingreso) is used.
	DocumentType string `json:"documentType,omitempty"`
	// FileType is the SOAP fileType sent to the PAC: "0" = TXT, "1" = XML.
	// Empty means the default "1" (XML) is used. The service always builds
	// XML, so this should normally stay "1".
	FileType string `json:"fileType,omitempty"`
	// Commercio is the SOAP commercio flag sent to the PAC. Empty means "0".
	Commercio string `json:"commercio,omitempty"`
}

// FromJSON decodes a CFDIRequest from its JSON representation.
func (req *CFDIRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, req)
}

// ParseCFDItoXML builds the CFDI 4.0 XML string from the request.
//
// Notes verified against the cfdi360 sandbox:
//   - Sello/NoCertificado/Certificado are always emitted (empty), see cfdi.go.
//   - A stale or missing Fecha is refreshed: the SAT rejects comprobantes
//     older than 72 hours.
//   - An explicit zero Descuento is omitted entirely.
//   - Concept-level amounts are emitted exactly as provided; callers must
//     send tax Importes already rounded to 2 decimals (Validate enforces
//     this). Global tax totals are computed as the exact sum of the given
//     values using integer arithmetic.
func (req *CFDIRequest) ParseCFDItoXML() string {
	cfdiXML := &CFDI{
		Version:           req.Version,
		Serie:             req.Serie,
		Folio:             req.Folio,
		Fecha:             ensureFreshFecha(req.Fecha),
		CondicionesDePago: req.CondicionesDePago,
		SubTotal:          req.SubTotal,
		Moneda:            req.Moneda,
		TipoCambio:        req.TipoCambio,
		Total:             req.Total,
		TipoDeComprobante: req.TipoDeComprobante,
		Exportacion:       req.Exportacion,
		MetodoPago:        req.MetodoPago,
		FormaPago:         req.FormaPago,
		LugarExpedicion:   req.LugarExpedicion,
		Confirmacion:      req.Confirmacion,
		XmlnsCFDI:         "http://www.sat.gob.mx/cfd/4",
		XmlnsXsi:          "http://www.w3.org/2001/XMLSchema-instance",
		Emisor:            req.Emisor,
		Receptor:          req.Receptor,
	}

	if d := strings.TrimSpace(req.Descuento); d != "" && !isZeroAmount(d) {
		cfdiXML.Descuento = d
	}

	if req.TipoDeComprobante == "P" {
		p20ns := "http://www.sat.gob.mx/Pagos20"
		cfdiXML.XmlnsPago20 = &p20ns
		cfdiXML.XsiSchemaLocation = "http://www.sat.gob.mx/cfd/4 http://www.sat.gob.mx/sitio_internet/cfd/4/cfdv40.xsd http://www.sat.gob.mx/Pagos20 http://www.sat.gob.mx/sitio_internet/cfd/Pagos/Pagos20.xsd"
	} else {
		cfdiXML.XsiSchemaLocation = "http://www.sat.gob.mx/cfd/4 http://www.sat.gob.mx/sitio_internet/cfd/4/cfdv40.xsd"
	}

	cfdiXML.Conceptos.Concepto = req.Conceptos

	if req.TipoDeComprobante == "I" {
		impuestos := buildComprobanteImpuestos(req.Conceptos)
		if impuestos != nil {
			cfdiXML.Impuestos = impuestos
		}
	}

	if req.TipoDeComprobante == "P" && req.Complemento != nil {
		cfdiXML.Complemento = req.Complemento
	}

	xmlBytes, err := xml.MarshalIndent(cfdiXML, "", "  ")
	if err != nil {
		return err.Error()
	}
	return xml.Header + string(xmlBytes)
}

// ensureFreshFecha returns fecha when it is a valid CFDI timestamp within
// the SAT 72-hour window (allowing one hour of clock skew into the future);
// otherwise it returns the current local time in CFDI format.
func ensureFreshFecha(fecha string) string {
	now := time.Now()
	if fecha != "" {
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", fecha, now.Location()); err == nil {
			d := now.Sub(t)
			if d >= -time.Hour && d < 72*time.Hour {
				return fecha
			}
		}
	}
	return now.Format("2006-01-02T15:04:05")
}

// buildComprobanteImpuestos derives the Comprobante level Impuestos node
// from the concept taxes. Global Traslado/Retencion nodes are grouped by
// (Impuesto, TipoFactor, TasaOCuota) as the SAT requires; their Base and
// Importe are the exact integer-cent sums of the grouped concept values,
// rendered with two decimals.
func buildComprobanteImpuestos(conceptos []Concepto) *Impuestos {
	type agg struct {
		baseCents, importeCents int64
	}
	traslados := map[string]*agg{}
	retenciones := map[string]*agg{}
	var trasladoOrder, retencionOrder []string

	for _, c := range conceptos {
		if c.Impuestos == nil {
			continue
		}
		if c.Impuestos.Traslados != nil {
			for _, t := range c.Impuestos.Traslados.Traslado {
				k := t.Impuesto + "|" + t.TipoFactor + "|" + t.TasaOCuota
				a := traslados[k]
				if a == nil {
					a = &agg{}
					traslados[k] = a
					trasladoOrder = append(trasladoOrder, k)
				}
				a.baseCents += amountCents(t.Base)
				a.importeCents += amountCents(t.Importe)
			}
		}
		if c.Impuestos.Retenciones != nil {
			for _, r := range c.Impuestos.Retenciones.Retencion {
				k := r.Impuesto + "|" + r.TipoFactor + "|" + r.TasaOCuota
				a := retenciones[k]
				if a == nil {
					a = &agg{}
					retenciones[k] = a
					retencionOrder = append(retencionOrder, k)
				}
				a.baseCents += amountCents(r.Base)
				a.importeCents += amountCents(r.Importe)
			}
		}
	}

	if len(trasladoOrder) == 0 && len(retencionOrder) == 0 {
		return nil
	}

	imp := &Impuestos{}
	if len(trasladoOrder) > 0 {
		var total int64
		grouped := make([]Traslado, 0, len(trasladoOrder))
		for _, k := range trasladoOrder {
			a := traslados[k]
			parts := strings.SplitN(k, "|", 3)
			grouped = append(grouped, Traslado{
				Base:       formatCents(a.baseCents),
				Impuesto:   parts[0],
				TipoFactor: parts[1],
				TasaOCuota: parts[2],
				Importe:    formatCents(a.importeCents),
			})
			total += a.importeCents
		}
		imp.TotalImpuestosTrasladados = formatCents(total)
		imp.Traslados = &TrasladosWrapper{Traslado: grouped}
	}
	if len(retencionOrder) > 0 {
		var total int64
		grouped := make([]Retencion, 0, len(retencionOrder))
		for _, k := range retencionOrder {
			a := retenciones[k]
			parts := strings.SplitN(k, "|", 3)
			grouped = append(grouped, Retencion{
				Base:       formatCents(a.baseCents),
				Impuesto:   parts[0],
				TipoFactor: parts[1],
				TasaOCuota: parts[2],
				Importe:    formatCents(a.importeCents),
			})
			total += a.importeCents
		}
		imp.TotalImpuestosRetenidos = formatCents(total)
		imp.Retenciones = &RetencionesWrapper{Retencion: grouped}
	}
	return imp
}
