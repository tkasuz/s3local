package s3error

import (
	"encoding/xml"
	"testing"
)

func TestError_MarshalXML(t *testing.T) {
	e := Error{
		Code:      "NoSuchKey",
		Message:   "The resource you requested does not exist",
		Resource:  "/mybucket/myfoto.jpg",
		RequestId: "4442587FB7D0A2F9",
	}

	xmlBytes, xmlErr := xml.MarshalIndent(e, "", "  ")
	if xmlErr != nil {
		t.Fatalf("failed to marshal XML: %v", xmlErr)
	}

	expected := `<Error>
  <Code>NoSuchKey</Code>
  <Message>The resource you requested does not exist</Message>
  <Resource>/mybucket/myfoto.jpg</Resource>
  <RequestId>4442587FB7D0A2F9</RequestId>
</Error>`

	if string(xmlBytes) != expected {
		t.Errorf("XML mismatch.\nGot:\n%s\n\nExpected:\n%s", string(xmlBytes), expected)
	}
}

func TestError_OmitEmpty(t *testing.T) {
	e := Error{
		Code:    "InternalError",
		Message: "Something went wrong",
	}

	xmlBytes, xmlErr := xml.MarshalIndent(e, "", "  ")
	if xmlErr != nil {
		t.Fatalf("failed to marshal XML: %v", xmlErr)
	}

	expected := `<Error>
  <Code>InternalError</Code>
  <Message>Something went wrong</Message>
</Error>`

	if string(xmlBytes) != expected {
		t.Errorf("XML mismatch.\nGot:\n%s\n\nExpected:\n%s", string(xmlBytes), expected)
	}
}
