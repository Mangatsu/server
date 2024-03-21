package utils

type OrderedMap struct {
	Keys []string
	Map  map[string]interface{}
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys: make([]string, 0),
		Map:  make(map[string]interface{}),
	}
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if _, exists := om.Map[key]; !exists {
		om.Keys = append(om.Keys, key)
	}
	om.Map[key] = value
}

func (om *OrderedMap) Get(key string) (interface{}, bool) {
	val, exists := om.Map[key]
	return val, exists
}

func (om *OrderedMap) OrderedKeys() []string {
	return om.Keys
}
