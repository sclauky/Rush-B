package post

type Game struct {
	ID   int
	Name string
}

type CreatePostData struct {
	Title   string
	Content string
	GameID  int
	Mode    string
	Games   []Game
	Error   string
	Success string
}
