package types

import "fmt"

func (cd *CollectionData) CanExpand(other *CollectionData) bool {
	return cd.ApiInterface == other.ApiInterface && cd.Type == other.Type && cd.InternalPath == other.InternalPath || other.ApiInterface == ""
}

// expand is called within the same spec apiCollections, to manage inheritance within collections of different add_ons
func (apic *ApiCollection) Expand(myCollections map[CollectionData]*ApiCollection, dependencies map[CollectionData]struct{}) error {
	dependencies[apic.CollectionData] = struct{}{}
	defer delete(dependencies, apic.CollectionData)
	inheritanceApis := apic.InheritanceApis
	apic.InheritanceApis = []*CollectionData{} // delete inheritance so if someone calls expand on this in the future without dependency we don't repeat this
	relevantCollections := []*ApiCollection{}
	for _, inheritingCollection := range inheritanceApis {
		if collection, ok := myCollections[*inheritingCollection]; ok {
			if !inheritingCollection.CanExpand(&collection.CollectionData) {
				return fmt.Errorf("invalid inheriting collection %v", inheritingCollection)
			}
			if _, ok := dependencies[collection.CollectionData]; ok {
				return fmt.Errorf("circular dependency in inheritance, %v", collection)
			}
			err := collection.Expand(myCollections, dependencies)
			if err != nil {
				return err
			}
			relevantCollections = append(relevantCollections, collection)
		} else {
			return fmt.Errorf("did not find inheritingCollection in myCollections %v", inheritingCollection)
		}
	}
	return apic.ApisMerge(relevantCollections, dependencies)
}

// inherit is
func (apic *ApiCollection) Inherit(relevantCollections []*ApiCollection, dependencies map[CollectionData]struct{}) error {
	// do not set dependencies because this mechanism protects inheritance within the same spec and inherit is inheritance between different specs so same type is allowed
	return apic.ApisMerge(relevantCollections, dependencies)
}

func (apic *ApiCollection) ApisMerge(relevantCollections []*ApiCollection, dependencies map[CollectionData]struct{}) error {
	currentApis := make(map[string]struct{})
	for _, api := range apic.Apis {
		currentApis[api.Name] = struct{}{}
	}
	var mergedApis []*Api
	mergedApisMap := make(map[string]*Api)
	// TODO: also Expand headers
	for _, collection := range relevantCollections {
		for _, api := range collection.Apis {
			if api.Enabled {
				// duplicate API(s) not allowed
				// (unless current Spec has an override for same API)
				if _, found := mergedApisMap[api.Name]; found {
					if _, found := currentApis[api.Name]; !found {
						return fmt.Errorf("duplicate imported api: %s (in collection: %v)", api.Name, collection.CollectionData)
					}
				}
				mergedApisMap[api.Name] = api
				mergedApis = append(mergedApis, api)
			}
		}
	}
	// merge collected APIs into current spec's APIs (unless overridden)
	for _, api := range mergedApis {
		if _, found := currentApis[api.Name]; !found {
			apic.Apis = append(apic.Apis, api)
		}
	}
	return nil
}

func (apic *ApiCollection) Equals(other *ApiCollection) bool {
	return other.CollectionData == apic.CollectionData
}

// all apiCollections are already expanded
func (apic *ApiCollection) InheritApis(myCollections map[CollectionData]*ApiCollection, relevantParentCollections []*ApiCollection) error {
	for _, other := range relevantParentCollections {
		if !apic.Equals(other) {
			return fmt.Errorf("incompatible inheritance, apiCollections aren't equal %v", apic)
		}
	}
	err := apic.Expand(myCollections, map[CollectionData]struct{}{})
	if err != nil {
		return err
	}
	return apic.Inherit(relevantParentCollections, map[CollectionData]struct{}{})
}

// this does not allow repetitions
func (apic *ApiCollection) CombineWithOthers(others []*ApiCollection) (*ApiCollection, error) {
	mergedApis := make(map[string]struct{})
	for _, api := range apic.Apis {
		mergedApis[api.Name] = struct{}{}
	}

	for _, collection := range others {
		for _, api := range collection.Apis {
			if _, ok := mergedApis[api.Name]; !ok {
				mergedApis[api.Name] = struct{}{}
				apic.Apis = append(apic.Apis, api)
			} else {
				return nil, fmt.Errorf("existing api in collection combination %s %v", api.Name, apic)
			}
		}
	}
	return apic, nil
}