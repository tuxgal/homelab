package main

type showConfigCmdHandler struct{}

func newShowConfigCmdHandler() *showConfigCmdHandler {
	return &showConfigCmdHandler{}
}

func (s *showConfigCmdHandler) containerAndGroupFlags() bool {
	return false
}

func (s *showConfigCmdHandler) run(options *cmdOptions) error {
	config := HomelabConfig{}
	err := parseHomelabConfig(&config)
	if err != nil {
		return err
	}

	log.Infof("Homelab config:\n%s", prettyPrintJSON(config))
	return nil
}
