package main

import "sort"

func containerMapToList(cm containerMap) containerList {
	res := make(containerList, 0, len(cm))
	for _, c := range cm {
		res = append(res, c)
	}

	// Return containers sorted by order. Group order takes higher priority
	// over container order within the same group. If two containers still
	// have the same order at both the group and container levels, then
	// the container name is used to lexicographically sort the containers.
	sort.Slice(res, func(i, j int) bool {
		c1 := res[i]
		c2 := res[j]
		if c1.group.config.Order == c2.group.config.Order {
			if c1.config.Order == c2.config.Order {
				return c1.Name() < c2.Name()
			}
			return c1.config.Order < c2.config.Order
		} else {
			return c1.group.config.Order < c2.group.config.Order
		}
	})
	return res
}

func queryContainers(dep *deployment, options *cmdOptions) containerList {
	if options.containerAndGroup.allGroups {
		return containerMapToList(dep.queryAllContainers())
	} else if options.containerAndGroup.group != "" && options.containerAndGroup.container == "" {
		return containerMapToList(dep.queryAllContainersInGroup(options.containerAndGroup.group))
	} else if options.containerAndGroup.group != "" {
		c := dep.queryContainer(options.containerAndGroup.group, options.containerAndGroup.container)
		if c != nil {
			return containerList{c}
		}
		return nil
	}
	log.Fatalf("Invalid scenario, possibly indicating a bug in the code")
	return nil
}
