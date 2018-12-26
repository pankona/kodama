package kodama

type Validator interface {
	Validate(desc string) error
}
