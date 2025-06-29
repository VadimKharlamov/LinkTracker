package bot

type UserState struct {
	Step    string
	Link    string
	Tags    []string
	Filters []string
}

type UserData struct {
	TrackedLinks map[string][]string
}
