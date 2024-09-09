package main

func queryContainers(dep *deployment, options *cmdOptions) containerMap {
	if options.containerAndGroup.allGroups {
		return dep.queryAllContainers()
	} else if options.containerAndGroup.group != "" && options.containerAndGroup.container == "" {
		return dep.queryAllContainersInGroup(options.containerAndGroup.group)
	} else if options.containerAndGroup.group != "" {
		c := dep.queryContainer(options.containerAndGroup.group, options.containerAndGroup.container)
		if c != nil {
			return containerMap{c.config.Name: c}
		}
		return nil
	}
	log.Fatalf("Invalid scenario, possibly indicating a bug in the code")
	return nil
}
