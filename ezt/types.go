package ezt

type IFile struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
	Size int64  `json:"size"`
}

type GetAllResult struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type File struct {
	Hash  string `json:"hash"`
	IFile IFile  `json:"ifile"`
}

type PostParams struct {
	Files []File `json:"files"`
	Addr  string `json:"addr"`
}

func (l IFile) Equals(r IFile) bool {
	return (l.Name == r.Name && l.Size == r.Size && l.Dir == r.Dir)
}
