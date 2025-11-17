package teams

type Member struct {
	Id       string
	Name     string
	IsActive bool
}

type Team struct {
	Name    string
	Members []Member
}
