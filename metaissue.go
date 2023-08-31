package jira

import (
	"fmt"
	"strings"

	"github.com/google/go-querystring/query"
	"github.com/trivago/tgo/tcontainer"
)

// MetaProject is the meta information about a project returned from createmeta api
type MetaProject struct {
	Self string `json:"self,omitempty"`
	Id   string `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	// omitted avatarUrls
	//IssueTypes []*MetaIssueType `json:"issuetypes,omitempty"`
	IssueTypes []*IssueType `json:"issuetypes,omitempty"`
}

// MetaIssueType represents the different issue types a project has.
//
// Note: Fields is interface because this is an object which can
// have arbitraty keys related to customfields. It is not possible to
// expect these for a general way. This will be returning a map.
// Further processing must be done depending on what is required.
type MetaIssueType struct {
	Self        string                `json:"self,omitempty"`
	Id          string                `json:"id,omitempty"`
	Description string                `json:"description,omitempty"`
	IconUrl     string                `json:"iconurl,omitempty"`
	Name        string                `json:"name,omitempty"`
	Subtasks    bool                  `json:"subtask,omitempty"`
	Expand      string                `json:"expand,omitempty"`
	Fields      tcontainer.MarshalMap `json:"fields,omitempty"`
}

type issueTypeMetaResp struct {
	Values []tcontainer.MarshalMap `json:"values,omitempty"`
}

// GetCreateMeta makes the api call to get the meta information required to create a ticket
func (s *IssueService) GetCreateMeta(projectKey string) (*MetaProject, *Response, error) {
	return s.GetCreateMetaWithOptions(projectKey, &GetQueryOptions{})
}

// GetCreateMetaWithOptions makes the api call to get the meta information without requiring to have a projectKey
func (s *IssueService) GetCreateMetaWithOptions(projectKey string, options *GetQueryOptions) (*MetaProject, *Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/project/%s", projectKey)

	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	if options != nil {
		q, err := query.Values(options)
		if err != nil {
			return nil, nil, err
		}
		req.URL.RawQuery = q.Encode()
	}

	fmt.Printf("%v\n", req.URL.String())

	meta := new(MetaProject)
	resp, err := s.client.Do(req, meta)

	if err != nil {
		return nil, resp, err
	}

	return meta, resp, nil
}

// GetCreateMeta makes the api call to get the meta information required to create a ticket
func (s *IssueService) GetIssueTypeMeta(projectKey string, issueType *IssueType) (*MetaIssueType, *Response, error) {
	return s.GetIssueTypeMetaWithOptions(projectKey, issueType, &GetQueryOptions{})
}

// GetCreateMetaWithOptions makes the api call to get the meta information without requiring to have a projectKey
func (s *IssueService) GetIssueTypeMetaWithOptions(projectKey string, issueType *IssueType, options *GetQueryOptions) (*MetaIssueType, *Response, error) {
	apiEndpoint := fmt.Sprintf("rest/api/2/issue/createmeta/%s/issuetypes/%s", projectKey, issueType.ID)

	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	if options != nil {
		q, err := query.Values(options)
		if err != nil {
			return nil, nil, err
		}
		req.URL.RawQuery = q.Encode()
	}

	fmt.Printf("%v\n", req.URL.String())

	meta := &MetaIssueType{
		Self:        issueType.Self,
		Id:          issueType.ID,
		Description: issueType.Description,
		IconUrl:     issueType.IconURL,
		Name:        issueType.Name,
		Subtasks:    issueType.Subtask,
	}

	content := new(issueTypeMetaResp)
	resp, err := s.client.Do(req, content)
	if err != nil {
		return nil, resp, err
	}

	meta.Fields = make(tcontainer.MarshalMap)
	for _, field := range content.Values {
		meta.Fields[field["fieldId"].(string)] = field
	}

	return meta, resp, nil
}

// GetIssueTypeWithName returns an ID of an IssueType with name from a given MetaProject.
// The comparison of the name is case insensitive
func (p *MetaProject) GetIssueTypeWithName(name string) (*IssueType, error) {
	for _, m := range p.IssueTypes {
		if strings.EqualFold(m.Name, strings.ToLower(name)) {
			return m, nil
		}
	}
	return nil, fmt.Errorf("IssueType with name %s not found", name)
}

// GetMandatoryFields returns a map of all the required fields from the MetaIssueTypes.
// if a field returned by the api was:
//
//	"customfield_10806": {
//						"required": true,
//						"schema": {
//							"type": "any",
//							"custom": "com.pyxis.greenhopper.jira:gh-epic-link",
//							"customId": 10806
//						},
//						"name": "Epic Link",
//						"hasDefaultValue": false,
//						"operations": [
//							"set"
//						]
//					}
//
// the returned map would have "Epic Link" as the key and "customfield_10806" as value.
// This choice has been made so that the it is easier to generate the create api request later.
func (t *MetaIssueType) GetMandatoryFields() (map[string]string, error) {
	ret := make(map[string]string)
	for key := range t.Fields {
		required, err := t.Fields.Bool(key + "/required")
		if err != nil {
			return nil, err
		}
		if required {
			name, err := t.Fields.String(key + "/name")
			if err != nil {
				return nil, err
			}
			ret[name] = key
		}
	}
	return ret, nil
}

// GetAllFields returns a map of all the fields for an IssueType. This includes all required and not required.
// The key of the returned map is what you see in the form and the value is how it is representated in the jira schema.
func (t *MetaIssueType) GetAllFields() (map[string]string, error) {
	ret := make(map[string]string)
	for key := range t.Fields {

		name, err := t.Fields.String(key + "/name")
		if err != nil {
			return nil, err
		}
		ret[name] = key
	}
	return ret, nil
}

// CheckCompleteAndAvailable checks if the given fields satisfies the mandatory field required to create a issue for the given type
// And also if the given fields are available.
func (t *MetaIssueType) CheckCompleteAndAvailable(config map[string]string) (bool, error) {
	mandatory, err := t.GetMandatoryFields()
	if err != nil {
		return false, err
	}
	all, err := t.GetAllFields()
	if err != nil {
		return false, err
	}

	// check templateconfig against mandatory fields
	for key := range mandatory {
		if _, okay := config[key]; !okay {
			var requiredFields []string
			for name := range mandatory {
				requiredFields = append(requiredFields, name)
			}
			return false, fmt.Errorf("required field not found in provided jira.fields. Required are: %#v", requiredFields)
		}
	}

	// check templateConfig against all fields to verify they are available
	for key := range config {
		if _, okay := all[key]; !okay {
			var availableFields []string
			for name := range all {
				availableFields = append(availableFields, name)
			}
			return false, fmt.Errorf("fields in jira.fields are not available in jira. Available are: %#v", availableFields)
		}
	}

	return true, nil
}
