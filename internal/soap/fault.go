package soap

import (
	"encoding/xml"
	"strings"

	"forge-api-client/internal/models"
)

type node struct {
	XMLName  xml.Name
	Text     string `xml:",chardata"`
	Children []node `xml:",any"`
}

func DetectFault(body string) *models.SOAPFault {
	var root node
	if xml.Unmarshal([]byte(body), &root) != nil {
		return nil
	}
	fault := find(&root, "Fault")
	if fault == nil {
		return nil
	}
	code := textAt(fault, "faultcode")
	if code == "" {
		if n := find(fault, "Code"); n != nil {
			code = textAt(n, "Value")
		}
	}
	message := textAt(fault, "faultstring")
	if message == "" {
		if n := find(fault, "Reason"); n != nil {
			message = textAt(n, "Text")
		}
	}
	return &models.SOAPFault{Code: strings.TrimSpace(code), Message: strings.TrimSpace(message)}
}
func find(n *node, local string) *node {
	if n.XMLName.Local == local {
		return n
	}
	for i := range n.Children {
		if found := find(&n.Children[i], local); found != nil {
			return found
		}
	}
	return nil
}
func textAt(n *node, local string) string {
	if found := find(n, local); found != nil {
		return found.Text
	}
	return ""
}
