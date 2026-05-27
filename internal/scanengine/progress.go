package scanengine

// Progress scan ilerlemesini SSE/UI'a iletmek için kullanılan event.
type Progress struct {
	Stage   Stage
	Done    int
	Total   int
	Message string
}
