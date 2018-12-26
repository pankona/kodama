package kodama

type Worker interface {
	Work(description string) error
}
