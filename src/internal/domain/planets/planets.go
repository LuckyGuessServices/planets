package planets

var mercury = &PlanetInfo{
	index: 1,
	name:  "Mercury",
}

func Mercury() *PlanetInfo {
	return mercury
}

type PlanetInfo struct {
	index int16
	name  string
}

func (info *PlanetInfo) Index() int16 {
	return info.index
}

func (info *PlanetInfo) Name() string {
	return info.name
}
