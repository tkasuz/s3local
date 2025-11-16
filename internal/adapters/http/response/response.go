package response

import (
	"encoding/xml"
	"net/http"
)

// DomainError interface that both bucket and object domain errors implement
type DomainError interface {
	error
	GetCode() string
	GetMessage() string
	GetStatusCode() int
}

// S3Error represents an S3 XML error response
type S3Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource"`
	RequestId string   `xml:"RequestId"`
}

// WriteXML writes an XML response with the given status code
func WriteXML(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	xml.NewEncoder(w).Encode(data)
}

// WriteError writes an S3-compatible error response in XML format
func WriteError(w http.ResponseWriter, code, message, resource string, status int) {
	errorResp := S3Error{
		Code:      code,
		Message:   message,
		Resource:  resource,
		RequestId: "local-request",
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	xml.NewEncoder(w).Encode(errorResp)
}

// WriteDomainError writes a domain error as an S3 XML error response
// This works with any domain error that has Code, Message, and StatusCode
func WriteDomainError(w http.ResponseWriter, code, message string, statusCode int, resource string) {
	errorResp := S3Error{
		Code:      code,
		Message:   message,
		Resource:  resource,
		RequestId: "local-request",
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)
	xml.NewEncoder(w).Encode(errorResp)
}
