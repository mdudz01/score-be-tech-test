package domain

type Case struct {
	CaseID  string   `json:"case_id"`
	Clients []Client `json:"clients"`
}

type Client struct {
	ClientID string  `json:"client_id"`
	Details  Details `json:"details"`
}

type Details struct {
	FirstName        string  `json:"first_name"`
	LastName         string  `json:"last_name"`
	Address          Address `json:"address"`
	DOB              string  `json:"dob"`
	Dependents       int     `json:"dependents"`
	EmploymentStatus string  `json:"employment_status"`
}

type Address struct {
	Address1 string `json:"address1"`
	PostTown string `json:"posttown"`
	Postcode string `json:"postcode"`
}
