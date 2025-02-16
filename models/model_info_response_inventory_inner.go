package models

type InfoResponseInventoryInner struct {

	// Тип предмета.
	Type string `json:"type,omitempty"`

	// Количество предметов.
	Quantity int32 `json:"quantity,omitempty"`
}

// AssertInfoResponseInventoryInnerRequired checks if the required fields are not zero-ed
func AssertInfoResponseInventoryInnerRequired(obj InfoResponseInventoryInner) error {
	return nil
}

// AssertInfoResponseInventoryInnerConstraints checks if the values respects the defined constraints
func AssertInfoResponseInventoryInnerConstraints(obj InfoResponseInventoryInner) error {
	return nil
}
