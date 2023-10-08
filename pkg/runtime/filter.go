package runtime

import (
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
	"sort"
	"strings"
)

type lessTypeFunc func(d1, d2 Device) bool

type typeSorter struct {
	ds        []Device
	lessFuncs []lessTypeFunc
}

func ByDevice(less ...lessTypeFunc) *typeSorter {
	return &typeSorter{
		lessFuncs: less,
	}
}
func (ms *typeSorter) Sort(ds []Device) {
	ms.ds = ds
	sort.Sort(ms)
}

func (ms *typeSorter) Len() int {
	return len(ms.ds)
}

func (ms *typeSorter) Swap(i, j int) {
	ms.ds[i], ms.ds[j] = ms.ds[j], ms.ds[i]
}

func (ms *typeSorter) Less(i, j int) bool {
	return ms.less(ms.ds[i], ms.ds[j])
}

func (ms *typeSorter) less(p, q Device) bool {
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.lessFuncs)-1; k++ {
		less := ms.lessFuncs[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return ms.lessFuncs[k](p, q)
}

func (ms *typeSorter) Insert(ds []Device, d Device) []Device {
	i := sort.Search(len(ds), func(i int) bool { return ms.less(ds[i], d) })
	ds = append(ds, d)
	copy(ds[i+1:], ds[i:])
	ds[i] = d
	return ds
}

type NameFilterFunc struct {
	Eq         string
	In         []string
	Contains   string
	StartsWith string
	EndsWith   string
}

type DeviceFilter struct {
	Name       interface{}
	Id         string
	DeviceCode string
	DeviceType string
}

type predicateType func(d Device) bool

func ParseTypeFilter(filter *DeviceFilter) []predicateType {
	predicates := make([]predicateType, 0)

	// id
	if len(filter.Id) > 0 {
		p := func(dd Device) bool {
			if filter.Id == dd.GetID() {
				return true
			}
			return false
		}
		predicates = append(predicates, p)
	}

	// name
	if filter.Name != nil {
		if name, ok := filter.Name.(string); ok {
			p := func(d Device) bool {
				if name == d.GetName() {
					return true
				}
				return false
			}
			predicates = append(predicates, p)
		} else {
			var ff NameFilterFunc
			if err := mapstructure.Decode(filter.Name, &ff); err != nil {
				klog.V(3).InfoS("Failed to parse filter.name", "err", err)
			}
			// eq
			if len(ff.Eq) > 0 {
				p := func(d Device) bool {
					if ff.Eq == d.GetName() {
						return true
					}
					return false
				}
				predicates = append(predicates, p)
			}
			// in
			if len(ff.In) > 0 {
				p := func(d Device) bool {
					for _, name := range ff.In {
						if name == d.GetName() {
							return true
						}
					}
					return false
				}
				predicates = append(predicates, p)
			}
			// contains
			if len(ff.Contains) > 0 {
				p := func(d Device) bool {
					if strings.Contains(d.GetName(), ff.Contains) {
						return true
					}
					return false
				}
				predicates = append(predicates, p)
			}
			// startsWith
			if len(ff.StartsWith) > 0 {
				p := func(d Device) bool {
					if strings.HasPrefix(d.GetName(), strings.TrimSpace(ff.StartsWith)) {
						return true
					}
					return false
				}
				predicates = append(predicates, p)
			}
			// endsWith
			if len(ff.EndsWith) > 0 {
				p := func(d Device) bool {
					if strings.HasSuffix(d.GetName(), strings.TrimSpace(ff.EndsWith)) {
						return true
					}
					return false
				}
				predicates = append(predicates, p)
			}
		}
	}

	return predicates
}
