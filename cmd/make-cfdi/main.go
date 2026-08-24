// Command make-cfdi is an example HTTP service built on top of the cfdi
// library. It exposes a single POST /make-cfdi endpoint that generates a
// CFDI 4.0 XML document and optionally stamps it via a PAC provider.
//
// Environment variables:
//
//	PORT                   - HTTP port (default 8080)
//	PAC_URL                - PAC endpoint (omit to only generate XML)
//	PAC_USERNAME           - SOAP username
//	PAC_PASSWORD           - SOAP password
//	PAC_INVOICE_PASSWORD   - certificate password sent as the body "password" param
//	PAC_SOAP_ACTION        - SOAPAction override (default: derived from PAC_URL)
//
// The JSON body may include an optional "documentType" field that is sent as
// the SOAP documentType parameter (default "1" when absent).
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chonchah/cfdi-go"
)

func getPACConfigFromEnv() *cfdi.PACConfig {
	url := os.Getenv("PAC_URL")
	username := os.Getenv("PAC_USERNAME")
	password := os.Getenv("PAC_PASSWORD")
	invoicePassword := os.Getenv("PAC_INVOICE_PASSWORD")

	if url == "" {
		return nil
	}

	cfg := cfdi.DefaultConfig()
	cfg.URL = url
	cfg.Username = username
	cfg.Password = password
	cfg.InvoicePassword = invoicePassword

	if action := os.Getenv("PAC_SOAP_ACTION"); action != "" {
		cfg.SOAPAction = action
	}

	return cfg
}

func makeCFDI(w http.ResponseWriter, r *http.Request) {
	var req cfdi.CFDIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	valid, errs := req.Validate()
	if !valid {
		errJSON, _ := json.Marshal(map[string]any{
			"error":  "validation failed",
			"fields": errs,
		})
		http.Error(w, string(errJSON), http.StatusBadRequest)
		return
	}

	xmlContent := req.ParseCFDItoXML()

	if req.Timbrar {
		pacCfg := getPACConfigFromEnv()
		if pacCfg == nil {
			errJSON, _ := json.Marshal(map[string]string{
				"error": "timbrar=true but PAC is not configured (set PAC_URL, PAC_USERNAME, PAC_PASSWORD, PAC_INVOICE_PASSWORD)",
			})
			http.Error(w, string(errJSON), http.StatusBadRequest)
			return
		}

		resp, err := cfdi.StampCFDI(pacCfg, xmlContent, req.DocumentType, req.FileType, req.Commercio)
		if err != nil {
			errJSON, _ := json.Marshal(map[string]string{
				"error": fmt.Sprintf("PAC stamping failed: %s", err.Error()),
			})
			http.Error(w, string(errJSON), http.StatusInternalServerError)
			return
		}

		if !resp.Success {
			errJSON, _ := json.Marshal(map[string]string{
				"error":  "PAC returned error",
				"detail": resp.Error,
			})
			http.Error(w, string(errJSON), http.StatusInternalServerError)
			return
		}

		resp.XML = xmlContent

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := cfdi.PACResponse{
		Success: true,
		XML:     xmlContent,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/make-cfdi", makeCFDI)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}