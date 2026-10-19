package domain

// Case is one mortgage case (the top level of case.json).
type Case struct {
	CaseID  string   `json:"case_id"`
	Clients []Client `json:"clients"`
}

// Client is one person on the case.
type Client struct {
	ClientID string  `json:"client_id"`
	Details  Details `json:"details"`
}

// Details holds a client's personal information.
type Details struct {
	FirstName        string  `json:"first_name"`
	LastName         string  `json:"last_name"`
	Address          Address `json:"address"`
	DOB              string  `json:"dob"`
	Dependents       int     `json:"dependents"`
	EmploymentStatus string  `json:"employment_status"`
}

// Address is a client's address. Only Postcode is used, to find the region.
type Address struct {
	Address1 string `json:"address1"`
	PostTown string `json:"posttown"`
	Postcode string `json:"postcode"`
}
