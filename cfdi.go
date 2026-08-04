package cfdi

import (
	"encoding/xml"
)

// CFDI represents the root Comprobante element for CFDI 4.0.
type CFDI struct {
	XMLName           xml.Name `xml:"cfdi:Comprobante"`
	XmlnsCFDI         string   `xml:"xmlns:cfdi,attr"`
	XmlnsXsi          string   `xml:"xmlns:xsi,attr"`
	XmlnsPago20       *string  `xml:"xmlns:pago20,attr,omitempty"`
	XsiSchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	Version           string   `xml:"Version,attr"`
	Serie             string   `xml:"Serie,attr"`
	Folio             string   `xml:"Folio,attr"`
	Fecha             string   `xml:"Fecha,attr"`
	SubTotal          string   `xml:"SubTotal,attr"`
	Moneda            string   `xml:"Moneda,attr"`
	Total             string   `xml:"Total,attr"`
	TipoDeComprobante string   `xml:"TipoDeComprobante,attr"`
	Exportacion       string   `xml:"Exportacion,attr"`
	MetodoPago        string   `xml:"MetodoPago,attr,omitempty"`
	FormaPago         string   `xml:"FormaPago,attr,omitempty"`
	LugarExpedicion   string   `xml:"LugarExpedicion,attr"`
	NoCertificado     string   `xml:"NoCertificado,attr,omitempty"`
	Certificado       string   `xml:"Certificado,attr,omitempty"`
	Sello             string   `xml:"Sello,attr,omitempty"`
	Emisor            Emisor   `xml:"cfdi:Emisor"`
	Receptor          Receptor `xml:"cfdi:Receptor"`
	Conceptos         Conceptos
	Impuestos         *Impuestos   `xml:"cfdi:Impuestos,omitempty"`
	Complemento       *Complemento `xml:"cfdi:Complemento,omitempty"`
}

// Conceptos groups the invoice line items.
type Conceptos struct {
	XMLName  xml.Name   `xml:"cfdi:Conceptos"`
	Concepto []Concepto `xml:"cfdi:Concepto"`
}

// Emisor is the issuer of the invoice.
type Emisor struct {
	Rfc           string `xml:"Rfc,attr" json:"Rfc"`
	Nombre        string `xml:"Nombre,attr" json:"Nombre"`
	RegimenFiscal string `xml:"RegimenFiscal,attr" json:"RegimenFiscal"`
}

// Receptor is the recipient of the invoice.
type Receptor struct {
	Rfc                     string `xml:"Rfc,attr" json:"Rfc"`
	Nombre                  string `xml:"Nombre,attr" json:"Nombre"`
	DomicilioFiscalReceptor string `xml:"DomicilioFiscalReceptor,attr" json:"DomicilioFiscalReceptor"`
	RegimenFiscalReceptor   string `xml:"RegimenFiscalReceptor,attr" json:"RegimenFiscalReceptor"`
	UsoCFDI                 string `xml:"UsoCFDI,attr" json:"UsoCFDI"`
}

// Concepto is a single line item of the invoice.
type Concepto struct {
	ClaveProdServ string             `xml:"ClaveProdServ,attr" json:"ClaveProdServ"`
	Cantidad      string             `xml:"Cantidad,attr" json:"Cantidad"`
	ClaveUnidad   string             `xml:"ClaveUnidad,attr" json:"ClaveUnidad"`
	Descripcion   string             `xml:"Descripcion,attr" json:"Descripcion"`
	ValorUnitario string             `xml:"ValorUnitario,attr" json:"ValorUnitario"`
	Importe       string             `xml:"Importe,attr" json:"Importe"`
	ObjetoImp     string             `xml:"ObjetoImp,attr" json:"ObjetoImp"`
	Impuestos     *ConceptoImpuestos `xml:"cfdi:Impuestos,omitempty" json:"Impuestos,omitempty"`
}

// ConceptoImpuestos are the taxes of a single line item.
type ConceptoImpuestos struct {
	Traslados   *TrasladosWrapper   `xml:"cfdi:Traslados,omitempty" json:"Traslados,omitempty"`
	Retenciones *RetencionesWrapper `xml:"cfdi:Retenciones,omitempty" json:"Retenciones,omitempty"`
}

// TrasladosWrapper wraps a list of traslados.
type TrasladosWrapper struct {
	Traslado []Traslado `xml:"cfdi:Traslado" json:"Traslado"`
}

// RetencionesWrapper wraps a list of retenciones.
type RetencionesWrapper struct {
	Retencion []Retencion `xml:"cfdi:Retencion" json:"Retencion"`
}

// Traslado represents an additive tax (e.g. IVA).
type Traslado struct {
	Base       string `xml:"Base,attr" json:"Base"`
	Impuesto   string `xml:"Impuesto,attr" json:"Impuesto"`
	TipoFactor string `xml:"TipoFactor,attr" json:"TipoFactor"`
	TasaOCuota string `xml:"TasaOCuota,attr" json:"TasaOCuota"`
	Importe    string `xml:"Importe,attr" json:"Importe"`
}

// Retencion represents a withholding tax.
type Retencion struct {
	Base       string `xml:"Base,attr" json:"Base"`
	Impuesto   string `xml:"Impuesto,attr" json:"Impuesto"`
	TipoFactor string `xml:"TipoFactor,attr" json:"TipoFactor"`
	TasaOCuota string `xml:"TasaOCuota,attr" json:"TasaOCuota"`
	Importe    string `xml:"Importe,attr" json:"Importe"`
}

// Impuestos are the total taxes of the invoice.
type Impuestos struct {
	TotalImpuestosTrasladados string              `xml:"TotalImpuestosTrasladados,attr,omitempty" json:"TotalImpuestosTrasladados,omitempty"`
	TotalImpuestosRetenidos   string              `xml:"TotalImpuestosRetenidos,attr,omitempty" json:"TotalImpuestosRetenidos,omitempty"`
	Traslados                 *TrasladosWrapper   `xml:"cfdi:Traslados,omitempty" json:"Traslados,omitempty"`
	Retenciones               *RetencionesWrapper `xml:"cfdi:Retenciones,omitempty" json:"Retenciones,omitempty"`
}

// Complemento holds tax complements, currently Pagos 2.0.
type Complemento struct {
	Pagos *Pagos `xml:"pago20:Pagos,omitempty" json:"Pagos,omitempty"`
}

// Pagos is the Pagos 2.0 complement.
type Pagos struct {
	Version string  `xml:"Version,attr" json:"Version"`
	Totales Totales `xml:"pago20:Totales" json:"Totales"`
	Pago    []Pago  `xml:"pago20:Pago" json:"Pago"`
}

// Totales are the totals of a Pagos 2.0 complement.
type Totales struct {
	TotalTrasladosBaseIVA16     string `xml:"TotalTrasladosBaseIVA16,attr" json:"TotalTrasladosBaseIVA16"`
	TotalTrasladosImpuestoIVA16 string `xml:"TotalTrasladosImpuestoIVA16,attr" json:"TotalTrasladosImpuestoIVA16"`
	MontoTotalPagos             string `xml:"MontoTotalPagos,attr" json:"MontoTotalPagos"`
}

// Pago is a single payment in a Pagos 2.0 complement.
type Pago struct {
	FechaPago        string             `xml:"FechaPago,attr" json:"FechaPago"`
	FormaDePagoP     string             `xml:"FormaDePagoP,attr" json:"FormaDePagoP"`
	MonedaP          string             `xml:"MonedaP,attr" json:"MonedaP"`
	Monto            string             `xml:"Monto,attr" json:"Monto"`
	DoctoRelacionado []DoctoRelacionado `xml:"pago20:DoctoRelacionado" json:"DoctoRelacionado"`
}

// DoctoRelacionado links a payment to the original invoice.
type DoctoRelacionado struct {
	IdDocumento      string            `xml:"IdDocumento,attr" json:"IdDocumento"`
	Serie            string            `xml:"Serie,attr" json:"Serie"`
	Folio            string            `xml:"Folio,attr" json:"Folio"`
	MonedaDR         string            `xml:"MonedaDR,attr" json:"MonedaDR"`
	EquivMonedaDR    string            `xml:"EquivMonedaDR,attr" json:"EquivMonedaDR"`
	NumParcialidad   string            `xml:"NumParcialidad,attr" json:"NumParcialidad"`
	ImpSaldoAnt      string            `xml:"ImpSaldoAnt,attr" json:"ImpSaldoAnt"`
	ImpPagado        string            `xml:"ImpPagado,attr" json:"ImpPagado"`
	ImpSaldoInsoluto string            `xml:"ImpSaldoInsoluto,attr" json:"ImpSaldoInsoluto"`
	ObjetoImpDR      string            `xml:"ObjetoImpDR,attr" json:"ObjetoImpDR"`
	ImpuestosDR      *DoctoImpuestosDR `xml:"pago20:ImpuestosDR,omitempty" json:"ImpuestosDR,omitempty"`
}

// DoctoImpuestosDR holds taxes of a related document.
type DoctoImpuestosDR struct {
	TrasladosDR *TrasladosDRWrapper `xml:"pago20:TrasladosDR,omitempty" json:"TrasladosDR,omitempty"`
}

// TrasladosDRWrapper wraps a list of traslados DR.
type TrasladosDRWrapper struct {
	TrasladoDR []TrasladoDR `xml:"pago20:TrasladoDR" json:"TrasladoDR"`
}

// TrasladoDR is a traslado of a related document.
type TrasladoDR struct {
	BaseDR       string `xml:"BaseDR,attr" json:"BaseDR"`
	ImpuestoDR   string `xml:"ImpuestoDR,attr" json:"ImpuestoDR"`
	TipoFactorDR string `xml:"TipoFactorDR,attr" json:"TipoFactorDR"`
	TasaOCuotaDR string `xml:"TasaOCuotaDR,attr" json:"TasaOCuotaDR"`
	ImporteDR    string `xml:"ImporteDR,attr" json:"ImporteDR"`
}