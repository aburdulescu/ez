package ezt

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

type GetResult struct {
	IFile IFile    `json:"ifile"`
	Peers []string `json:"peers"`
}
