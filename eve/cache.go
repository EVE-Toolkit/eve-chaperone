package eve

type Cache struct {
	Characters   map[int64]string
	Corporations map[int64]string
	Alliances    map[int64]string
	Ships        map[int64]string
	Killmails    map[int64]FrontendKillmail
}

func NewCache() Cache {
	return Cache{
		Characters:   make(map[int64]string),
		Corporations: make(map[int64]string),
		Alliances:    make(map[int64]string),
		Ships:        make(map[int64]string),
		Killmails:    make(map[int64]FrontendKillmail),
	}
}
