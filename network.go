package main

import "fmt"

type network struct {
	deployment          *deployment
	mode                networkMode
	bridgeModeConfig    *BridgeModeNetworkConfig
	containerModeConfig *ContainerModeNetworkConfig
}

type networkMap map[string]*network

const (
	networkModeUnknown networkMode = iota
	networkModeBridge
	networkModeContainer
)

type networkMode uint8

func newBridgeModeNetwork(dep *deployment, config *BridgeModeNetworkConfig) *network {
	n := network{
		deployment:       dep,
		mode:             networkModeBridge,
		bridgeModeConfig: config,
	}
	return &n
}

func newContainerModeNetwork(dep *deployment, config *ContainerModeNetworkConfig) *network {
	n := network{
		deployment:          dep,
		mode:                networkModeContainer,
		containerModeConfig: config,
	}
	return &n
}

func (n *network) name() string {
	if n.mode == networkModeBridge {
		return n.bridgeModeConfig.Name
	} else if n.mode == networkModeContainer {
		return n.containerModeConfig.Name
	} else {
		log.Fatalf("unknown network, possibly indicating a bug in the code!")
		return "{Network Unknown}"
	}
}

func (n *network) String() string {
	if n.mode == networkModeBridge {
		return fmt.Sprintf("{Network (Bridge) Name: %s}", n.name())
	} else if n.mode == networkModeContainer {
		return fmt.Sprintf("{Network (Container) Name: %s}", n.name())
	} else {
		log.Fatalf("unknown network, possibly indicating a bug in the code!")
		return "{Network Unknown}"
	}
}

func (n networkMap) String() string {
	return stringifyMap(n)
}
