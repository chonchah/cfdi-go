package cfdi

import "encoding/json"

// UnmarshalJSON allows simpler JSON input where Traslados/Retenciones are
// plain arrays instead of wrapped objects.
func (ci *ConceptoImpuestos) UnmarshalJSON(data []byte) error {
	type alias struct {
		Traslados   []Traslado  `json:"Traslados,omitempty"`
		Retenciones []Retencion `json:"Retenciones,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	if a.Traslados != nil {
		ci.Traslados = &TrasladosWrapper{Traslado: a.Traslados}
	}
	if a.Retenciones != nil {
		ci.Retenciones = &RetencionesWrapper{Retencion: a.Retenciones}
	}
	return nil
}

// UnmarshalJSON allows simpler JSON input for the invoice-level taxes.
func (i *Impuestos) UnmarshalJSON(data []byte) error {
	type alias struct {
		TotalImpuestosTrasladados string      `json:"TotalImpuestosTrasladados,omitempty"`
		TotalImpuestosRetenidos   string      `json:"TotalImpuestosRetenidos,omitempty"`
		Traslados                 []Traslado  `json:"Traslados,omitempty"`
		Retenciones               []Retencion `json:"Retenciones,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	i.TotalImpuestosTrasladados = a.TotalImpuestosTrasladados
	i.TotalImpuestosRetenidos = a.TotalImpuestosRetenidos
	if a.Traslados != nil {
		i.Traslados = &TrasladosWrapper{Traslado: a.Traslados}
	}
	if a.Retenciones != nil {
		i.Retenciones = &RetencionesWrapper{Retencion: a.Retenciones}
	}
	return nil
}

// UnmarshalJSON allows simpler JSON input for related-document taxes.
func (didr *DoctoImpuestosDR) UnmarshalJSON(data []byte) error {
	type alias struct {
		TrasladosDR []TrasladoDR `json:"TrasladosDR,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	if a.TrasladosDR != nil {
		didr.TrasladosDR = &TrasladosDRWrapper{TrasladoDR: a.TrasladosDR}
	}
	return nil
}