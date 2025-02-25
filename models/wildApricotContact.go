package models

import (
	"errors"
	"fmt"
	"rfid-backend/config"
	"strconv"
)

// Contact represents the structure of a contact in the Wild Apricot API's /Contacts response.
type Contact struct {
	FirstName              string       `json:"FirstName"`
	LastName               string       `json:"LastName"`
	Email                  string       `json:"Email"`
	DisplayName            string       `json:"DisplayName"`
	Organization           string       `json:"Organization"`
	ProfileLastUpdated     string       `json:"ProfileLastUpdated"`
	FieldValues            []FieldValue `json:"FieldValues"`
	Id                     int          `json:"Id"`
	Url                    string       `json:"Url"`
	IsAccountAdministrator bool         `json:"IsAccountAdministrator"`
	TermsOfUseAccepted     bool         `json:"TermsOfUseAccepted"`
	Status                 string       `json:"Status"`
}

// FieldValue represents the structure for field values in a contact.
type FieldValue struct {
	FieldName  string      `json:"FieldName"`
	Value      interface{} `json:"Value"`
	SystemCode string      `json:"SystemCode"`
}

// Training represents a training item in FieldValue containing active values from a WA multiple choice list.
type SafetyTraining struct {
	Id    int    `json:"Id"`
	Label string `json:"Label"`
}

// Returns contact_id, tagId, trainings
func (c *Contact) ExtractTagID(cfg *config.Config) (uint32, error) {
	for _, val := range c.FieldValues {
		if val.FieldName == cfg.TagIdFieldName {
			return parseTagId(val)
		}
	}
	return 0, nil // Return 0 if TagId field is not found
}

// Extracts training labels from contact field values.
func (c *Contact) ExtractTrainingLabels(cfg *config.Config) ([]string, error) {
	for _, val := range c.FieldValues {
		if val.FieldName == cfg.TrainingFieldName {
			return parseTrainingLabels(val)
		}
	}
	return nil, nil // Return nil if Training field is not found
}

// Combines extraction of Tag ID and Training Labels.
func (c *Contact) ExtractContactData(cfg *config.Config) (int, uint32, []string, error) {
	tagID, err := c.ExtractTagID(cfg)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("error extracting TagId for contact %d: %v", c.Id, err)
	}

	trainingLabels, err := c.ExtractTrainingLabels(cfg)
	if err != nil {
		err = fmt.Errorf("error extracting training labels for contact %d: %v", c.Id, err)
	}

	return c.Id, tagID, trainingLabels, err
}

func parseTagId(fieldValue FieldValue) (uint32, error) {
	// Check that the field has a value before trying to convert it to a string.
	if fieldValue.Value == nil {
		return 0, nil
	}

	strVal, ok := fieldValue.Value.(string)
	if !ok {
		return 0, errors.New("TagId value is not a string")
	}

	if len(strVal) <= 0 {
		// Suppress error on empty TagId field value, return 0
		return uint32(0), nil
	}

	tagId, err := strconv.ParseInt(strVal, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string TagId to int: %v", err)
	}

	if tagId <= 0 {
		return 0, errors.New("TagId value is non-positive")
	}

	return uint32(tagId), nil
}

func parseTrainingLabels(fieldValue FieldValue) ([]string, error) {
	trainingValues, ok := fieldValue.Value.([]interface{})
	if !ok {
		return nil, errors.New("training value is not a slice")
	}

	var labels []string
	for _, t := range trainingValues {
		trainingMap, ok := t.(map[string]interface{})
		if !ok {
			return nil, errors.New("training item is not a map")
		}

		label, err := extractLabelFromTrainingMap(trainingMap)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, nil
}

func extractLabelFromTrainingMap(trainingMap map[string]interface{}) (string, error) {
	label, ok := trainingMap["Label"].(string)
	if !ok {
		return "", errors.New("training label is not a string")
	}

	return label, nil
}
