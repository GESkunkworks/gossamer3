package gossamer3

import (
	"fmt"
	"strconv"

	"github.com/beevik/etree"
)

const (
	assertionTag           = "Assertion"
	attributeStatementTag  = "AttributeStatement"
	attributeTag           = "Attribute"
	attributeValueTag      = "AttributeValue"
	subjectConfirmationTag = "SubjectConfirmationData"
)

//ErrMissingElement is the error type that indicates an element and/or attribute is
//missing. It provides a structured error that can be more appropriately acted
//upon.
type ErrMissingElement struct {
	Tag, Attribute string
}

//ErrMissingAssertion indicates that an appropriate assertion element could not
//be found in the SAML Response
var (
	ErrMissingAssertion = ErrMissingElement{Tag: assertionTag}
)

func (e ErrMissingElement) Error() string {
	if e.Attribute != "" {
		return fmt.Sprintf("missing %s attribute on %s element", e.Attribute, e.Tag)
	}
	return fmt.Sprintf("missing %s element", e.Tag)
}

// ExtractSessionDuration this will attempt to extract a session duration from the assertion
// see https://aws.amazon.com/SAML/Attributes/SessionDuration
func ExtractSessionDuration(data []byte) (int64, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return 0, err
	}

	assertionElement := doc.FindElement(".//Assertion")
	if assertionElement == nil {
		return 0, ErrMissingAssertion
	}

	// log.Printf("tag: %s", assertionElement.Tag)

	//Get the actual assertion attributes
	attributeStatement := assertionElement.FindElement(childPath(assertionElement.Space, attributeStatementTag))
	if attributeStatement == nil {
		return 0, ErrMissingElement{Tag: attributeStatementTag}
	}

	attributes := attributeStatement.FindElements(childPath(assertionElement.Space, attributeTag))

	for _, attribute := range attributes {
		if attribute.SelectAttrValue("Name", "") != "https://aws.amazon.com/SAML/Attributes/SessionDuration" {
			continue
		}
		atributeValues := attribute.FindElements(childPath(assertionElement.Space, attributeValueTag))
		for _, attrValue := range atributeValues {
			return strconv.ParseInt(attrValue.Text(), 10, 64)
		}
	}

	return 0, nil
}

// ExtractDestinationURL will find the Destination URL to POST the SAML assertion to.
// This is necessary to support AWS instances with custom endpoints such as GovCloud and AWS China without requiring
// hardcoded endpoints on the gossamer3 side.
func ExtractDestinationURL(data []byte) (string, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return "", err
	}

	dataElement := doc.FindElement(".//" + subjectConfirmationTag)
	if dataElement == nil {
		return "", ErrMissingElement{Tag: subjectConfirmationTag}
	}

	destination := dataElement.SelectAttrValue("Recipient", "none")
	if destination == "none" {
		return "", ErrMissingElement{Tag: subjectConfirmationTag}
	}

	return destination, nil
}

// ExtractAwsRoles given an assertion document extract the aws roles
func ExtractAwsRoles(data []byte) ([]string, error) {

	awsroles := []string{}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return awsroles, err
	}

	// log.Printf("root tag: %s", doc.Root().Tag)

	assertionElement := doc.FindElement(".//Assertion")
	if assertionElement == nil {
		return nil, ErrMissingAssertion
	}

	// log.Printf("tag: %s", assertionElement.Tag)

	//Get the actual assertion attributes
	attributeStatement := assertionElement.FindElement(childPath(assertionElement.Space, attributeStatementTag))
	if attributeStatement == nil {
		return nil, ErrMissingElement{Tag: attributeStatementTag}
	}

	// log.Printf("tag: %s", attributeStatement.Tag)

	attributes := attributeStatement.FindElements(childPath(assertionElement.Space, attributeTag))
	for _, attribute := range attributes {
		if attribute.SelectAttrValue("Name", "") != "https://aws.amazon.com/SAML/Attributes/Role" {
			continue
		}
		attributeValues := attribute.FindElements(childPath(assertionElement.Space, attributeValueTag))
		for _, attrValue := range attributeValues {
			awsroles = append(awsroles, attrValue.Text())
		}
	}

	return awsroles, nil
}

// ExtractRoleSessionName given an assertion document extract the role session name
func ExtractRoleSessionName(data []byte) (string, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return "", err
	}

	assertionElement := doc.FindElement(".//Assertion")
	if assertionElement == nil {
		return "", ErrMissingAssertion
	}

	// log.Printf("tag: %s", assertionElement.Tag)

	//Get the actual assertion attributes
	attributeStatement := assertionElement.FindElement(childPath(assertionElement.Space, attributeStatementTag))
	if attributeStatement == nil {
		return "", ErrMissingElement{Tag: attributeStatementTag}
	}

	attributes := attributeStatement.FindElements(childPath(assertionElement.Space, attributeTag))

	for _, attribute := range attributes {
		if attribute.SelectAttrValue("Name", "") != "https://aws.amazon.com/SAML/Attributes/RoleSessionName" {
			continue
		}
		attributeValues := attribute.FindElements(childPath(assertionElement.Space, attributeValueTag))
		for _, attrValue := range attributeValues {
			return attrValue.Text(), nil
		}
	}

	return "", nil
}

func childPath(space, tag string) string {
	if space == "" {
		return "./" + tag
	}
	//log.Printf("query = %s", "./"+space+":"+tag)
	return "./" + space + ":" + tag
}
