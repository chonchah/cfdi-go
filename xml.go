package cfdi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// CFDIRequest is the structured input used to generate a CFDI 4.0.
type CFDIRequest struct {
	Timbrar           bool         `json:"timbrar"`
	Version           string       `json:"Version"`
	Serie             string       `json:"Serie"`
	Folio             string       `json:"Folio"`
	Fecha             string       `json:"Fecha"`
	SubTotal          string       `json:"SubTotal,omitempty"`
	Moneda            string       `json:"Moneda,omitempty"`
	Total             string       `json:"Total,omitempty"`
	TipoDeComprobante string       `json:"TipoDeComprobante"`
	Exportacion       string       `json:"Exportacion"`
	MetodoPago        string       `json:"MetodoPago,omitempty"`
	FormaPago         string       `json:"FormaPago,omitempty"`
	LugarExpedicion   string       `json:"LugarExpedicion"`
	Emisor            Emisor       `json:"Emisor"`
	Receptor          Receptor     `json:"Receptor"`
	Conceptos         []Concepto   `json:"Conceptos"`
	Complemento       *Complemento `json:"Complemento,omitempty"`
}

// FromJSON decodes a CFDIRequest from its JSON representation.
func (req *CFDIRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, req)
}

// ParseCFDItoXML builds the CFDI 4.0 XML string from the request.
func (req *CFDIRequest) ParseCFDItoXML() string {
	cfdiXML := &CFDI{
		Version:           req.Version,
		Serie:             req.Serie,
		Folio:             req.Folio,
		Fecha:             req.Fecha,
		SubTotal:          req.SubTotal,
		Moneda:            req.Moneda,
		Total:             req.Total,
		TipoDeComprobante: req.TipoDeComprobante,
		Exportacion:       req.Exportacion,
		MetodoPago:        req.MetodoPago,
		FormaPago:         req.FormaPago,
		LugarExpedicion:   req.LugarExpedicion,
		XmlnsCFDI:         "http://www.sat.gob.mx/cfd/4",
		XmlnsXsi:          "http://www.w3.org/2001/XMLSchema-instance",
		Emisor:            req.Emisor,
		Receptor:          req.Receptor,
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

func buildComprobanteImpuestos(conceptos []Concepto) *Impuestos {
	var totalTrasladados float64
	var totalRetenidos float64
	var allTraslados []Traslado
	var allRetenciones []Retencion
	hasTraslados := false
	hasRetenciones := false

	for _, c := range conceptos {
		if c.Impuestos != nil {
			if c.Impuestos.Traslados != nil {
				for _, t := range c.Impuestos.Traslados.Traslado {
					allTraslados = append(allTraslados, t)
					totalTrasladados += parseFloat(t.Importe)
					hasTraslados = true
				}
			}
			if c.Impuestos.Retenciones != nil {
				for _, r := range c.Impuestos.Retenciones.Retencion {
					allRetenciones = append(allRetenciones, r)
					totalRetenidos += parseFloat(r.Importe)
					hasRetenciones = true
				}
			}
		}
	}

	if !hasTraslados && !hasRetenciones {
		return nil
	}

	imp := &Impuestos{}
	if hasTraslados {
		imp.TotalImpuestosTrasladados = formatFloat(totalTrasladados)
		imp.Traslados = &TrasladosWrapper{Traslado: allTraslados}
	}
	if hasRetenciones {
		imp.TotalImpuestosRetenidos = formatFloat(totalRetenidos)
		imp.Retenciones = &RetencionesWrapper{Retencion: allRetenciones}
	}
	return imp
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}