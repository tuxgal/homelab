package main

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